package steps

import (
	"gopkg.in/telebot.v3"
)

type StepI interface {
	Start(bot *telebot.Bot, rec telebot.Recipient, userID int64, state *UserState)
	HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error
}

type Step struct {
}
