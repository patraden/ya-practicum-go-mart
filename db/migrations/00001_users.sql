-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  password BYTEA NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE user_balances (
  userID UUID PRIMARY KEY,
  balance DECIMAL(12, 2) NOT NULL DEFAULT 0,
  withdrawn DECIMAL(12, 2) NOT NULL DEFAULT 0,
  updated_at NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_locks (
  user_id UUID PRIMARY KEY,
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_locks;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
