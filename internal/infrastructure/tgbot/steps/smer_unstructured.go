package steps

import (
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"gopkg.in/telebot.v3"
)

type StepSMERUnstructured struct {
	Step
}

func (ch StepSMERUnstructured) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	bot.Send(rec, "Что произошло и послужило триггером эмоций?")
}

func (ch StepSMERUnstructured) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	state.SMER.Unstructured = entity.NewUnstructured(m.Text)
	return nil
}
