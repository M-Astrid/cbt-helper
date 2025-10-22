package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/app/usecases/get_smers"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/renderer"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/smer_steps"
	tb_cal "github.com/oramaz/telebot-calendar"
	"gopkg.in/telebot.v3"
)

func DownloadPDFByPeriod(c telebot.Context, ctx context.Context, config *common.BotConfig, from, to time.Time, filename string) error {
	userID := c.Sender().ID
	state := main2.getUserState(userID, config)
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

func parseDate(input string) (time.Time, error) {
	return time.Parse("02.01.2006", input)
}

func HandleDateSelection(c telebot.Context, state *smer_steps.UserState, config *common.BotConfig, calendar *tb_cal.Calendar) error {
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
			InlineKeyboard: calendar.GetKeyboard(),
		})
	case common.PICK_SHORT_DATE_TO_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.EndDate = date
		config.UserStates[c.Sender().ID] = *state
		return SendSMERSInMessages(c, state.StartDate, state.EndDate, config)
	case common.PICK_DOC_DATE_FROM_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.StartDate = date
		state.Status = common.PICK_DOC_DATE_TO_STATUS
		config.UserStates[c.Sender().ID] = *state
		return c.Send("Выберите дату окончания", &telebot.ReplyMarkup{
			InlineKeyboard: calendar.GetKeyboard(),
		})
	case common.PICK_DOC_DATE_TO_STATUS:
		date, err := parseDate(c.Data())
		if err != nil {
			return err
		}
		state.EndDate = date
		config.UserStates[c.Sender().ID] = *state
		return GenerateAndSendPDF(c, c.Sender().ID, *state, config)
	}
	return nil
}

func GenerateAndSendPDF(c telebot.Context, userID int64, state smer_steps.UserState, config *common.BotConfig) error {
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

func SetupDownloadSMER(c telebot.Context) error {
	getDocLastWeekBtn := telebot.InlineButton{Unique: "get_doc_last_week", Text: "Последняя неделя"}
	//getDocLast2WeeksBtn := telebot.InlineButton{Unique: "get_doc_last_2_weeks", Text: "Последние две недели"}
	getDocLastMonthBtn := telebot.InlineButton{Unique: "get_doc_last_month", Text: "Последний месяц"}
	getDocCustomDatesBtn := telebot.InlineButton{Unique: "get_doc_custom_dates", Text: "Выбрать даты"}

	inlineKeyboard := [][]telebot.InlineButton{
		{getDocLastWeekBtn, getDocLastMonthBtn},
		{getDocCustomDatesBtn},
	}
	return c.Send("За какой период нужна выгрузка?", &telebot.ReplyMarkup{
		InlineKeyboard: inlineKeyboard,
	})
}

func SendSMERSInMessages(c telebot.Context, from, to time.Time, config *common.BotConfig) error {
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
					{
						telebot.InlineButton{
							Unique: "add_work_smer",
							Text:   "Работать с мыслями",
							Data:   fmt.Sprintf("smer_id:%s", s.ID)},
						telebot.InlineButton{
							Unique: "ai_analize_smer",
							Text:   "AI анализ",
							Data:   fmt.Sprintf("smer_id:%s", s.ID)},
					},
					{
						telebot.InlineButton{
							Unique: "del_smer",
							Text:   "🗑 Удалить",
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

func SetupGetSMERsShort(c telebot.Context) error {
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
