package common

import (
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/smer_steps"
)

type BotConfig struct {
	DBUri      string
	DBName     string
	Token      string
	UserStates map[int64]smer_steps.UserState
}
