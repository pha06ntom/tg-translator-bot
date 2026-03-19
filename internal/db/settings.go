package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSettings struct {
	Direction string
	Mode      string
}

func GetSettings(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64) (UserSettings, error) {
	var s UserSettings

	err := pool.QueryRow(ctx, `
		SELECT direction, mode
		FROM user_settings
		WHERE telegram_user_id = $1
	`, telegramUserID).Scan(&s.Direction, &s.Mode)

	if err != nil {
		return UserSettings{
			Direction: "ru_en",
			Mode:      "default",
		}, nil
	}

	if s.Direction == "" {
		s.Direction = "ru_en"
	}
	if s.Mode == "" {
		s.Mode = "default"
	}

	return s, nil
}

func GetDirection(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64) (string, error) {
	s, err := GetSettings(ctx, pool, telegramUserID)
	if err != nil {
		return "ru_en", err
	}
	return s.Direction, nil
}

func SetDirection(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64, dir string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO user_settings (telegram_user_id, direction, mode)
		VALUES ($1, $2, 'default')
		ON CONFLICT (telegram_user_id)
		DO UPDATE SET
			direction = EXCLUDED.direction,
			updated_at = now()
	`, telegramUserID, dir)

	return err
}

func GetMode(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64) (string, error) {
	s, err := GetSettings(ctx, pool, telegramUserID)
	if err != nil {
		return "default", err
	}
	return s.Mode, nil
}

func SetMode(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64, mode string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO user_settings (telegram_user_id, direction, mode)
		VALUES ($1, 'ru_en', $2)
		ON CONFLICT (telegram_user_id)
		DO UPDATE SET
			mode = EXCLUDED.mode,
			updated_at = now()
	`, telegramUserID, mode)

	return err
}
