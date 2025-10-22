package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/M-Astrid/cbt-helper/internal/app/usecases/del_smer"
	"github.com/M-Astrid/cbt-helper/internal/app/usecases/get_smer"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/renderer"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"gopkg.in/telebot.v3"
)

func DeleteSMER(c telebot.Context, config *common.BotConfig) error {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(context.Background())

	smerId := strings.Split(c.Data(), ":")[1]
	i := getSingleSMERUsecase.NewInteractor(adapter)
	s, err := i.Call(context.Background(), smerId)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
	}

	rnd := renderer.NewSmerRenderer()
	smerRepr, err := rnd.RenderShortSingle(s)
	if err != nil {
		return c.Send(fmt.Sprintf("Произошла ошибка %s", err))
	}
	message := fmt.Sprintf("Удалить эту запись без возможности восстановления?\n\n%s", *smerRepr)
	return c.Send(message, &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{telebot.InlineButton{
					Unique: "del_smer_for_sure",
					Text:   "Да",
					Data:   fmt.Sprintf("smer_id:%s", s.ID)}},
			},
		},
		ParseMode: telebot.ModeHTML,
	})
}

func DeleteSMERForSure(c telebot.Context, config *common.BotConfig) error {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(context.Background())

	smerId := strings.Split(c.Data(), ":")[1]
	i := delete_smer_usecase.NewInteractor(adapter)
	err = i.Call(context.Background(), smerId)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка удаления данных: %s", err))
	}
	return c.Send(fmt.Sprintf("Запись удалена"))
}
