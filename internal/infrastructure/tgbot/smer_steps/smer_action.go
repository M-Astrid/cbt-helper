package smer_steps

import (
	"gopkg.in/telebot.v3"
)

type StepSMERAction struct {
	Step
}

func (ch StepSMERAction) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	bot.Send(rec, "Какие были ваши действия?")
}

func (ch StepSMERAction) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	//state.SMER.Action = entity.NewAction(m.Text)
	return nil
}
