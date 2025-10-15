package utils

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция для отправки клавиатуры с выбором даты
func sendDatePicker(chatID int64, bot *tgbotapi.BotAPI) {
	// Создаем клавиатуру с кнопками для выбора месяца и года
	// Например, для текущего месяца
	now := time.Now()
	year := now.Year()
	month := now.Month()

	// Создаем кнопки для дней месяца
	daysInMonth := daysIn(month, year)
	var rows [][]tgbotapi.KeyboardButton

	for day := 1; day <= daysInMonth; day++ {
		btn := tgbotapi.NewKeyboardButton(fmt.Sprintf("%d-%02d-%02d", year, month, day))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(btn))
	}

	keyboard := tgbotapi.NewReplyKeyboard(rows...)

	msg := tgbotapi.NewMessage(chatID, "Выберите дату:")
	msg.ReplyMarkup = keyboard

	bot.Send(msg)
}

// Функция для подсчета количества дней в месяце
func daysIn(month time.Month, year int) int {
	// Проверка на високосность
	t := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}
