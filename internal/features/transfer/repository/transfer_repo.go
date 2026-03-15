package transferrepository

import (
	"context"
	"database/sql"

	"github.com/codepnw/go-starter-kit/internal/features/transfer"
)

//go:generate mockgen -source=transfer_repo.go -destination=transfer_repo_mock.go -package=transferrepository
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
		INSERT INTO transfers (from_account_id, to_account_id, amount, idempotency_key)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at
	`
	err := tx.QueryRowContext(
		ctx, 
		query, 
		input.FromAccountID, 
		input.ToAccountID, 
		input.Amount,
		input.IdempotencyKey,
	).Scan(
		&input.ID,
		&input.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
