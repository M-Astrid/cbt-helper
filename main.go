package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/steps"
	"github.com/joho/godotenv"
	"gopkg.in/tucnak/telebot.v2"
)

var userStates = make(map[int64]steps.UserState)

var smerStepTrigger = &steps.StepSMERTrigger{}
var smerStepEmotions = &steps.StepSMEREmotions{}

var smerSteps = []steps.StepI{smerStepTrigger, smerStepEmotions}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Создаем inline-кнопки
	setup_smer_btn := telebot.InlineButton{Unique: "setup_smer", Text: "Настроить блоки для СМЭР"}
	write_smer_btn := telebot.InlineButton{Unique: "write_smer", Text: "Сделать запись СМЭР"}

	// Отправляем сообщение с inline-клавиатурой
	bot.Handle("/start", func(m *telebot.Message) {
		inlineKeyboard := [][]telebot.InlineButton{
			{setup_smer_btn, write_smer_btn},
		}
		bot.Send(m.Sender, "Чем займемся?", &telebot.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		})
	})

	// Обработка нажатия inline-кнопки
	bot.Handle(&write_smer_btn, func(c *telebot.Callback) {
		userID := c.Sender.ID

		state := userStates[userID]
		smer := entity.NewSMEREntry()
		state.SMERID = smer.ID
		//state.SMERSteps = storage.SMERStorage.GetUserSteps(userID) // todo:
		state.SMERSteps = smerSteps
		state.CurrentSMERStepIdx = 0
		state.SMERSteps[state.CurrentSMERStepIdx].Start(bot, c.Sender, userID, &state)
		userStates[userID] = state
	})

	// smer steps input handler
	bot.Handle(telebot.OnText, func(m *telebot.Message) {
		userID := m.Sender.ID
		state, exists := userStates[userID]
		if exists && state.CurrentSMERStepIdx < len(state.SMERSteps) {
			// handle user inp
			state.SMERSteps[state.CurrentSMERStepIdx].HandleTextInput(bot, m, userID, &state)

			// switch to the next step
			state.CurrentSMERStepIdx++
			userStates[userID] = state
			if state.CurrentSMERStepIdx < len(state.SMERSteps) {
				bot.Send(m.Sender, fmt.Sprintf("Переходим к шагу: %d, %v", state.CurrentSMERStepIdx, state.SMERSteps[state.CurrentSMERStepIdx]))
				state.SMERSteps[state.CurrentSMERStepIdx].Start(bot, m.Sender, userID, &state)
			}

		} else {
			// Обработка других сообщений или игнор
			bot.Send(m.Sender, "Используйте команду /start для начала.")
		}
	})

	bot.Start()
}
