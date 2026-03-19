package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/pha06ntom/tg-translator-bot/internal/db"
	"github.com/pha06ntom/tg-translator-bot/internal/files"
	"github.com/pha06ntom/tg-translator-bot/internal/translate"
)

func (b *Bot) handleUpdate(ctx context.Context, upd tgbotapi.Update) {
	if upd.CallbackQuery != nil {
		cb := upd.CallbackQuery

		if cb.From != nil {
			_ = db.UpsertUser(ctx, b.db, db.UserUpsert{
				TelegramUserID: cb.From.ID,
				Username:       cb.From.UserName,
				DisplayName:    strings.TrimSpace(cb.From.FirstName + " " + cb.From.LastName),
			})
		}

		if !b.isAllowed(ctx, cb.From.ID) {
			b.replyCallback(cb, "⛔️ Нет доступа.")
			return
		}

		if !b.hasRealName(ctx, cb.From.ID) {
			b.replyCallback(cb, "Сначала укажите имя и фамилию командой /set_name")
			return
		}

		b.handleCallback(ctx, cb)
		return
	}

	msg := upd.Message
	if msg == nil || msg.From == nil {
		return
	}

	_ = db.UpsertUser(ctx, b.db, db.UserUpsert{
		TelegramUserID: msg.From.ID,
		Username:       msg.From.UserName,
		DisplayName:    strings.TrimSpace(msg.From.FirstName + " " + msg.From.LastName),
	})

	if msg.IsCommand() {
		b.handleCommand(ctx, msg)
		return
	}

	if !b.isAllowed(ctx, msg.From.ID) {
		b.replyText(msg.Chat.ID, "⛔️ Нет доступа. Попросите администратора выдать доступ.")
		return
	}

	if !b.hasRealName(ctx, msg.From.ID) {
		b.replyText(msg.Chat.ID, "Для работы с ботом нужно указать имя и фамилию:\n/set_name Иван Иванов")
		return
	}

	if msg.Document != nil {
		b.handleDocument(ctx, msg)
		return
	}

	if len(msg.Photo) > 0 {
		b.handlePhoto(ctx, msg)
		return
	}

	if strings.TrimSpace(msg.Text) != "" {
		switch msg.Text {
		case "Начать работу":
			if !b.isAllowed(ctx, msg.From.ID) {
				b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
				return
			}
			if !b.hasRealName(ctx, msg.From.ID) {
				b.replyText(msg.Chat.ID, "Для работы с ботом нужно указать имя и фамилию:\n/set_name Иван Иванов")
				return
			}
			b.replyText(msg.Chat.ID, "Выберите направление и режим, затем отправьте текст, документ, PDF или изображение.")
			b.sendMainMenu(msg.Chat.ID)
			return

		case "Указать ФИО":
			if !b.isAllowed(ctx, msg.From.ID) {
				b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
				return
			}
			b.replyText(msg.Chat.ID, "Введите имя и фамилию командой:\n/set_name Иван Иванов")
			return

		case "Мой ID":
			b.replyText(msg.Chat.ID, fmt.Sprintf("Ваш Telegram user_id: %d\nusername: @%s", msg.From.ID, msg.From.UserName))
			return

		case "Рус → Англ":
			_ = db.SetDirection(ctx, b.db, msg.From.ID, string(translate.RU_EN))
			b.replyText(msg.Chat.ID, "✅ Направление: Русский → Английский.\nОтправьте текст, документ, PDF или изображение.")
			return

		case "Англ → Рус":
			_ = db.SetDirection(ctx, b.db, msg.From.ID, string(translate.EN_RU))
			b.replyText(msg.Chat.ID, "✅ Направление: Английский → Русский.\nОтправьте текст, документ, PDF или изображение.")
			return

		case "Обычный режим":
			if err := b.setMode(ctx, msg.From.ID, translate.ModeDefault); err != nil {
				b.replyText(msg.Chat.ID, "Не смог сохранить режим: "+err.Error())
				return
			}
			b.replyText(msg.Chat.ID, "✅ Включён обычный режим перевода.\nТеперь отправьте текст или файл.")
			return

		case "Режим чертежа":
			if err := b.setMode(ctx, msg.From.ID, translate.ModeDrawing); err != nil {
				b.replyText(msg.Chat.ID, "Не смог сохранить режим: "+err.Error())
				return
			}
			b.replyText(msg.Chat.ID, "✅ Включён режим чертежа.\nБуду переводить текстовые требования, не изменяя размеры, коды и обозначения.\nТеперь отправьте чертёж, PDF или изображение.")
			return

		case "Статус":
			b.replyText(msg.Chat.ID, b.buildStatus(ctx, msg.From.ID))
			return

		case "Помощь":
			b.replyText(msg.Chat.ID, b.helpText(msg.From.ID))
			return
		}
		b.handleText(ctx, msg, msg.Text)
		return
	}

	b.replyText(msg.Chat.ID, "Пришлите текст, документ, PDF или изображение (.txt, .docx, .pdf, .png, .jpg, .jpeg, .tif, .tiff, .bmp).")
}

func (b *Bot) handleText(ctx context.Context, msg *tgbotapi.Message, text string) {
	d := b.getDirection(ctx, msg.From.ID)
	from, to := d.FromTo()

	tctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	mode := b.getMode(ctx, msg.From.ID)
	out, err := b.tr.TranslateWithMode(tctx, from, to, text, mode)
	if err != nil {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(d),
			SourceType:     "text",
			SourceFilename: "",
			InputChars:     len([]rune(text)),
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   err.Error(),
		})

		b.replyText(msg.Chat.ID, "Ошибка перевода: "+err.Error())
		return
	}

	_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
		TelegramUserID: msg.From.ID,
		Direction:      string(d),
		SourceType:     "text",
		SourceFilename: "",
		InputChars:     len([]rune(text)),
		OutputChars:    len([]rune(out)),
		IsSuccess:      true,
	})

	b.sendTextOrFile(msg.Chat.ID, out, "translation.txt")
}

func (b *Bot) handleDocument(ctx context.Context, msg *tgbotapi.Message) {
	doc := msg.Document
	chatID := msg.Chat.ID

	sourceType := strings.TrimPrefix(strings.ToLower(filepath.Ext(doc.FileName)), ".")
	if sourceType == "" {
		sourceType = "file"
	}

	tmpDir := os.TempDir()
	localPath := filepath.Join(tmpDir, fmt.Sprintf("tg_%s_%s", doc.FileUniqueID, doc.FileName))

	if err := b.downloadTelegramFile(doc.FileID, localPath); err != nil {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(b.getDirection(ctx, msg.From.ID)),
			SourceType:     sourceType,
			SourceFilename: doc.FileName,
			InputChars:     0,
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   err.Error(),
		})

		b.replyText(chatID, "Не смог скачать файл: "+err.Error())
		return
	}
	defer os.Remove(localPath)

	text, err := files.ExtractText(localPath)
	if err != nil {
		errMsg := err.Error()

		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(b.getDirection(ctx, msg.From.ID)),
			SourceType:     sourceType,
			SourceFilename: doc.FileName,
			InputChars:     0,
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   errMsg,
		})

		if strings.Contains(strings.ToLower(errMsg), "pdf: extracted empty text") {
			b.replyText(chatID, "Не удалось извлечь текст из PDF. Скорее всего, это сканированный PDF.")
			return
		}

		b.replyText(chatID, "Не смог извлечь текст из файла: "+errMsg)
		return
	}

	if strings.TrimSpace(text) == "" {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(b.getDirection(ctx, msg.From.ID)),
			SourceType:     sourceType,
			SourceFilename: doc.FileName,
			InputChars:     0,
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   "empty extracted text",
		})

		b.replyText(chatID, "Файл пустой или текст не извлёкся.")
		return
	}

	d := b.getDirection(ctx, msg.From.ID)
	from, to := d.FromTo()

	tctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	mode := b.getMode(ctx, msg.From.ID)

	out, err := b.tr.TranslateWithMode(tctx, from, to, text, mode)
	if err != nil {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(d),
			SourceType:     sourceType,
			SourceFilename: doc.FileName,
			InputChars:     len([]rune(text)),
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   err.Error(),
		})

		b.replyText(chatID, "Ошибка перевода: "+err.Error())
		return
	}

	_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
		TelegramUserID: msg.From.ID,
		Direction:      string(d),
		SourceType:     sourceType,
		SourceFilename: doc.FileName,
		InputChars:     len([]rune(text)),
		OutputChars:    len([]rune(out)),
		IsSuccess:      true,
	})

	outName := "translated_" + sanitizeName(doc.FileName) + ".txt"
	b.sendTextOrFile(chatID, out, outName)
}

func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	switch cb.Data {
	case "dir:ru_en":
		_ = db.SetDirection(ctx, b.db, cb.From.ID, string(translate.RU_EN))
		b.replyCallback(cb, "✅ RU → EN")
	case "dir:en_ru":
		_ = db.SetDirection(ctx, b.db, cb.From.ID, string(translate.EN_RU))
		b.replyCallback(cb, "✅ EN → RU")
	default:
		b.replyCallback(cb, "Не понял выбор.")
	}
}

func (b *Bot) handlePhoto(ctx context.Context, msg *tgbotapi.Message) {
	if len(msg.Photo) == 0 {
		b.replyText(msg.Chat.ID, "Не удалось получить изображение.")
		return
	}

	chatID := msg.Chat.ID
	photos := msg.Photo
	photo := photos[len(photos)-1]

	tmpDir := os.TempDir()
	localPath := filepath.Join(tmpDir, fmt.Sprintf("tg_photo_%s.jpg", photo.FileUniqueID))

	if err := b.downloadTelegramFile(photo.FileID, localPath); err != nil {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(b.getDirection(ctx, msg.From.ID)),
			SourceType:     "photo",
			SourceFilename: "telegram_photo.jpg",
			InputChars:     0,
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   err.Error(),
		})

		b.replyText(chatID, "Не смог скачать изображение: "+err.Error())
		return
	}
	defer os.Remove(localPath)

	text, err := files.ExtractText(localPath)
	if err != nil {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(b.getDirection(ctx, msg.From.ID)),
			SourceType:     "photo",
			SourceFilename: "telegram_photo.jpg",
			InputChars:     0,
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   err.Error(),
		})

		b.replyText(chatID, "Не смог извлечь текст из изображения: "+err.Error())
		return
	}

	if strings.TrimSpace(text) == "" {
		b.replyText(chatID, "Не удалось распознать текст на изображении.")
		return
	}

	d := b.getDirection(ctx, msg.From.ID)
	mode := b.getMode(ctx, msg.From.ID)
	from, to := d.FromTo()

	tctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	out, err := b.tr.TranslateWithMode(tctx, from, to, text, mode)
	if err != nil {
		_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
			TelegramUserID: msg.From.ID,
			Direction:      string(d),
			SourceType:     "photo",
			SourceFilename: "telegram_photo.jpg",
			InputChars:     len([]rune(text)),
			OutputChars:    0,
			IsSuccess:      false,
			ErrorMessage:   err.Error(),
		})

		b.replyText(chatID, "Ошибка перевода: "+err.Error())
		return
	}

	_ = db.InsertTranslationLog(ctx, b.db, db.TranslationLog{
		TelegramUserID: msg.From.ID,
		Direction:      string(d),
		SourceType:     "photo",
		SourceFilename: "telegram_photo.jpg",
		InputChars:     len([]rune(text)),
		OutputChars:    len([]rune(out)),
		IsSuccess:      true,
	})

	b.sendTextOrFile(chatID, out, "translated_photo.txt")
}
