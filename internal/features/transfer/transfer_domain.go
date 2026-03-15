package transfer

import "time"

type Transfer struct {
	ID             int64     `db:"id"`
	FromAccountID  int64     `db:"from_account_id"`
	ToAccountID    int64     `db:"to_account_id"`
	Amount         int64     `db:"amount"`
	IdempotencyKey string    `db:"idempotency_key"`
	CreatedAt      time.Time `db:"created_at"`
}
