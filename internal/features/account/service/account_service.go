package accservice

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/account"
	accrepository "github.com/codepnw/go-starter-kit/internal/features/account/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
)

type AccountService interface {
	CreateAccount(ctx context.Context, input *account.Account) error
	GetAllAccounts(ctx context.Context, ownerID string) ([]*account.Account, error)
	GetAccountByID(ctx context.Context, ownerID string, accountID int64) (*account.Account, error)
	DepositMoney(ctx context.Context, accountID, amount int64) (*account.Account, error)
	WithdrawMoney(ctx context.Context, accountID, amount int64) (*account.Account, error)
}

type accountService struct {
	tx   database.TxManager
	repo accrepository.AccountRepository
}

func NewAccountSevice(tx database.TxManager, repo accrepository.AccountRepository) AccountService {
	return &accountService{
		tx:   tx,
		repo: repo,
	}
}

// CreateAccount implements AccountService.
func (s *accountService) CreateAccount(ctx context.Context, input *account.Account) error {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	return s.repo.InsertAccount(ctx, input)
}

// GetAllAccounts implements AccountService.
func (s *accountService) GetAllAccounts(ctx context.Context, ownerID string) ([]*account.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	return s.repo.FindAccountsByOwner(ctx, ownerID)
}

// GetAccountByID implements AccountService.
func (s *accountService) GetAccountByID(ctx context.Context, userID string, accountID int64) (*account.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	data, err := s.repo.FindAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if userID != data.OwnerID {
		return nil, errs.ErrForbidden
	}
	return data, nil
}

// DepositMoney implements AccountService.
func (s *accountService) DepositMoney(ctx context.Context, accountID int64, amount int64) (*account.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if amount <= 0 {
		return nil, errs.ErrAmountGeaterThanZero
	}

	var acc *account.Account

	err := s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Insert Entry
		if err := s.repo.InsertEntryTx(ctx, tx, accountID, amount); err != nil {
			return fmt.Errorf("insert entry failed: %w", err)
		}

		// Update Account Balance
		resp, err := s.repo.UpdateAccountBalanceTx(ctx, tx, accountID, amount)
		if err != nil {
			return fmt.Errorf("update balance failed: %w", err)
		}
		acc = resp

		return nil
	})
	return acc, err
}

func (s *accountService) WithdrawMoney(ctx context.Context, accountID, amount int64) (*account.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if amount <= 0 {
		return nil, errs.ErrAmountGeaterThanZero
	}

	var acc *account.Account

	err := s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Insert Entry
		if err := s.repo.InsertEntryTx(ctx, tx, accountID, -amount); err != nil {
			return fmt.Errorf("insert entry failed: %w", err)
		}

		resp, err := s.repo.UpdateAccountBalanceTx(ctx, tx, accountID, -amount)
		if err != nil {
			return err
		}
		acc = resp

		return nil
	})
	return acc, err
}
