-- +goose Up
-- +goose StatementBegin
ALTER TABLE books ADD COLUMN deleted_at TIMESTAMP DEFAULT NULL;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP DEFAULT NULL;

-- Add index for performance when filtering out deleted records
CREATE INDEX idx_books_deleted_at ON books(deleted_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_books_deleted_at;
DROP INDEX IF EXISTS idx_users_deleted_at;

ALTER TABLE books DROP COLUMN deleted_at;
ALTER TABLE users DROP COLUMN deleted_at;
-- +goose StatementEnd