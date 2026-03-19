package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TranslationLog struct {
	TelegramUserID int64
	Direction      string
	SourceType     string
	SourceFilename string
	InputChars     int
	OutputChars    int
	IsSuccess      bool
	ErrorMessage   string
}

func InsertTranslationLog(ctx context.Context, pool *pgxpool.Pool, l TranslationLog) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO translation_logs (
			telegram_user_id,
			direction,
			source_type,
			source_filename,
			input_chars,
			output_chars,
			is_success,
			error_message
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, NULLIF($8, ''))
	`,
		l.TelegramUserID,
		l.Direction,
		l.SourceType,
		l.SourceFilename,
		l.InputChars,
		l.OutputChars,
		l.IsSuccess,
		l.ErrorMessage,
	)

	return err
}
