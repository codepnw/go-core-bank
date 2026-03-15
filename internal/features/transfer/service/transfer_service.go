package transferservice

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/errs"
	accrepository "github.com/codepnw/go-starter-kit/internal/features/account/repository"
	"github.com/codepnw/go-starter-kit/internal/features/transfer"
	transferrepository "github.com/codepnw/go-starter-kit/internal/features/transfer/repository"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
)

type TransferService interface {
	TransferMoney(ctx context.Context, userID string, input *transfer.Transfer) (*TransferMoneyResponse, error)
}

type transferService struct {
	tx            database.TxManager
	transferRepo  transferrepository.TransferRepository
	transferRedis transferrepository.TransferRedisRepository
	accountRepo   accrepository.AccountRepository
	userRepo      userrepository.UserRepository
}

type TransferServiceDeps struct {
	Tx            database.TxManager
	TransferRepo  transferrepository.TransferRepository
	TransferRedis transferrepository.TransferRedisRepository
	AccountRepo   accrepository.AccountRepository
	UserRepo      userrepository.UserRepository
}

func NewTransferService(deps *TransferServiceDeps) TransferService {
	return &transferService{
		tx:            deps.Tx,
		transferRepo:  deps.TransferRepo,
		transferRedis: deps.TransferRedis,
		accountRepo:   deps.AccountRepo,
		userRepo:      deps.UserRepo,
	}
}

type TransferMoneyResponse struct {
	TransferID int64    `json:"transfer_id"`
	Amount     string   `json:"amount"`
	Sender     userInfo `json:"sender"`
	Receiver   userInfo `json:"receiver"`
}

type userInfo struct {
	AccountID int64  `json:"account_id"`
	FullName  string `json:"full_name"`
}

// TransferMoney implements TransferService.
func (s *transferService) TransferMoney(ctx context.Context, userID string, input *transfer.Transfer) (resp *TransferMoneyResponse, err error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if input.FromAccountID == input.ToAccountID {
		return nil, errs.ErrTransferSameAccount
	}
	if input.Amount <= 0 {
		return nil, errs.ErrAmountGeaterThanZero
	}

	// Redis: Check Idempotency Key
	isNewKey, err := s.transferRedis.CheckIdempotencyKey(ctx, input.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("check redis idem key failed: %w", err)
	}
	if !isNewKey {
		return nil, errs.ErrDuplicateTransfer
	}

	// If Error Delete Redis-Key
	defer func() {
		if err != nil {
			_ = s.transferRedis.DeleteIdempotencyKey(ctx, input.IdempotencyKey)
		}
	}()
	
	// Validate Transfer Data Response
	resp, err = s.validateTransferData(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	// Start Transaction
	err = s.tx.WithTx(ctx, func(tx *sql.Tx) error {
		// Create Transfer
		if err := s.transferRepo.InsertTransferTx(ctx, tx, input); err != nil {
			return fmt.Errorf("insert transfer failed: %w", err)
		}
		resp.TransferID = input.ID
		resp.Amount = fmt.Sprintf("%.2f", float64(input.Amount)/100)

		// Create Entry From Account
		if err := s.accountRepo.InsertEntryTx(ctx, tx, input.FromAccountID, -input.Amount); err != nil {
			return fmt.Errorf("insert from-account entry failed: %w", err)
		}

		// Create Entry To Account
		if err := s.accountRepo.InsertEntryTx(ctx, tx, input.ToAccountID, input.Amount); err != nil {
			return fmt.Errorf("insert to-account entry failed: %w", err)
		}

		// Update Balance + Row Lock (Deadlock)
		if input.FromAccountID < input.ToAccountID {
			// Update From Account Balance
			fromAccResp, err := s.accountRepo.UpdateAccountBalanceTx(ctx, tx, input.FromAccountID, -input.Amount)
			if err != nil {
				return fmt.Errorf("update from-account failed: %w", err)
			}
			resp.Sender.AccountID = fromAccResp.ID

			// Update To Account Balance
			toAccResp, err := s.accountRepo.UpdateAccountBalanceTx(ctx, tx, input.ToAccountID, input.Amount)
			if err != nil {
				return fmt.Errorf("update to-account failed: %w", err)
			}
			resp.Receiver.AccountID = toAccResp.ID
		} else {
			// Update To Account Balance
			toAccResp, err := s.accountRepo.UpdateAccountBalanceTx(ctx, tx, input.ToAccountID, input.Amount)
			if err != nil {
				return fmt.Errorf("update to-account failed: %w", err)
			}
			resp.Receiver.AccountID = toAccResp.ID

			// Update From Account Balance
			fromAccResp, err := s.accountRepo.UpdateAccountBalanceTx(ctx, tx, input.FromAccountID, -input.Amount)
			if err != nil {
				return fmt.Errorf("update from-account failed: %w", err)
			}
			resp.Sender.AccountID = fromAccResp.ID
		}

		return nil
	})
	if err != nil {
		// if error defer delete redis-key
		return nil, err
	}
	return resp, err
}

func (s *transferService) validateTransferData(ctx context.Context, userID string, input *transfer.Transfer) (*TransferMoneyResponse, error) {
	// Data Response
	resp := &TransferMoneyResponse{}

	// Get From Account Data
	fromAccData, err := s.accountRepo.FindAccountByID(ctx, input.FromAccountID)
	if err != nil {
		return nil, fmt.Errorf("get from-account failed: %w", err)
	}
	if userID != fromAccData.OwnerID {
		return nil, errs.ErrForbidden
	}

	// Get To Account Data
	toAccData, err := s.accountRepo.FindAccountByID(ctx, input.ToAccountID)
	if err != nil {
		return nil, fmt.Errorf("get to-account failed: %w", err)
	}

	// Get From Account User
	fromUserData, err := s.userRepo.FindUserByID(ctx, fromAccData.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("get from-user failed: %w", err)
	}
	resp.Sender.FullName = formatAccountFullName(fromUserData.FirstName, fromUserData.LastName)

	// Get To Account User
	toUserData, err := s.userRepo.FindUserByID(ctx, toAccData.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("get to-user failed: %w", err)
	}
	resp.Receiver.FullName = formatAccountFullName(toUserData.FirstName, toUserData.LastName)
	
	return resp, nil
}

func formatAccountFullName(firstName, lastName string) string {
	return fmt.Sprintf("%s %s", firstName, lastName)
}
