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
	tx           database.TxManager
	transferRepo transferrepository.TransferRepository
	accountRepo  accrepository.AccountRepository
	userRepo     userrepository.UserRepository
}

func NewTransferService(tx database.TxManager, transferRepo transferrepository.TransferRepository, accountRepo accrepository.AccountRepository, userRepo userrepository.UserRepository) TransferService {
	return &transferService{
		tx:           tx,
		transferRepo: transferRepo,
		accountRepo:  accountRepo,
		userRepo:     userRepo,
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
func (s *transferService) TransferMoney(ctx context.Context, userID string, input *transfer.Transfer) (*TransferMoneyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	if input.FromAccountID == input.ToAccountID {
		return nil, errs.ErrTransferSameAccount
	}
	if input.Amount <= 0 {
		return nil, errs.ErrAmountGeaterThanZero
	}

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

	// ------------- Start Transaction ----------------
	//
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
	return resp, err
}

func formatAccountFullName(firstName, lastName string) string {
	return fmt.Sprintf("%s %s", firstName, lastName)
}
