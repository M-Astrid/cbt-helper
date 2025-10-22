package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	delete_smer_usecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/del_smer"
	getSingleSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/get_smer"
	save_smer_usecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/save_smer"
	"github.com/M-Astrid/cbt-helper/internal/di/providers"
	domain_error "github.com/M-Astrid/cbt-helper/internal/domain/error"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/handlers"
	"github.com/joho/godotenv"
	tb_cal "github.com/oramaz/telebot-calendar"
	"go.uber.org/dig"
	"gopkg.in/telebot.v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	ctx := context.Background()

	container := createDIContainer()

	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig) {
		registerHandlers(ctx, container)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigs
			log.Println("Завершение работы бота...")
			container.Invoke(func(storage *storage.SMERStorage) {
				storage.Close(ctx)
			})
			bot.Stop()
			os.Exit(0)
		}()

		bot.Start()
	})
}

func createDIContainer() *dig.Container {
	container := dig.New()
	container.Provide(providers.ProvideConfig)
	container.Provide(providers.ProvideBot)
	container.Provide(providers.ProvideCalendar)

	container.Provide(providers.ProvideSMERStorage)

	container.Provide(providers.ProvideDeleteSMERUsecase)
	container.Provide(providers.ProvideGetSingleSMERUsecase)
	container.Provide(providers.ProvideGetUserSMERsUsecase)
	container.Provide(providers.ProvideSaveSMERUsecase)
	return container
}

func registerHandlers(ctx context.Context, container *dig.Container) {
	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig) {
		bot.Handle("/writeSMER", func(c telebot.Context) error {
			return handlers.WriteSMER(c, config)
		})

		bot.Handle("/downloadSMERs", func(c telebot.Context) error {
			return handlers.SetupDownloadSMER(c)
		})

		bot.Handle("/getSMERs", func(c telebot.Context) error {
			return handlers.SetupGetSMERsShort(c)
		})
	})
	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig, calendar *tb_cal.Calendar) {
		registerPeriodHandlers(bot, ctx, config, calendar)
	})
	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig) {
		bot.Handle(telebot.OnText, func(c telebot.Context) error {
			return handleTextInput(c, config, container)
		})
	})
	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig, interactor *getSingleSMERUsecase.Interactor) {
		bot.Handle(&telebot.InlineButton{Unique: "del_smer"}, func(c telebot.Context) error {
			return handlers.DeleteSMER(c, config)
		})
	})
	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig, interactor *delete_smer_usecase.Interactor) {
		bot.Handle(&telebot.InlineButton{Unique: "del_smer_for_sure"}, func(c telebot.Context) error {
			return handlers.DeleteSMERForSure(c, config)
		})
	})
	container.Invoke(func(bot *telebot.Bot, config *common.BotConfig, interactor *save_smer_usecase.Interactor) {
		bot.Handle(&telebot.InlineButton{Unique: "work_smer"}, func(c telebot.Context) error {
			return handlers.WorkSMER(c, interactor)
		})
	})
}

func handleTextInput(c telebot.Context, config *common.BotConfig, container *dig.Container) error {
	userID := c.Sender().ID
	state, exists := config.UserStates[userID]

	if !exists {
		return c.Send("Используйте команду /start для начала.")
	}

	// Обработка шагов СМЭР
	if state.CurrentSMERStepIdx >= 0 && state.CurrentSMERStepIdx < len(state.SMERSteps) {
		err := state.SMERSteps[state.CurrentSMERStepIdx].HandleInput(c.Bot(), c.Message(), userID, &state)
		var ve domain_error.ValidationError
		if err != nil && !errors.As(err, &ve) {
			return c.Send("Непредвиденная ошибка, попробуйте сначала.")
		}
		if err == nil {
			state.CurrentSMERStepIdx++
		}
		config.UserStates[userID] = state

		if state.CurrentSMERStepIdx < len(state.SMERSteps) {
			state.SMERSteps[state.CurrentSMERStepIdx].Start(c.Bot(), c.Sender(), userID, &state)
			return nil
		}
		state.CurrentSMERStepIdx = -1
		config.UserStates[userID] = state

		// Генерация и отправка результата
		container.Invoke(func(interactor *save_smer_usecase.Interactor) {
			handlers.RenderAndSaveSmer(c, state, interactor)
		})

	}

	// Обработка выбора дат и выгрузки
	err := handlers.HandleDateSelection(c, &state, config)
	return err
}

func toDayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func registerPeriodHandlers(bot *telebot.Bot, ctx context.Context, config *common.BotConfig, calendar *tb_cal.Calendar) {
	// Обработчики для выгрузки по периодам
	// tommorow := time.Now().AddDate(0, 0, 1)
	aMonthAgo := time.Now().AddDate(0, -1, 0)
	aWeekAgo := time.Now().AddDate(0, -7, 0)

	bot.Handle(&telebot.InlineButton{Unique: "get_short_last_week", Text: ""}, func(c telebot.Context) error {
		return handlers.SendSMERSInMessages(c, toDayStart(aWeekAgo), time.Now(), config)
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_short_last_month", Text: ""}, func(c telebot.Context) error {
		return handlers.SendSMERSInMessages(c, toDayStart(aMonthAgo), time.Now(), config)
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_doc_last_week", Text: ""}, func(c telebot.Context) error {
		return handlers.DownloadPDFByPeriod(c, ctx, config, toDayStart(aWeekAgo), time.Now(), "smer.pdf")
	})

	bot.Handle(&telebot.InlineButton{Unique: "get_doc_last_month", Text: ""}, func(c telebot.Context) error {
		return handlers.DownloadPDFByPeriod(c, ctx, config, toDayStart(aMonthAgo), time.Now(), "smer.pdf")
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_doc_custom_dates", Text: ""}, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := common.GetUserState(userID, config)
		state.Status = common.PICK_DOC_DATE_FROM_STATUS
		config.UserStates[8403079291] = state // todo: find out why calendar set this id to user
		return c.Send("Выберите дату начала", &telebot.ReplyMarkup{
			InlineKeyboard: calendar.GetKeyboard(),
		})
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_short_custom_dates", Text: ""}, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := common.GetUserState(userID, config)
		state.Status = common.PICK_SHORT_DATE_FROM_STATUS
		config.UserStates[8403079291] = state // todo: find out why calendar set this id to user
		return c.Send("Выберите дату начала", &telebot.ReplyMarkup{
			InlineKeyboard: calendar.GetKeyboard(),
		})
	})
}
