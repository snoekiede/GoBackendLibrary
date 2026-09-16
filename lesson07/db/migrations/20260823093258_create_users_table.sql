-- +goose Up
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
                       updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);


-- +goose Down
DROP TABLE IF EXISTS users;

