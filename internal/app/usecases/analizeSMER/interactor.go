package analizeSMER

import (
	"encoding/json"
	"errors"

	"github.com/M-Astrid/cbt-helper/internal/app/port"
)

type AnalizeSMERCmd struct {
	SMERID string
}

type AnalizeSMERResponse struct {
	Text string
}

type Interactor struct {
	storage  port.SMERStorage
	aiClient port.AIAssistant
}

func NewInteractor(storage port.SMERStorage, aiClient port.AIAssistant) *Interactor {
	return &Interactor{
		storage:  storage,
		aiClient: aiClient,
	}
}

func (i *Interactor) Call(cmd *AnalizeSMERCmd) (*AnalizeSMERResponse, error) {
	smer, err := i.storage.GetByID(cmd.SMERID)
	if err != nil {
		return nil, err
	}
	if smer == nil {
		return nil, errors.New("запись СМЭР не найдена")
	}

	jsonMap := map[string]interface{}{
		"триггер": smer.Trigger,
		"эмоции":  smer.Emotions,
		"мысли":   smer.Thoughts,
	}

	jsonData, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	res, err := i.aiClient.AnalizeSMER(string(jsonData))
	if err != nil {
		return nil, err
	}

	return &AnalizeSMERResponse{
		Text: res.Message,
	}, nil
}
