package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pha06ntom/tg-translator-bot/internal/db"
	"github.com/pha06ntom/tg-translator-bot/internal/translate"
)

type Translator interface {
	Translate(ctx context.Context, fromLang, toLang, text string) (string, error)
	TranslateWithMode(ctx context.Context, fromLang, toLang, text string, mode translate.Mode) (string, error)
}

type Bot struct {
	api    *tgbotapi.BotAPI
	tr     Translator
	db     *pgxpool.Pool
	admins map[int64]struct{}
}

func New(token string, pool *pgxpool.Pool, adminIDs []int64, tr Translator) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	admins := make(map[int64]struct{}, len(adminIDs))
	for _, id := range adminIDs {
		admins[id] = struct{}{}
	}

	return &Bot{
		api:    api,
		tr:     tr,
		db:     pool,
		admins: admins,
	}, nil
}

func (b *Bot) isAdmin(userID int64) bool {
	_, ok := b.admins[userID]
	return ok
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case upd := <-updates:
			b.handleUpdate(ctx, upd)
		}
	}
}

func (b *Bot) isAllowed(ctx context.Context, userID int64) bool {
	if b.isAdmin(userID) {
		return true
	}

	allowed, err := db.IsAllowed(ctx, b.db, userID)
	if err == nil && allowed {
		return true
	}

	// Автоматически выдаём доступ пользователю, который уже написал боту
	if err := db.SetAllowedByID(ctx, b.db, userID, true); err != nil {
		return false
	}

	return true
}

func (b *Bot) hasRealName(ctx context.Context, userID int64) bool {
	u, err := db.GetUser(ctx, b.db, userID)
	if err != nil {
		return false
	}
	return strings.TrimSpace(u.RealName) != ""
}
