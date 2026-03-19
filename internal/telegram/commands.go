package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/pha06ntom/tg-translator-bot/internal/db"
	"github.com/pha06ntom/tg-translator-bot/internal/translate"
)

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}

		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Здравствуйте.\nДля работы с ботом необходимо указать имя и фамилию.\n\nВведите:\n/set_name Иван Иванов")
			return
		}

		b.replyText(msg.Chat.ID, "Здравствуйте. Я помогу перевести текст, документ, PDF или изображение.\n\nВыберите направление и режим кнопками ниже, затем отправьте файл или текст.")
		b.sendMainMenu(msg.Chat.ID)
		return

	case "help":
		b.replyText(msg.Chat.ID, b.helpText(msg.From.ID))
		if b.isAllowed(ctx, msg.From.ID) && b.hasRealName(ctx, msg.From.ID) {
			b.sendMainMenu(msg.Chat.ID)
		}
		return

	case "status":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}
		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Сначала укажите имя и фамилию:\n/set_name Иван Иванов")
			return
		}
		b.replyText(msg.Chat.ID, b.buildStatus(ctx, msg.From.ID))
		return

	case "set_name":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}

		args := strings.TrimSpace(msg.CommandArguments())
		if args == "" {
			b.replyText(msg.Chat.ID, "Введите имя и фамилию:\n/set_name Иван Иванов")
			return
		}

		if len(strings.Fields(args)) < 2 {
			b.replyText(msg.Chat.ID, "Пожалуйста, укажите имя и фамилию.\nПример:\n/set_name Иван Иванов")
			return
		}

		if err := db.SetRealName(ctx, b.db, msg.From.ID, args); err != nil {
			b.replyText(msg.Chat.ID, "Ошибка сохранения: "+err.Error())
			return
		}

		b.replyText(msg.Chat.ID, "✅ Имя и фамилия сохранены. Теперь можно пользоваться ботом.")
		b.sendMainMenu(msg.Chat.ID)
		return

	case "dir":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}
		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Сначала укажите имя и фамилию:\n/set_name Иван Иванов")
			return
		}
		b.sendDirKeyboard(msg.Chat.ID, "Направление перевода:")
		return

	case "ruen":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}
		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Сначала укажите имя и фамилию:\n/set_name Иван Иванов")
			return
		}
		_ = db.SetDirection(ctx, b.db, msg.From.ID, string(translate.RU_EN))
		b.replyText(msg.Chat.ID, "✅ Направление: RU → EN")
		return

	case "enru":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}
		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Сначала укажите имя и фамилию:\n/set_name Иван Иванов")
			return
		}
		_ = db.SetDirection(ctx, b.db, msg.From.ID, string(translate.EN_RU))
		b.replyText(msg.Chat.ID, "✅ Направление: EN → RU")
		return

	case "drawing":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}
		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Сначала укажите имя и фамилию:\n/set_name Иван Иванов")
			return
		}
		if err := b.setMode(ctx, msg.From.ID, translate.ModeDrawing); err != nil {
			b.replyText(msg.Chat.ID, "Не смог сохранить режим: "+err.Error())
			return
		}
		b.replyText(msg.Chat.ID, "✅ Режим: чертёж. Буду переводить только текстовые требования, не трогая размеры и обозначения.")
		return

	case "normal":
		if !b.isAllowed(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Нет доступа.")
			return
		}
		if !b.hasRealName(ctx, msg.From.ID) {
			b.replyText(msg.Chat.ID, "Сначала укажите имя и фамилию:\n/set_name Иван Иванов")
			return
		}
		if err := b.setMode(ctx, msg.From.ID, translate.ModeDefault); err != nil {
			b.replyText(msg.Chat.ID, "Не смог сохранить режим: "+err.Error())
			return
		}
		b.replyText(msg.Chat.ID, "✅ Режим: обычный перевод.")
		return

	case "whoami":
		b.replyText(msg.Chat.ID, fmt.Sprintf("Ваш Telegram user_id: %d\nusername: @%s", msg.From.ID, msg.From.UserName))
		return

	case "allow_id":
		if !b.isAdmin(msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Только админ может выдавать доступ.")
			return
		}
		args := strings.TrimSpace(msg.CommandArguments())
		if args == "" {
			b.replyText(msg.Chat.ID, "Пример: /allow_id 123456789")
			return
		}
		id, err := strconv.ParseInt(args, 10, 64)
		if err != nil || id <= 0 {
			b.replyText(msg.Chat.ID, "Некорректный id. Пример: /allow_id 123456789")
			return
		}
		if err := db.SetAllowedByID(ctx, b.db, id, true); err != nil {
			b.replyText(msg.Chat.ID, "Ошибка: "+err.Error())
			return
		}
		b.replyText(msg.Chat.ID, fmt.Sprintf("✅ Доступ выдан user_id=%d", id))
		return

	case "deny_id":
		if !b.isAdmin(msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Только админ может забирать доступ.")
			return
		}
		args := strings.TrimSpace(msg.CommandArguments())
		if args == "" {
			b.replyText(msg.Chat.ID, "Пример: /deny_id 123456789")
			return
		}
		id, err := strconv.ParseInt(args, 10, 64)
		if err != nil || id <= 0 {
			b.replyText(msg.Chat.ID, "Некорректный id. Пример: /deny_id 123456789")
			return
		}
		if err := db.SetAllowedByID(ctx, b.db, id, false); err != nil {
			b.replyText(msg.Chat.ID, "Ошибка: "+err.Error())
			return
		}
		b.replyText(msg.Chat.ID, fmt.Sprintf("🚫 Доступ снят user_id=%d", id))
		return

	case "allow":
		if !b.isAdmin(msg.From.ID) {
			b.replyText(msg.Chat.ID, "⛔️ Только админ.")
			return
		}
		arg := strings.TrimSpace(msg.CommandArguments())
		arg = strings.TrimPrefix(arg, "@")
		if arg == "" {
			b.replyText(msg.Chat.ID, "Пример: /allow @username")
			return
		}
		id, err := db.SetAllowedByUsername(ctx, b.db, arg, true)
		if err != nil {
			b.replyText(msg.Chat.ID, "Не смог выдать доступ по username. Скорее всего пользователь ещё ни разу не писал боту (/start). Ошибка: "+err.Error())
			return
		}
		b.replyText(msg.Chat.ID, fmt.Sprintf("✅ Доступ выдан @%s (user_id=%d)", arg, id))
		return

	default:
		b.replyText(msg.Chat.ID, "Неизвестная команда. /help")
		return
	}
}

func (b *Bot) helpText(userID int64) string {
	base := `Я бот-переводчик.

Что я умею:
- переводить текст из сообщений
- переводить файлы .txt, .docx, .pdf
- распознавать текст на изображениях и сканах через OCR
- переводить текстовые требования с чертежей

Как начать работу:
1. Нажмите /start
2. Если бот попросит, укажите имя и фамилию командой /set_name Иван Иванов
3. Выберите направление перевода
4. Выберите режим
5. Отправьте текст, документ, PDF или изображение

Основные действия доступны кнопками внизу экрана.

Основные команды:
  /start       — старт
  /help        — справка
  /status      — текущие настройки
  /set_name    — указать имя и фамилию
  /whoami      — показать ваш user_id
  /ruen        — Рус → Англ
  /enru        — Англ → Рус
  /drawing     — режим перевода чертежей
  /normal      — обычный режим`

	if b.isAdmin(userID) {
		base += `

Админ-команды:
  /allow_id <id>  — выдать доступ по user_id
  /deny_id <id>   — снять доступ по user_id
  /allow @user    — выдать доступ по username`
	}

	return base
}

func (b *Bot) buildStatus(ctx context.Context, userID int64) string {
	dir := b.getDirection(ctx, userID)
	mode := b.getMode(ctx, userID)

	dirText := "Рус → Англ"
	if dir == translate.EN_RU {
		dirText = "Англ → Рус"
	}

	modeText := "Обычный"
	if mode == translate.ModeDrawing {
		modeText = "Чертёж"
	}

	return fmt.Sprintf(
		"Текущие настройки:\nНаправление: %s\nРежим: %s\n\nМожете отправить текст, документ, PDF или изображение.",
		dirText,
		modeText,
	)
}

func (b *Bot) getDirection(ctx context.Context, userID int64) translate.Direction {
	dir, err := db.GetDirection(ctx, b.db, userID)
	if err != nil {
		return translate.RU_EN
	}

	switch dir {
	case string(translate.EN_RU):
		return translate.EN_RU
	default:
		return translate.RU_EN
	}
}

func (b *Bot) getMode(ctx context.Context, userID int64) translate.Mode {
	mode, err := db.GetMode(ctx, b.db, userID)
	if err != nil {
		return translate.ModeDefault
	}

	switch mode {
	case string(translate.ModeDrawing):
		return translate.ModeDrawing
	default:
		return translate.ModeDefault
	}
}

func (b *Bot) setMode(ctx context.Context, userID int64, mode translate.Mode) error {
	return db.SetMode(ctx, b.db, userID, string(mode))
}
