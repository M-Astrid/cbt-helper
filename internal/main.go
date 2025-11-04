package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/app/usecases/addNotesSMER"
	"github.com/M-Astrid/cbt-helper/internal/app/usecases/analizeSMER"
	deleteSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/deleteSMER"
	getSingleSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/getSingleSMER"
	getUserSMERsUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/getUserSMERs"
	saveSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/saveSMER"
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
	domainError "github.com/M-Astrid/cbt-helper/internal/domain/error"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/http/ai_client"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/http/http_client"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/renderer"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/steps"
	"github.com/joho/godotenv"
	tb_cal "github.com/oramaz/telebot-calendar"
	"gopkg.in/telebot.v3"
)

// Конфигурация и состояние бота
type BotConfig struct {
	DBUri      string
	DBName     string
	Token      string
	UserStates map[int64]steps.UserState
	Calendar   *tb_cal.Calendar
}

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("Не найден файл .env: %v", err)
	}

	ctx := context.Background()

	config := &BotConfig{
		DBUri:      fmt.Sprintf("mongodb://%s:%s@%s:%s", os.Getenv("MONGODB_USER"), os.Getenv("MONGODB_PASSWORD"), os.Getenv("MONGODB_HOST"), os.Getenv("MONGODB_PORT")),
		DBName:     os.Getenv("MONGODB_DB"),
		Token:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		UserStates: make(map[int64]steps.UserState),
		Calendar:   tb_cal.NewCalendar(nil, tb_cal.Options{}),
	}

	pref := telebot.Settings{
		Token:  config.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	config.Calendar = tb_cal.NewCalendar(bot, tb_cal.Options{})

	smerStorage, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Fatal(err)
	}
	defer smerStorage.Close(ctx)

	deepSeekAdapter := ai_client.NewDeepSeekAdapter(
		os.Getenv("DEEPSEEK_API_KEY"),
		http_client.NewHttpClient(),
		"https://api.deepseek.com",
	)

	// Регистрация обработчиков
	registerHandlers(bot, ctx, config, smerStorage, deepSeekAdapter)

	bot.Start()
}

func registerHandlers(
	bot *telebot.Bot,
	ctx context.Context,
	config *BotConfig,
	smerStorage *storage.SMERStorage,
	deepSeekAdapter *ai_client.DeepSeekAdapter,
) {

	writeSMERBtn := telebot.InlineButton{Unique: "write_smer", Text: "Сделать запись СМЭР"}
	braindumpBtn := telebot.InlineButton{Unique: "brain_dump", Text: "Записать неструктурированные мысли"}
	getShortSMERsBtn := telebot.InlineButton{Unique: "get_short_smers", Text: "Просмотреть записи СМЭР"}
	getSMERsPDFBtn := telebot.InlineButton{Unique: "get_smers", Text: "Выгрузить файл СМЭР"}

	// /start
	bot.Handle("/start", func(c telebot.Context) error {
		inlineKeyboard := [][]telebot.InlineButton{
			{writeSMERBtn, braindumpBtn},
			{getShortSMERsBtn, getSMERsPDFBtn},
		}
		return c.Send("Чем займемся?", &telebot.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		})
	})

	bot.Handle(&writeSMERBtn, func(c telebot.Context) error {
		return handleWriteSMER(c, config)
	})
	bot.Handle("/write_smer", func(c telebot.Context) error {
		return handleWriteSMER(c, config)
	})

	bot.Handle(&getSMERsPDFBtn, func(c telebot.Context) error {
		return handleDownloadSMER(c)
	})
	bot.Handle("/get_smers_pdf", func(c telebot.Context) error {
		return handleDownloadSMER(c)
	})

	bot.Handle(&getShortSMERsBtn, func(c telebot.Context) error {
		return handleGetSMERs(c)
	})
	bot.Handle("/get_smers_list", func(c telebot.Context) error {
		return handleGetSMERs(c)
	})

	// Обработка текстовых сообщений
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		return handleTextInput(c, ctx, config, smerStorage)
	})

	// Обработка "Сохранить"
	bot.Handle(&telebot.InlineButton{Unique: "save_smer"}, func(c telebot.Context) error {
		return handleSaveSMER(c, ctx, config)
	})

	// Обработка "Работать с мыслью"
	bot.Handle(&telebot.InlineButton{Unique: "work_smer"}, func(c telebot.Context) error {
		return handleWorkSMER(c, config)
	})

	// Обработка "AI анализ СМЭР"
	bot.Handle(&telebot.InlineButton{Unique: "analize_smer"}, func(c telebot.Context) error {
		return handleAnalizeSMER(c, config, analizeSMER.NewInteractor(smerStorage, deepSeekAdapter))
	})

	// Обработка "Удалить"
	bot.Handle(&telebot.InlineButton{Unique: "del_smer"}, func(c telebot.Context) error {
		return handleDeleteSMER(c, config)
	})
	bot.Handle(&telebot.InlineButton{Unique: "del_smer_for_sure"}, func(c telebot.Context) error {
		return handleDeleteSMERForSure(c, config)
	})

	// Обработка выгрузки по периоду
	registerPeriodHandlers(bot, ctx, config)
}

func handleAnalizeSMER(c telebot.Context, config *BotConfig, interactor *analizeSMER.Interactor) error {
	res, err := interactor.Call(&analizeSMER.AnalizeSMERCmd{SMERID: strings.Split(c.Data(), ":")[1]})
	if err != nil {
		return c.Send(err)
	}
	return c.Send(res.Text, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
}

func handleWriteSMER(c telebot.Context, config *BotConfig) error {
	userID := c.Sender().ID
	var state steps.UserState
	if s, ok := config.UserStates[userID]; ok {
		state = s
	}
	smer := entity.NewSMEREntry(userID)
	state.SMER = smer
	state.SMERSteps = []steps.StepI{&steps.StepSMERTrigger{}, &steps.StepSMEREmotions{}, &steps.StepSMERThoughts{}}
	state.CurrentSMERStepIdx = 0
	state.SMERSteps[state.CurrentSMERStepIdx].Start(c.Bot(), c.Sender(), userID, &state)
	config.UserStates[userID] = state
	return nil
}

func handleTextInput(c telebot.Context, ctx context.Context, config *BotConfig, smerStorage *storage.SMERStorage) error {
	userID := c.Sender().ID
	state, exists := config.UserStates[userID]

	if !exists {
		return c.Send("Используйте команду /start для начала.")
	}

	// Обработка шагов СМЭР
	if state.CurrentSMERStepIdx >= 0 && state.CurrentSMERStepIdx < len(state.SMERSteps) {
		err := state.SMERSteps[state.CurrentSMERStepIdx].HandleInput(c.Bot(), c.Message(), userID, &state)
		var ve domainError.ValidationError
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
		renderAndSendSmer(c, state)
	}

	if state.Status == common.WAIT_FOR_SMER_NOTES {
		return handleSMERNotes(c, config, ctx, addNotesSMER.NewInteractor(smerStorage))
	}

	// Обработка выбора дат и выгрузки
	err := handleDateSelection(c, &state, config)
	return err
}

func handleSaveSMER(c telebot.Context, ctx context.Context, config *BotConfig) error {
	userID := c.Sender().ID
	state, exists := config.UserStates[userID]
	if !exists {
		return nil
	}
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(ctx)

	i := saveSMERUsecase.NewInteractor(adapter)
	if err := i.Call(ctx, state.SMER); err != nil {
		c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
	} else {
		c.Send("Успешно сохранили запись.")
	}
	return nil
}

func handleWorkSMER(c telebot.Context, config *BotConfig) error {
	userID := c.Sender().ID
	state := config.UserStates[userID]
	state.Status = common.WAIT_FOR_SMER_NOTES
	state.WorkingSMERID = strings.Split(c.Data(), ":")[1]
	config.UserStates[userID] = state
	return c.Send("Проанализируйте мысли и запишите результат в свободной форме. " +
		"Чтобы удалить анализ, отправьте '-'.")

}

func handleSMERNotes(c telebot.Context, config *BotConfig, ctx context.Context, interactor *addNotesSMER.Interactor) error {
	msg := &c.Message().Text
	if *msg == "-" {
		msg = nil
	}
	state := config.UserStates[c.Sender().ID]
	err := interactor.Call(ctx, state.WorkingSMERID, msg)
	state.Status = 0
	config.UserStates[c.Sender().ID] = state
	if err != nil {
		c.Send(err)
	}
	return c.Send("Сохранили анализ")
}

func toDayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func registerPeriodHandlers(bot *telebot.Bot, ctx context.Context, config *BotConfig) {
	// Обработчики для выгрузки по периодам
	// tommorow := time.Now().AddDate(0, 0, 1)
	aMonthAgo := time.Now().AddDate(0, -1, 0)
	aWeekAgo := time.Now().AddDate(0, -7, 0)

	bot.Handle(&telebot.InlineButton{Unique: "get_short_last_week", Text: ""}, func(c telebot.Context) error {
		return sendSMERSInMessages(c, toDayStart(aWeekAgo), time.Now(), config)
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_short_last_month", Text: ""}, func(c telebot.Context) error {
		return sendSMERSInMessages(c, toDayStart(aMonthAgo), time.Now(), config)
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_doc_last_week", Text: ""}, func(c telebot.Context) error {
		return handlePeriodPDF(c, ctx, config, toDayStart(aWeekAgo), time.Now(), "smer.pdf")
	})

	bot.Handle(&telebot.InlineButton{Unique: "get_doc_last_month", Text: ""}, func(c telebot.Context) error {
		return handlePeriodPDF(c, ctx, config, toDayStart(aMonthAgo), time.Now(), "smer.pdf")
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_doc_custom_dates", Text: ""}, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := getUserState(userID, config)
		state.Status = common.PICK_DOC_DATE_FROM_STATUS
		config.UserStates[8403079291] = state // todo: find out why calendar set this id to user
		return c.Send("Выберите дату начала", &telebot.ReplyMarkup{
			InlineKeyboard: config.Calendar.GetKeyboard(),
		})
	})
	bot.Handle(&telebot.InlineButton{Unique: "get_short_custom_dates", Text: ""}, func(c telebot.Context) error {
		userID := c.Sender().ID
		state := getUserState(userID, config)
		state.Status = common.PICK_SHORT_DATE_FROM_STATUS
		config.UserStates[8403079291] = state // todo: find out why calendar set this id to user
		return c.Send("Выберите дату начала", &telebot.ReplyMarkup{
			InlineKeyboard: config.Calendar.GetKeyboard(),
		})
	})
}

func handlePeriodPDF(c telebot.Context, ctx context.Context, config *BotConfig, from, to time.Time, filename string) error {
	userID := c.Sender().ID
	state := getUserState(userID, config)
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(ctx)

	i := getUserSMERsUsecase.NewInteractor(adapter)
	smers, err := i.Call(ctx, userID, from, to)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
	}

	rnd := renderer.NewSmerRenderer()
	docBuff, err := rnd.RenderPDF(smers, from, to)
	if err != nil {
		log.Println("Ошибка рендеринга PDF:", err)
		return c.Send(fmt.Sprintf("Ошибка: %v", err))
	}

	doc := &telebot.Document{
		File:     telebot.FromReader(bytes.NewReader(docBuff)),
		FileName: filename,
	}
	state.Status = -1
	config.UserStates[userID] = state
	return c.Send(doc)
}

func getUserState(userID int64, config *BotConfig) steps.UserState {
	if s, ok := config.UserStates[userID]; ok {
		return s
	}
	return steps.UserState{}
}

func renderAndSendSmer(c telebot.Context, state steps.UserState) {
	rnd := renderer.NewSmerRenderer()
	msg, err := rnd.RenderMessageSingle(state.SMER)
	if err != nil {
		c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
		return
	}
	c.Send(fmt.Sprintf("%s", *msg), &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{telebot.InlineButton{Unique: "save_smer", Text: "Сохранить"}},
			},
		},
		ParseMode: telebot.ModeHTML,
	})
}

func parseDate(input string) (time.Time, error) {
	return time.Parse("02.01.2006", input)
}

func handleDateSelection(c telebot.Context, state *steps.UserState, config *BotConfig) error {
	switch state.Status {
	case common.PICK_SHORT_DATE_FROM_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.StartDate = date
		state.Status = common.PICK_SHORT_DATE_TO_STATUS
		config.UserStates[c.Sender().ID] = *state
		return c.Send("Выберите дату окончания", &telebot.ReplyMarkup{
			InlineKeyboard: config.Calendar.GetKeyboard(),
		})
	case common.PICK_SHORT_DATE_TO_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.EndDate = date
		config.UserStates[c.Sender().ID] = *state
		return sendSMERSInMessages(c, state.StartDate, state.EndDate, config)
	case common.PICK_DOC_DATE_FROM_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.StartDate = date
		state.Status = common.PICK_DOC_DATE_TO_STATUS
		config.UserStates[c.Sender().ID] = *state
		return c.Send("Выберите дату окончания", &telebot.ReplyMarkup{
			InlineKeyboard: config.Calendar.GetKeyboard(),
		})
	case common.PICK_DOC_DATE_TO_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.EndDate = date
		config.UserStates[c.Sender().ID] = *state
		return generateAndSendPDF(c, c.Sender().ID, *state, config)
	}
	return nil
}

func generateAndSendPDF(c telebot.Context, userID int64, state steps.UserState, config *BotConfig) error {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(context.Background())

	i := getUserSMERsUsecase.NewInteractor(adapter)
	smers, err := i.Call(context.Background(), userID, state.StartDate, state.EndDate)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
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
	return c.Send(doc)
}

func sendSMERSInMessages(c telebot.Context, from, to time.Time, config *BotConfig) error {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(context.Background())

	i := getUserSMERsUsecase.NewInteractor(adapter)
	smers, err := i.Call(context.Background(), c.Sender().ID, from, to)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
	}

	rnd := renderer.NewSmerRenderer()
	for _, s := range smers {
		message, err := rnd.RenderShortSingle(s)
		if err != nil {
			log.Fatal(err)
		}
		err = c.Send(*message, &telebot.SendOptions{
			ReplyMarkup: &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{telebot.InlineButton{
						Unique: "del_smer",
						Text:   "Удалить",
						Data:   fmt.Sprintf("smer_id:%s", s.ID)}},
					{
						telebot.InlineButton{
							Unique: "work_smer",
							Text:   "Самостоятельная работа",
							Data:   fmt.Sprintf("smer_id:%s", s.ID),
						},
						telebot.InlineButton{
							Unique: "analize_smer",
							Text:   "AI анализ",
							Data:   fmt.Sprintf("smer_id:%s", s.ID)},
					},
				},
			},
			ParseMode: telebot.ModeHTML,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func handleDeleteSMER(c telebot.Context, config *BotConfig) error {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(context.Background())

	smerId := strings.Split(c.Data(), ":")[1]
	i := getSingleSMERUsecase.NewInteractor(adapter)
	s, err := i.Call(context.Background(), smerId)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка получения данных: %s", err))
	}

	rnd := renderer.NewSmerRenderer()
	smerRepr, err := rnd.RenderShortSingle(s)
	if err != nil {
		return c.Send(fmt.Sprintf("Произошла ошибка %s", err))
	}
	message := fmt.Sprintf("Удалить эту запись без возможности восстановления?\n\n%s", *smerRepr)
	return c.Send(message, &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{telebot.InlineButton{
					Unique: "del_smer_for_sure",
					Text:   "Да",
					Data:   fmt.Sprintf("smer_id:%s", s.ID)}},
			},
		},
		ParseMode: telebot.ModeHTML,
	})
}

func handleDeleteSMERForSure(c telebot.Context, config *BotConfig) error {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		log.Println("Ошибка подключения к MongoDB:", err)
		return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	}
	defer adapter.Close(context.Background())

	smerId := strings.Split(c.Data(), ":")[1]
	i := deleteSMERUsecase.NewInteractor(adapter)
	err = i.Call(context.Background(), smerId)
	if err != nil {
		return c.Send(fmt.Sprintf("Ошибка удаления данных: %s", err))
	}
	return c.Send(fmt.Sprintf("Запись удалена"))
}

func handleDownloadSMER(c telebot.Context) error {
	getDocLastWeekBtn := telebot.InlineButton{Unique: "get_doc_last_week", Text: "Последняя неделя"}
	//getDocLast2WeeksBtn := telebot.InlineButton{Unique: "get_doc_last_2_weeks", Text: "Последние две недели"}
	getDocLastMonthBtn := telebot.InlineButton{Unique: "get_doc_last_month", Text: "Последний месяц"}
	getDocCustomDatesBtn := telebot.InlineButton{Unique: "get_doc_custom_dates", Text: "Выбрать даты"}

	inlineKeyboard := [][]telebot.InlineButton{
		{getDocLastWeekBtn, getDocLastMonthBtn},
		{getDocCustomDatesBtn},
	}
	return c.Send("За какой период нужны выгрузка?", &telebot.ReplyMarkup{
		InlineKeyboard: inlineKeyboard,
	})
}

func handleGetSMERs(c telebot.Context) error {
	getShortLastWeekBtn := telebot.InlineButton{Unique: "get_short_last_week", Text: "Последняя неделя"}
	//getShortLast2WeeksBtn := telebot.InlineButton{Unique: "get_short_last_2_weeks", Text: "Последние две недели"}
	getShortLastMonthBtn := telebot.InlineButton{Unique: "get_short_last_month", Text: "Последний месяц"}
	getShortCustomDatesBtn := telebot.InlineButton{Unique: "get_short_custom_dates", Text: "Выбрать даты"}

	inlineKeyboard := [][]telebot.InlineButton{
		{getShortLastWeekBtn, getShortLastMonthBtn},
		{getShortCustomDatesBtn},
	}
	return c.Send("За какой период нужны записи?", &telebot.ReplyMarkup{
		InlineKeyboard: inlineKeyboard,
	})
}
