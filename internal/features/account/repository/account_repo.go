package accrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/account"
)

type AccountRepository interface {
	InsertAccount(ctx context.Context, acc *account.Account) error
	FindAllAccounts(ctx context.Context, ownerID string) ([]*account.Account, error)
	FindAccountByID(ctx context.Context, ownerID string, accountID int64) (*account.Account, error)
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

// InsertAccount implements AccountRepository.
func (r *accountRepository) InsertAccount(ctx context.Context, acc *account.Account) error {
	query := `
		INSERT INTO accounts (owner_id, title, balance)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, acc.OwnerID, acc.Title, acc.Balance).Scan(
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// FindAllAccounts implements AccountRepository.
func (r *accountRepository) FindAllAccounts(ctx context.Context, ownerID string) ([]*account.Account, error) {
	var accounts []*account.Account
	query := `
		SELECT id, title, balance FROM accounts
		WHERE owner_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}

	for rows.Next() {
		var acc account.Account

		if err := rows.Scan(
			&acc.ID,
			&acc.Title,
			&acc.Balance,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, &acc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

// FindAccountByID implements AccountRepository.
func (r *accountRepository) FindAccountByID(ctx context.Context, ownerID string, accountID int64) (*account.Account, error) {
	var acc account.Account
	query := `
		SELECT title, balance, updated_at
		FROM accounts WHERE id = $1 AND owner_id = $2
	`
	err := r.db.QueryRowContext(ctx, query, accountID, ownerID).Scan(
		&acc.Title,
		&acc.Balance,
		&acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}
