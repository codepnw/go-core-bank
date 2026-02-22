package account

import (
	"time"
)

type Account struct {
	ID        int64     `db:"id"`
	OwnerID   string    `db:"owner_id"`
	Title     string    `db:"title"`
	Balance   int64     `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
