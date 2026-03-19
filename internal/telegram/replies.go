package telegram

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) sendTextOrFile(chatID int64, text, filename string) {
	if len([]rune(text)) <= 3500 {
		b.replyText(chatID, text)
		return
	}

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("tg_out_%d_%s", time.Now().UnixNano(), filename))
	_ = os.WriteFile(tmp, []byte(text), 0644)
	defer os.Remove(tmp)

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(tmp))
	doc.Caption = "Перевод (файлом, потому что текст длинный)"
	_, _ = b.api.Send(doc)
}

func (b *Bot) replyText(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	_, _ = b.api.Send(m)
}

func (b *Bot) replyCallback(cb *tgbotapi.CallbackQuery, text string) {
	cfg := tgbotapi.NewCallback(cb.ID, text)
	_, _ = b.api.Request(cfg)
}
