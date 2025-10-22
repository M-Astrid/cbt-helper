package common

import "github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/smer_steps"

func GetUserState(userID int64, config *BotConfig) smer_steps.UserState {
	if s, ok := config.UserStates[userID]; ok {
		return s
	}
	return smer_steps.UserState{}
}
