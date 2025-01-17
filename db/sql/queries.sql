-- name: CreateUser :one
INSERT INTO users (id, username, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (username) DO UPDATE
SET id = users.id,
    username = users.username,
    password = users.password,
    created_at = users.created_at,
    updated_at = users.updated_at
RETURNING id, username, password, created_at, updated_at;

-- name: GetUser :one
SELECT id, username, password, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserBalances :one
SELECT 
  userID AS userid, 
  SUM(CASE WHEN is_debit THEN amount ELSE -amount END)::DECIMAL(12, 2) AS balance, 
  SUM(CASE WHEN is_debit THEN 0 ELSE amount END)::DECIMAL(12, 2) AS withdrawn
FROM order_transactions
WHERE userID = $1
GROUP BY userID;

-- name: CreateOrder :one
INSERT INTO orders (id, userid, created_at, status, accrual, updated_at, created_at_epoch)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE 
SET id = orders.id,
    userid = orders.userid,
    created_at = orders.created_at,
    status = orders.status,
    accrual = orders.accrual,
    updated_at = orders.updated_at,
    created_at_epoch = orders.created_at_epoch
RETURNING id, userid, created_at, status, accrual, updated_at, created_at_epoch;

-- name: GetOrders :many
SELECT id, userid, created_at, status, accrual, updated_at, created_at_epoch
FROM orders
WHERE userID = $1
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $1, accrual = $2, updated_at = now()
WHERE id = $3;

-- name: CreateOrderAccrual :exec
INSERT INTO order_transactions 
(orderId, userId, is_debit, amount, created_at, created_at_epoch)
VALUES ($1, $2, TRUE, $3, $4, $5);

-- name: LockUserTransactions :exec
SELECT pg_advisory_xact_lock($1);

-- name: CreateUserWithdrawal :one
WITH order_balance AS (
  SELECT
   COALESCE(SUM(CASE WHEN trx.is_debit THEN trx.amount ELSE -trx.amount END), 0)::DECIMAL(12, 2) AS balance
  FROM order_transactions trx
  WHERE trx.userId = $2
)
INSERT INTO order_transactions 
(orderId, userId, is_debit, amount, created_at, created_at_epoch)
SELECT $1, $2, FALSE, $3, $4, $5
FROM order_balance
WHERE order_balance.balance - $3 >= 0
RETURNING orderId;

-- name: GetUserWithdrawals :many
SELECT orderId, amount, created_at, created_at_epoch
FROM order_transactions
WHERE userId = $1 AND NOT is_debit
ORDER BY created_at DESC;