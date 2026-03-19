package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) sendDirKeyboard(chatID int64, caption string) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("RU → EN", "dir:ru_en"),
			tgbotapi.NewInlineKeyboardButtonData("EN → RU", "dir:en_ru"),
		),
	)

	m := tgbotapi.NewMessage(chatID, caption)
	m.ReplyMarkup = kb
	_, _ = b.api.Send(m)
}

// Кнопка-меню
func (b *Bot) sendMainMenu(chatID int64) {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Начать работу"),
			tgbotapi.NewKeyboardButton("Указать ФИО"),
			tgbotapi.NewKeyboardButton("Мой ID"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Рус → Англ"),
			tgbotapi.NewKeyboardButton("Англ → Рус"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Обычный режим"),
			tgbotapi.NewKeyboardButton("Режим чертежа"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Статус"),
			tgbotapi.NewKeyboardButton("Помощь"),
		),
	)

	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = kb
	_, _ = b.api.Send(msg)
}
