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
	"github.com/M-Astrid/cbt-helper/internal/presentation/renderer"
	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/common"
	"github.com/M-Astrid/cbt-helper/internal/presentation/tgbot/steps"
	"github.com/joho/godotenv"
	tb_cal "github.com/oramaz/telebot-calendar"
	"gopkg.in/telebot.v3"
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

	pref := telebot.Settings{
		Token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем inline-кнопки
	writeSMERBtn := telebot.InlineButton{Unique: "write_smer", Text: "Сделать запись СМЭР"}
	braindumpBtn := telebot.InlineButton{Unique: "brain_dump", Text: "Записать неструктурированные мысли"}
	getShortSMERsBtn := telebot.InlineButton{Unique: "get_short_smers", Text: "Просмотреть записи СМЭР"}
	getSMERsPDFBtn := telebot.InlineButton{Unique: "get_smers", Text: "Выгрузить файл СМЭР"}

	// Обработчик /start
	bot.Handle("/start", func(c telebot.Context) error {
		inlineKeyboard := [][]telebot.InlineButton{
			{writeSMERBtn, braindumpBtn},
			{getShortSMERsBtn, getSMERsPDFBtn},
		}
		return c.Send("Чем займемся?", &telebot.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		})
	})

	// Обработка нажатия кнопки "Сделать запись СМЭР"
	bot.Handle(&writeSMERBtn, func(c telebot.Context) error {
		userID := c.Sender().ID

		var state steps.UserState
		if s, ok := userStates[userID]; ok {
			state = s
		}
		smer := entity.NewSMEREntry(userID)
		state.SMER = smer
		state.SMERSteps = smerSteps
		state.CurrentSMERStepIdx = 0
		state.SMERSteps[state.CurrentSMERStepIdx].Start(bot, c.Sender(), userID, &state)
		userStates[userID] = state
		return nil
	})

	// Обработка текстовых сообщений
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		userID := c.Sender().ID
		if state, exists := userStates[userID]; exists {
			if state.CurrentSMERStepIdx >= 0 && state.CurrentSMERStepIdx < len(state.SMERSteps) {
				err := state.SMERSteps[state.CurrentSMERStepIdx].HandleInput(bot, c.Message(), userID, &state)
				var ve domainError.ValidationError
				if err != nil && !errors.As(err, &ve) {
					return c.Send("Непредвиденная ошибка, попробуйте сначала.")
				}
				if err == nil {
					state.CurrentSMERStepIdx++
				}
				userStates[userID] = state

				if state.CurrentSMERStepIdx < len(state.SMERSteps) {
					state.SMERSteps[state.CurrentSMERStepIdx].Start(bot, c.Sender(), userID, &state)
					return nil
				}
				state.CurrentSMERStepIdx = -1
				userStates[userID] = state

				rnd := renderer.NewSmerRenderer()
				msg, err := rnd.RenderMessageSingle(state.SMER)
				if err != nil {
					c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
				}

				return c.Send(fmt.Sprintf("Результат: %s", *msg), &telebot.ReplyMarkup{
					InlineKeyboard: [][]telebot.InlineButton{
						{telebot.InlineButton{Unique: "save_smer", Text: "Сохранить"}},
					},
				})
			}

			calendar := tb_cal.NewCalendar(bot, tb_cal.Options{})

			switch state.Status {
			case common.PICK_SHORT_DATE_FROM_STATUS:
				date, err := time.Parse("02.01.2006", c.Data())
				if err != nil {
					return err
				}
				state.StartDate = date
				state.Status = common.PICK_SHORT_DATE_TO_STATUS
				userStates[userID] = state
				return c.Send("Выберите дату окончания", &telebot.ReplyMarkup{
					InlineKeyboard: calendar.GetKeyboard(),
				})
			case common.PICK_DOC_DATE_FROM_STATUS:
				date, err := time.Parse("02.01.2006", c.Data())
				if err != nil {
					return err
				}
				state.StartDate = date
				state.Status = common.PICK_DOC_DATE_TO_STATUS
				userStates[userID] = state
				return c.Send("Выберите дату окончания", &telebot.ReplyMarkup{
					InlineKeyboard: calendar.GetKeyboard(),
				})
			case common.PICK_SHORT_DATE_TO_STATUS:
				date, err := time.Parse("02.01.2006", c.Data())
				if err != nil {
					return err
				}
				state.EndDate = date
				userStates[userID] = state
				//adapter, err := storage.NewSMERStorage(uri, dbName)
				//if err != nil {
				//	log.Println("Ошибка подключения к MongoDB:", err)
				//	return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
				//}
				//i := getUserSMERsUsecase.NewInteractor(adapter)
				//smers, err := i.Call(ctx, userID)
				//if err != nil {
				//	return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
				//}
				//if err := adapter.Close(ctx); err != nil {
				//	log.Println("Ошибка закрытия соединения:", err)
				//}
				return nil
			case common.PICK_DOC_DATE_TO_STATUS:
				date, err := time.Parse("02.01.2006", c.Data())
				if err != nil {
					return err
				}
				state.EndDate = date
				userStates[userID] = state
				adapter, err := storage.NewSMERStorage(uri, dbName)
				if err != nil {
					log.Println("Ошибка подключения к MongoDB:", err)
					return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
				}
				i := getUserSMERsUsecase.NewInteractor(adapter)
				smers, err := i.Call(ctx, userID, state.StartDate, state.EndDate)
				if err != nil {
					return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
				}
				if err := adapter.Close(ctx); err != nil {
					log.Println("Ошибка закрытия соединения:", err)
				}
				rnd := renderer.NewSmerRenderer()
				docBuff, err := rnd.RenderPDF(smers, state.StartDate, state.EndDate)
				if err != nil {
					log.Println("Ошибка рендеринга PDF:", err)
					return c.Send(fmt.Sprintf("Ошибка: %v", err))
				}

				doc := &telebot.Document{
					File:     telebot.FromReader(bytes.NewReader(docBuff)),
					FileName: "smer.pdf",
				}
				state.Status = -1
				userStates[userID] = state
				return c.Send(doc)
			}
		} else {
			return c.Send("Используйте команду /start для начала.")
		}

		return nil
	})

	// Обработка "Сохранить"
	bot.Handle(&telebot.InlineButton{Unique: "save_smer"}, func(c telebot.Context) error {
		userID := c.Sender().ID
		if state, exists := userStates[userID]; exists {
			adapter, err := storage.NewSMERStorage(uri, dbName)
			if err != nil {
				log.Println("Ошибка подключения к MongoDB:", err)
				return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
			}
			i := saveSMERUsecase.NewInteractor(adapter)
			err = i.Call(ctx, state.SMER)
			if err == nil {
				c.Send("Успешно сохранили запись.")
			} else {
				c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
			}
			if err := adapter.Close(ctx); err != nil {
				log.Println("Ошибка закрытия соединения:", err)
			}
		}
		return nil
	})

	getShortLastWeekBtn := telebot.InlineButton{Unique: "get_short_last_week", Text: "Последняя неделя"}
	getShortLast2WeeksBtn := telebot.InlineButton{Unique: "get_short_last_2_weeks", Text: "Последние две недели"}
	getShortLastMonthBtn := telebot.InlineButton{Unique: "get_short_last_month", Text: "Последний месяц"}
	getShortCustomDatesBtn := telebot.InlineButton{Unique: "get_short_custom_dates", Text: "Выбрать даты"}

	// Обработка "Просмотреть записи СМЭР" — выбор даты
	bot.Handle(&getShortSMERsBtn, func(c telebot.Context) error {
		inlineKeyboard := [][]telebot.InlineButton{
			{getShortLastWeekBtn, getShortLast2WeeksBtn},
			{getShortLastMonthBtn, getShortCustomDatesBtn},
		}
		return c.Send("За какой период нужна выгрузка?", &telebot.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		})
	})

	calendar := tb_cal.NewCalendar(bot, tb_cal.Options{})

	// Обработка "Просмотреть записи СМЭР" — выбор даты
	bot.Handle(&getShortCustomDatesBtn, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := userStates[userID]
		state.Status = common.PICK_SHORT_DATE_FROM_STATUS
		userStates[userID] = state
		return c.Send("Выберите дату начала", &telebot.ReplyMarkup{
			InlineKeyboard: calendar.GetKeyboard(),
		})
	})

	getDocLastWeekBtn := telebot.InlineButton{Unique: "get_doc_last_week", Text: "Последняя неделя"}
	getDocLast2WeeksBtn := telebot.InlineButton{Unique: "get_doc_last_2_weeks", Text: "Последние две недели"}
	getDocLastMonthBtn := telebot.InlineButton{Unique: "get_doc_last_month", Text: "Последний месяц"}
	getDocCustomDatesBtn := telebot.InlineButton{Unique: "get_doc_custom_dates", Text: "Выбрать даты"}

	bot.Handle(&getSMERsPDFBtn, func(c telebot.Context) error {
		inlineKeyboard := [][]telebot.InlineButton{
			{getDocLastWeekBtn, getDocLast2WeeksBtn},
			{getDocLastMonthBtn, getDocCustomDatesBtn},
		}
		return c.Send("За какой период нужна выгрузка?", &telebot.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		})
	})

	bot.Handle(&getDocLastWeekBtn, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := userStates[userID]
		adapter, err := storage.NewSMERStorage(uri, dbName)
		if err != nil {
			log.Println("Ошибка подключения к MongoDB:", err)
			return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
		}
		i := getUserSMERsUsecase.NewInteractor(adapter)
		now := time.Now()
		oneWeekAgo := now.AddDate(0, 0, -7)
		smers, err := i.Call(ctx, userID, oneWeekAgo, now)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
		}
		if err := adapter.Close(ctx); err != nil {
			log.Println("Ошибка закрытия соединения:", err)
		}
		rnd := renderer.NewSmerRenderer()
		docBuff, err := rnd.RenderPDF(smers, oneWeekAgo, now)
		if err != nil {
			log.Println("Ошибка рендеринга PDF:", err)
			return c.Send(fmt.Sprintf("Ошибка: %v", err))
		}

		doc := &telebot.Document{
			File:     telebot.FromReader(bytes.NewReader(docBuff)),
			FileName: "smer.pdf",
		}
		state.Status = -1
		userStates[userID] = state
		return c.Send(doc)
	})

	bot.Handle(&getDocLastMonthBtn, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := userStates[userID]
		adapter, err := storage.NewSMERStorage(uri, dbName)
		if err != nil {
			log.Println("Ошибка подключения к MongoDB:", err)
			return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
		}
		i := getUserSMERsUsecase.NewInteractor(adapter)
		now := time.Now()
		oneMonthAgo := now.AddDate(0, -1, 0)
		smers, err := i.Call(ctx, userID, oneMonthAgo, now)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
		}
		if err := adapter.Close(ctx); err != nil {
			log.Println("Ошибка закрытия соединения:", err)
		}
		rnd := renderer.NewSmerRenderer()
		docBuff, err := rnd.RenderPDF(smers, oneMonthAgo, now)
		if err != nil {
			log.Println("Ошибка рендеринга PDF:", err)
			return c.Send(fmt.Sprintf("Ошибка: %v", err))
		}

		doc := &telebot.Document{
			File:     telebot.FromReader(bytes.NewReader(docBuff)),
			FileName: "smer.pdf",
		}
		state.Status = -1
		userStates[userID] = state
		return c.Send(doc)
	})

	bot.Handle(&getDocCustomDatesBtn, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := userStates[userID]
		state.Status = common.PICK_DOC_DATE_FROM_STATUS
		userStates[8403079291] = state
		return c.Send("Выберите дату начала", &telebot.ReplyMarkup{
			InlineKeyboard: calendar.GetKeyboard(),
		})
	})

	bot.Start()
}
