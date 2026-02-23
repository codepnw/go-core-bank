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

	// Transaction
	UpdateDepositBalance(ctx context.Context, tx *sql.Tx, accountID, balance int64) (*account.Account, error)
	UpdateWithdrawBalance(ctx context.Context, tx *sql.Tx, accountID, balance int64) (*account.Account, error)
	// Entry Table
	InsertEntry(ctx context.Context, tx *sql.Tx, accountID, amount int64) error
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

// UpdateDepositBalance implements AccountRepository.
func (r *accountRepository) UpdateDepositBalance(ctx context.Context, tx *sql.Tx, accountID int64, balance int64) (*account.Account, error) {
	var acc account.Account
	query := `
		UPDATE accounts SET balance = balance + $1
		WHERE id = $2 RETURNING id, title, balance
	`
	err := tx.QueryRowContext(ctx, query, balance, accountID).Scan(
		&acc.ID,
		&acc.Title,
		&acc.Balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

// UpdateWithdrawBalance implements AccountRepository.
func (r *accountRepository) UpdateWithdrawBalance(ctx context.Context, tx *sql.Tx, accountID int64, balance int64) (*account.Account, error) {
	var acc account.Account
	query := `
		UPDATE accounts SET balance = balance - $1
		WHERE id = $2 AND balance >= $1
		RETURNING id, title, balance
	`
	err := tx.QueryRowContext(ctx, query, balance, accountID).Scan(
		&acc.ID,
		&acc.Title,
		&acc.Balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

// InsertEntry implements AccountRepository.
func (r *accountRepository) InsertEntry(ctx context.Context, tx *sql.Tx, accountID int64, amount int64) error {
	query := `INSERT INTO entries (account_id, amount) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, query, accountID, amount)
	if err != nil {
		return err
	}
	return nil
}
