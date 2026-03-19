BEGIN;

CREATE TABLE IF NOT EXISTS translation_logs (
  id BIGSERIAL PRIMARY KEY,
  telegram_user_id BIGINT NOT NULL REFERENCES users(telegram_user_id) ON DELETE CASCADE,
  direction TEXT NOT NULL,
  source_type TEXT NOT NULL,
  source_filename TEXT NULL,
  input_chars INTEGER NOT NULL DEFAULT 0,
  output_chars INTEGER NOT NULL DEFAULT 0,
  is_success BOOLEAN NOT NULL DEFAULT FALSE,
  error_message TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_translation_logs_user_id ON translation_logs (telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_translation_logs_created_at ON translation_logs (created_at);

COMMIT;