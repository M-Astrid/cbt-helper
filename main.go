package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	getUserSMERsUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/getUserSMERs"
	saveSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/saveSMER"
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	domainError "github.com/M-Astrid/cbt-helper/internal/domain/error"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/presentation/pdf/renderer"
	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/steps"
	"github.com/joho/godotenv"
	"gopkg.in/tucnak/telebot.v2"
)

var userStates = make(map[int64]steps.UserState)

var smerStepTrigger = &steps.StepSMERTrigger{}
var smerStepEmotions = &steps.StepSMEREmotions{}
var smerStepThoughts = &steps.StepSMERThoughts{}

var smerSteps = []steps.StepI{smerStepTrigger, smerStepEmotions, smerStepThoughts}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	ctx := context.Background()

	// Настройки подключения
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", os.Getenv("MONGODB_USER"), os.Getenv("MONGODB_PASSWORD"), os.Getenv("MONGODB_HOST"), os.Getenv("MONGODB_PORT"))
	dbName := os.Getenv("MONGODB_DB")

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Создаем inline-кнопки
	braindumpBtn := telebot.InlineButton{Unique: "brain_dump", Text: "Записать неструктурированные мысли"}
	writeSMERBtn := telebot.InlineButton{Unique: "write_smer", Text: "Сделать запись СМЭР"}
	getBraindumpBtn := telebot.InlineButton{Unique: "get_brain_dumps", Text: "Просмотреть неструктурированные заметки"}
	getSMERsPDFBtn := telebot.InlineButton{Unique: "get_smers", Text: "Выгрузить записи СМЭР"}

	// Отправляем сообщение с inline-клавиатурой
	bot.Handle("/start", func(m *telebot.Message) {
		inlineKeyboard := [][]telebot.InlineButton{
			{writeSMERBtn, braindumpBtn},
			{getBraindumpBtn, getSMERsPDFBtn},
		}
		bot.Send(m.Sender, "Чем займемся?", &telebot.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		})
	})

	// Обработка нажатия inline-кнопки
	bot.Handle(&writeSMERBtn, func(c *telebot.Callback) {
		userID := c.Sender.ID

		state := userStates[userID]
		smer := entity.NewSMEREntry(userID)
		state.SMER = smer
		state.SMERSteps = smerSteps
		state.CurrentSMERStepIdx = 0
		state.SMERSteps[state.CurrentSMERStepIdx].Start(bot, c.Sender, userID, &state)
		userStates[userID] = state
	})

	saveSMERBtn := telebot.InlineButton{Unique: "save_smer", Text: "Сохранить"}
	saveSMERinlineKeyboard := [][]telebot.InlineButton{
		{saveSMERBtn},
	}

	// smer steps input handler
	bot.Handle(telebot.OnText, func(m *telebot.Message) {
		userID := m.Sender.ID
		state, exists := userStates[userID]
		if exists && state.CurrentSMERStepIdx < len(state.SMERSteps) && state.CurrentSMERStepIdx >= 0 {
			// handle user inp
			err := state.SMERSteps[state.CurrentSMERStepIdx].HandleInput(bot, m, userID, &state)
			var ve domainError.ValidationError
			if err != nil && !errors.As(err, &ve) {
				bot.Send(m.Sender, "Непредвиденная ошибка, попробуйте сначала.")
				return
			}

			if err == nil {
				state.CurrentSMERStepIdx++
			}
			userStates[userID] = state

			if state.CurrentSMERStepIdx < len(state.SMERSteps) {
				//bot.Send(m.Sender, fmt.Sprintf("Переходим к шагу: %d, %v", state.CurrentSMERStepIdx, state.SMERSteps[state.CurrentSMERStepIdx]))
				state.SMERSteps[state.CurrentSMERStepIdx].Start(bot, m.Sender, userID, &state)
				return
			}

			bot.Send(m.Sender, fmt.Sprintf("Результат: %+v", *state.SMER), &telebot.ReplyMarkup{
				InlineKeyboard: saveSMERinlineKeyboard,
			})

			state.CurrentSMERStepType = -1
			state.CurrentSMERStepIdx = -1
			userStates[userID] = state

		} else {
			// Обработка других сообщений или игнор
			bot.Send(m.Sender, "Используйте команду /start для начала.")
		}
	})

	bot.Handle(&saveSMERBtn, func(c *telebot.Callback) {
		userID := c.Sender.ID
		state, exists := userStates[userID]
		if exists {
			adapter, err := storage.NewSMERStorage(uri, dbName)
			if err != nil {
				log.Fatal("Ошибка подключения к MongoDB:", err)
				bot.Send(c.Sender, fmt.Sprintf("Произошла ошибка. %s", err))
				return
			}
			i := saveSMERUsecase.NewInteractor(adapter)

			err = i.Call(ctx, state.SMER)
			if err == nil {
				bot.Send(c.Sender, "Успешно сохранили запись.")
			} else {
				bot.Send(c.Sender, fmt.Sprintf("Произошла ошибка. %s", err))
			}

			if err := adapter.Close(ctx); err != nil {
				log.Println("Ошибка закрытия соединения:", err)
			}
		}
	})

	bot.Handle(&getSMERsPDFBtn, func(c *telebot.Callback) {
		userID := c.Sender.ID
		adapter, err := storage.NewSMERStorage(uri, dbName)
		if err != nil {
			log.Fatal("Ошибка подключения к MongoDB:", err)
			bot.Send(c.Sender, fmt.Sprintf("Произошла ошибка. %s", err))
			return
		}
		i := getUserSMERsUsecase.NewInteractor(adapter)
		smers, err := i.Call(ctx, userID)
		if err := adapter.Close(ctx); err != nil {
			log.Println("Ошибка закрытия соединения:", err)
		}

		rnd := renderer.NewSmerRenderer()
		docBuff, err := rnd.Render(smers)
		if err != nil {
			bot.Send(c.Sender, fmt.Sprintf("Произошла ошибка. %s", err))
			log.Fatal(err)
			return
		}

		doc := &telebot.Document{
			File:     telebot.FromReader(bytes.NewReader(docBuff)),
			FileName: "smer.pdf",
		}
		_, err = bot.Send(c.Sender, doc)

		if err != nil {
			bot.Send(c.Sender, fmt.Sprintf("Произошла ошибка. %s", err))
		}

	})

	bot.Start()
}
