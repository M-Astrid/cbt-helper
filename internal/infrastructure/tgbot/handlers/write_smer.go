package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/M-Astrid/cbt-helper/internal/app/usecases/save_smer"
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/renderer"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/smer_steps"
	"gopkg.in/telebot.v3"
)

func WriteSMER(c telebot.Context, config *common.BotConfig) error {
	userID := c.Sender().ID
	var state smer_steps.UserState
	if s, ok := config.UserStates[userID]; ok {
		state = s
	}
	smer := entity.NewSMEREntry(userID)
	state.SMER = smer
	state.SMERSteps = []smer_steps.StepI{&smer_steps.StepSMERTrigger{}, &smer_steps.StepSMEREmotions{}, &smer_steps.StepSMERThoughts{}}
	state.CurrentSMERStepIdx = 0
	state.SMERSteps[state.CurrentSMERStepIdx].Start(c.Bot(), c.Sender(), userID, &state)
	config.UserStates[userID] = state
	return nil
}

func SaveSMER(c telebot.Context, ctx context.Context, config *common.BotConfig) error {
	userID := c.Sender().ID
	state, exists := config.UserStates[userID]
	if !exists {
		return nil
	}
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(ctx)

	i := save_smer_usecase.NewInteractor(adapter)
	if err := i.Call(ctx, state.SMER); err != nil {
		c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
	} else {
		c.Send("Успешно сохранили запись.")
	}
	return nil
}

func RenderAndSaveSmer(c telebot.Context, state smer_steps.UserState, interactor *save_smer_usecase.Interactor) {
	if _, err := interactor.Call(context.Background(), state.SMER); err != nil {
		c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
		return
	}
	rnd := renderer.NewSmerRenderer()
	msg, err := rnd.RenderMessageSingle(state.SMER)
	if err != nil {
		c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
		return
	}
	c.Send(fmt.Sprintf("%s", *msg), &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					telebot.InlineButton{Unique: "work_smer", Text: "Работать с мыслью"},
					telebot.InlineButton{Unique: "del_smer", Text: "Удалить", Data: state.SMER.ID},
				},
			},
		},
		ParseMode: telebot.ModeHTML,
	})
}
