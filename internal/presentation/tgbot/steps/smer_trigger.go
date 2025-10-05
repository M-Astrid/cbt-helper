package steps

import (
	"fmt"

	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/common"
	"gopkg.in/tucnak/telebot.v2"
)

type StepSMERTrigger struct {
	Step
}

func (ch StepSMERTrigger) Start(bot *telebot.Bot, rec telebot.Recipient, _ int64, state *UserState) {
	state.CurrentSMERStepType = common.SMER_STEP_TRIGGER
	bot.Send(rec, "Что произошло и послужило триггером эмоций?")
}

func (ch StepSMERTrigger) HandleTextInput(bot *telebot.Bot, m *telebot.Message, userID int64, state *UserState) {
	//storage.TriggerRepository.Save(&entity.Trigger{
	//
	//})
	bot.Send(m.Sender, fmt.Sprintf("Сохранили вашу ситуацию: %s", m.Text))
}
