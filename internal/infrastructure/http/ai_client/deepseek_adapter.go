package ai_client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/M-Astrid/cbt-helper/internal/app/dto"
)

type HttpClient interface {
	Post(url string, json_ interface{}, headers map[string]string) ([]byte, error)
}

type DeepSeekAdapter struct {
	apiKey     string
	authHeader map[string]string
	client     HttpClient
	baseURL    string
}

func (d DeepSeekAdapter) AnalizeSMER(smer string) (*dto.AnalizeSMERResponse, error) {
	prompt := fmt.Sprintf(`Проанализируй следующую запись на предмет когнитивных искажений, выяви типы искажения и предложи рекомендации по работе с мыслями:
"%s"`, smer)

	reqBody := RequestBody{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "system", Content: "Ты — психологический помощник, работающий в рамках КПТ, анализирующий записи для выявления когнитивных искажений и дающий рекомендации."},
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	respBytes, err := d.client.Post(d.baseURL+"/chat/completions", reqBody, d.authHeader)

	if err != nil {
		fmt.Println("Ошибка выполнения запроса:", err)
		return nil, err
	}

	var respBody ResponseBody
	if err := json.Unmarshal(respBytes, &respBody); err != nil {
		fmt.Println("Ошибка парсинга JSON ответа:", err)
		fmt.Println("Ответ:", string(respBytes))
		return nil, err
	}

	if len(respBody.Choices) > 0 {
		fmt.Println("Анализ нейросети:\n", respBody.Choices[0].Message.Content)
	} else {
		fmt.Println("Пустой ответ от нейросети.")
		return nil, errors.New("Пустой ответ от нейросети.")
	}
	return &dto.AnalizeSMERResponse{Message: respBody.Choices[0].Message.Content}, nil
}

func NewDeepSeekAdapter(apiKey string, client HttpClient, baseUrl string) *DeepSeekAdapter {
	return &DeepSeekAdapter{
		apiKey: apiKey,
		authHeader: map[string]string{
			"Authorization": "Bearer " + apiKey,
		},
		baseURL: baseUrl,
		client:  client,
	}
}
