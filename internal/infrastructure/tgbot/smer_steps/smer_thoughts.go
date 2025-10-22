package smer_steps

import (
	"strings"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"gopkg.in/telebot.v3"
)

type StepSMERThoughts struct {
	Step
}

func (ch StepSMERThoughts) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	bot.Send(rec, "Какие мысли приходили в голову?")
}

func (ch StepSMERThoughts) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	parts := strings.Split(m.Text, "\n")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		t, err := entity.NewThought(p)
		if err != nil {
			return err
		}
		state.SMER.Thoughts = append(state.SMER.Thoughts, t)
	}
	return nil
}
