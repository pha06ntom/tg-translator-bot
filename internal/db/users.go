package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserUpsert struct {
	TelegramUserID int64
	Username       string
	DisplayName    string
}

type User struct {
	TelegramUserID int64
	Username       string
	DisplayName    string
	RealName       string
}

func UpsertUser(ctx context.Context, pool *pgxpool.Pool, u UserUpsert) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO users (telegram_user_id, username, display_name)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''))
		ON CONFLICT (telegram_user_id)
		DO UPDATE SET
		  username = COALESCE(NULLIF(EXCLUDED.username,''), users.username),
		  display_name = COALESCE(NULLIF(EXCLUDED.display_name,''), users.display_name),
		  updated_at = now()
	`, u.TelegramUserID, u.Username, u.DisplayName)
	return err
}

func IsAllowed(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64) (bool, error) {
	var allowed bool
	err := pool.QueryRow(ctx, "SELECT is_allowed FROM users WHERE telegram_user_id=$1", telegramUserID).Scan(&allowed)
	if err != nil {
		return false, nil
	}
	return allowed, nil
}

func SetAllowedByID(ctx context.Context, pool *pgxpool.Pool, telegramUserID int64, allowed bool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO users (telegram_user_id, is_allowed)
		VALUES ($1, $2)
		ON CONFLICT (telegram_user_id)
		DO UPDATE SET 
			is_allowed = EXCLUDED.is_allowed,
			updated_at = now()
	`, telegramUserID, allowed)

	return err
}

func SetAllowedByUsername(ctx context.Context, pool *pgxpool.Pool, username string, allowed bool) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx, `
		UPDATE users
		SET is_allowed=$2, updated_at=now()
		WHERE lower(username)=lower($1)
		RETURNING telegram_user_id
	`, username, allowed).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func SetRealName(ctx context.Context, pool *pgxpool.Pool, userID int64, name string) error {
	_, err := pool.Exec(ctx, `
		UPDATE users
		SET real_name = $1,
		    updated_at = now()
		WHERE telegram_user_id = $2
	`, name, userID)

	return err
}

func GetUser(ctx context.Context, pool *pgxpool.Pool, userID int64) (User, error) {
	var u User

	err := pool.QueryRow(ctx, `
		SELECT telegram_user_id, COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(real_name, '')
		FROM users
		WHERE telegram_user_id = $1
	`, userID).Scan(&u.TelegramUserID, &u.Username, &u.DisplayName, &u.RealName)

	return u, err
}
