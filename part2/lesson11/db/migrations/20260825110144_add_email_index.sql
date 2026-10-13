-- +goose Up
ALTER TABLE users DROP CONSTRAINT users_email_key;

CREATE UNIQUE INDEX idx_unique_active_user_email
    ON users(email)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_unique_active_user_email;

ALTER TABLE users
    ADD CONSTRAINT users_email_key UNIQUE (email);
