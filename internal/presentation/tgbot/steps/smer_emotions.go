package steps

import (
	"fmt"
	"strings"

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

func (ch StepSMEREmotions) HandleTextInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) {
	parts := strings.Split(m.Text, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		parts[i] = p
		fmt.Println(p)
		// todo: save emotion
	}
	bot.Send(m.Sender, fmt.Sprintf("Сохранили эмоции %v", parts))
}
