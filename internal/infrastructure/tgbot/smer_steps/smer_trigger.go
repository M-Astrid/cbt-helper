package smer_steps

import (
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"gopkg.in/telebot.v3"
)

type StepSMERTrigger struct {
	Step
}

func (ch StepSMERTrigger) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	bot.Send(rec, "Что произошло и послужило триггером эмоций?")
}

func (ch StepSMERTrigger) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	state.SMER.Trigger = entity.NewTrigger(m.Text)
	return nil
}
