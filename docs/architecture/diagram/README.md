# architecture
![Architecture diagram](gophermart_architecture.png)
# database schema
![database schema](gophermart_db.png)

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  password BYTEA NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    userId UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status order_status_enum NOT NULL,
    accrual DECIMAL(12, 2) NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_at_epoch BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS order_transactions (
    orderId BIGINT NOT NULL,
    userId UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_debit BOOLEAN NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    created_at_epoch BIGINT NOT NULL
);
```
