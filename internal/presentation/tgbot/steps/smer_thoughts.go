package steps

import (
	"fmt"
	"strings"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/common"
	"gopkg.in/tucnak/telebot.v2"
)

type StepSMERThoughts struct {
	Step
}

func (ch StepSMERThoughts) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	state.CurrentSMERStepType = common.SMER_STEP_THOUGHTS
	bot.Send(rec, "Какие мысли приходили в голову?")
}

func (ch StepSMERThoughts) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	parts := strings.Split(m.Text, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		t, err := entity.NewThought(p)
		if err != nil {
			return err
		}
		state.SMER.Thoughts = append(state.SMER.Thoughts, t)
	}
	bot.Send(m.Sender, fmt.Sprintf("Сохранили мысли %v", parts))
	return nil
}
