-- +goose Up
-- +goose StatementBegin
ALTER TABLE books ADD COLUMN available BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE books DROP COLUMN IF EXISTS available;
-- +goose StatementEnd