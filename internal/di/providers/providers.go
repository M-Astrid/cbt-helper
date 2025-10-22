package providers

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/app/port"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/smer_steps"
	tb_cal "github.com/oramaz/telebot-calendar"
	"gopkg.in/telebot.v3"
)

func ProvideConfig() *common.BotConfig {
	return &common.BotConfig{
		DBUri:      fmt.Sprintf("mongodb://%s:%s@%s:%s", os.Getenv("MONGODB_USER"), os.Getenv("MONGODB_PASSWORD"), os.Getenv("MONGODB_HOST"), os.Getenv("MONGODB_PORT")),
		DBName:     os.Getenv("MONGODB_DB"),
		Token:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		UserStates: make(map[int64]smer_steps.UserState),
		Calendar:   tb_cal.NewCalendar(nil, tb_cal.Options{}),
	}
}

func ProvideSMERStorage(config *common.BotConfig) port.SMERStorage {
	adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	if err != nil {
		panic(fmt.Sprintf("Ошибка подключения к MongoDB: %s", err))
	}
	return adapter
}

func ProvideBot(config *common.BotConfig) *telebot.Bot {
	pref := telebot.Settings{
		Token:  config.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	return bot
}

func ProvideCalendar(bot *telebot.Bot) *tb_cal.Calendar {
	return tb_cal.NewCalendar(bot, tb_cal.Options{})
}
