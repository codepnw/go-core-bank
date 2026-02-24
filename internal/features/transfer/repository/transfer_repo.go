package transferrepository

import (
	"context"
	"database/sql"

	"github.com/codepnw/go-starter-kit/internal/features/transfer"
)

type TransferRepository interface {
	InsertTransferTx(ctx context.Context, tx *sql.Tx, input *transfer.Transfer) error
}

type transferRepository struct {
	db *sql.DB
}

func NewTransferRepository(db *sql.DB) TransferRepository {
	return &transferRepository{db: db}
}

// InsertTransferTx implements TransferRepository.
func (r *transferRepository) InsertTransferTx(ctx context.Context, tx *sql.Tx, input *transfer.Transfer) error {
	query := `
		INSERT INTO transfers (from_account_id, to_account_id, amount)
		VALUES ($1, $2, $3) RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query, input.FromAccountID, input.ToAccountID, input.Amount).Scan(
		&input.ID,
		&input.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
