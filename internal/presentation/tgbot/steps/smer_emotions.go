package steps

import (
	"fmt"
	"strings"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/common"
	"gopkg.in/tucnak/telebot.v2"
)

type StepSMEREmotions struct {
	Step
}

func (ch StepSMEREmotions) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	state.CurrentSMERStepType = common.SMER_STEP_EMOTION
	bot.Send(rec, "Какие эмоции вы испытали?")
}

func (ch StepSMEREmotions) HandleInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) error {
	parts := strings.Split(m.Text, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		ems := strings.Split(p, " ")
		em, err := entity.NewEmotion(strings.TrimSpace(ems[0]), strings.TrimSpace(ems[1]))
		if err != nil {
			bot.Send(m.Sender, fmt.Sprintf("Ошибка сохранения эмоции: %v", err))
			return err
		}
		state.SMER.Emotions = append(state.SMER.Emotions, em)
	}
	bot.Send(m.Sender, fmt.Sprintf("Сохранили эмоции %v", parts))
	return nil
}
