package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/M-Astrid/cbt-helper/internal/app/usecases/get_smer"
	save_smer_usecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/save_smer"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/storage"
	"github.com/M-Astrid/cbt-helper/internal/infrastructure/tgbot/common"
	"gopkg.in/telebot.v3"
)

func WorkSMER(c telebot.Context, interactor *save_smer_usecase.Interactor) error {
	//userID := c.Sender().ID
	//smerId := strings.Split(c.Data(), ":")[1]
	//
	//state := config.UserStates[userID]
	//state.Status = common.WORK_WITH_THOUGHTS_STATUS
	//config.UserStates[userID] = state
	//
	//adapter, err := storage.NewSMERStorage(config.DBUri, config.DBName)
	//if err != nil {
	//	log.Println("Ошибка подключения к MongoDB:", err)
	//	return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
	//}
	//defer adapter.Close(ctx)
	//
	//i := getSingleSMERUsecase.NewInteractor(adapter)
	//if s, err := i.Call(ctx, smerId); err != nil {
	//	c.Send(fmt.Sprintf("Произошла ошибка: %s", err))
	//} else {
	//	c.Send("Успешно сохранили запись.")
	//}
	return c.Send(fmt.Sprintf("Произошла ошибка: %v", err))
}
