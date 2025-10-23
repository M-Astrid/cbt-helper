package http_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type HttpClient struct {
}

func NewHttpClient() *HttpClient {
	return &HttpClient{}
}

func (h *HttpClient) Post(url string, json_ interface{}, headers map[string]string) ([]byte, error) {
	// Сериализация в JSON
	jsonData, err := json.Marshal(json_)
	if err != nil {
		panic(err)
	}

	if err != nil {
		fmt.Println("Ошибка маршалинга JSON:", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if json_ != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка выполнения запроса:", err)
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return nil, err
	}
	s := string(bodyBytes)
	log.Println(s)
	return bodyBytes, nil
}
