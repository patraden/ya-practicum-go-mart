-- +goose Up
-- +goose StatementBegin

-- Create the ENUM type for order statuses
CREATE TYPE order_status_enum AS ENUM (
  'NEW', 
  'REGISTERED', 
  'PROCESSING', 
  'INVALID', 
  'PROCESSED'
);

CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    userId UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status order_status_enum NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS order_transactions (
    orderId BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    userId UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_debit BOOLEAN NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(userId);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_transactions_user_id ON order_transactions(userId);
CREATE UNIQUE INDEX uniq_order_id_is_debit_true ON order_transactions (orderId) WHERE is_debit;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders cascade;
DROP TABLE IF EXISTS order_transactions cascade;
DROP TYPE IF EXISTS order_status_enum;
-- +goose StatementEnd
