package accservice

import (
	"context"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/codepnw/go-starter-kit/internal/features/account"
	accrepository "github.com/codepnw/go-starter-kit/internal/features/account/repository"
)

type AccountService interface {
	CreateAccount(ctx context.Context, input *account.Account) error
	GetAllAccounts(ctx context.Context, ownerID string) ([]*account.Account, error)
	GetAccountByID(ctx context.Context, ownerID string, accountID int64) (*account.Account, error)
}

type accountService struct {
	repo accrepository.AccountRepository
}

func NewAccountSevice(repo accrepository.AccountRepository) AccountService {
	return &accountService{repo: repo}
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

	return s.repo.FindAllAccounts(ctx, ownerID)
}

// GetAccountByID implements AccountService.
func (s *accountService) GetAccountByID(ctx context.Context, ownerID string, accountID int64) (*account.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, config.ContextTimeout)
	defer cancel()

	return s.repo.FindAccountByID(ctx, ownerID, accountID)
}
