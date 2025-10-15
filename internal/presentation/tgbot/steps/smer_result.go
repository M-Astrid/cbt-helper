package steps

import (
	"errors"
	"fmt"

	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/common"
	"gopkg.in/tucnak/telebot.v2"
)

type StepSMERResult struct {
	Step
}

func (ch StepSMERResult) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	state.CurrentSMERStepType = common.SMER_STEP_SAVE_RESULT
	bot.Send(rec, fmt.Sprintf("Результат: %v", state.SMER))

	saveBtn := telebot.InlineButton{Unique: "save_smer", Text: "Сохранить"}

	inlineKeyboard := [][]telebot.InlineButton{
		{saveBtn},
	}
	bot.Send(rec, fmt.Sprintf("Результат: %v", state.SMER), &telebot.ReplyMarkup{
		InlineKeyboard: inlineKeyboard,
	})
}

func (ch StepSMERResult) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	return errors.New("Method not implemented")
}
