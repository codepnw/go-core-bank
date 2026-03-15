package transferservice_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/account"
	accrepository "github.com/codepnw/go-starter-kit/internal/features/account/repository"
	"github.com/codepnw/go-starter-kit/internal/features/transfer"
	transferrepository "github.com/codepnw/go-starter-kit/internal/features/transfer/repository"
	transferservice "github.com/codepnw/go-starter-kit/internal/features/transfer/service"
	"github.com/codepnw/go-starter-kit/internal/features/user"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var ErrDB = errors.New("database error")

var mockInputTransfer = &transfer.Transfer{
	FromAccountID:  101,
	ToAccountID:    102,
	Amount:         999,
	IdempotencyKey: "MockKey",
}

var mockInputUserID = "mock-user-id-001"

func TestTransferMoney(t *testing.T) {
	type testCase struct {
		name        string
		userID      string
		input       *transfer.Transfer
		mockFn      func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer)
		expectedErr error
	}

	testCases := []testCase{
		{
			name:   "success",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				// ------ Find Account -----------
				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.ToAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				// ------ Find User --------------
				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				mockUserToAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(mockUserToAcc, nil).Times(1)

				// ------ Start Transaction --------------
				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				// ------ Insert Transfer --------------
				tranRepo.EXPECT().InsertTransferTx(gomock.Any(), gomock.Any(), input).Return(nil).Times(1)

				// ------ Insert Entry --------------
				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(nil).Times(1)

				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.ToAccountID, input.Amount).Return(nil).Times(1)

				// ------ Update Account Balance --------------
				accRepo.EXPECT().UpdateAccountBalanceTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(&account.Account{ID: input.FromAccountID}, nil).Times(1)

				accRepo.EXPECT().UpdateAccountBalanceTx(gomock.Any(), gomock.Any(), input.ToAccountID, input.Amount).Return(&account.Account{ID: input.ToAccountID}, nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name:   "fail same account",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				input.FromAccountID = input.ToAccountID
			},
			expectedErr: errs.ErrTransferSameAccount,
		},
		{
			name:   "fail amount is zero",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				input.Amount = 0
			},
			expectedErr: errs.ErrAmountGeaterThanZero,
		},
		{
			name:   "fail userID != OwnerID",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: "wrong-owner-id"}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail from account not found",
			userID: "mock-uuid-user-id-1",
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(nil, ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail to account not found",
			userID: "mock-uuid-user-id-1",
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(nil, ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail from user not found",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				// mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(nil, ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail to user not found",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(nil, ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail insert transfer",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				mockUserToAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(mockUserToAcc, nil).Times(1)

				// ------ Start Transaction --------------
				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				// ------ Insert Transfer --------------
				tranRepo.EXPECT().InsertTransferTx(gomock.Any(), gomock.Any(), input).Return(ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail insert from acc entry",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				mockUserToAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(mockUserToAcc, nil).Times(1)

				// ------ Start Transaction --------------
				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				// ------ Insert Transfer --------------
				tranRepo.EXPECT().InsertTransferTx(gomock.Any(), gomock.Any(), input).Return(nil).Times(1)

				// ------ Insert Entry --------------
				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail insert to acc entry",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				mockUserToAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(mockUserToAcc, nil).Times(1)

				// ------ Start Transaction --------------
				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				// ------ Insert Transfer --------------
				tranRepo.EXPECT().InsertTransferTx(gomock.Any(), gomock.Any(), input).Return(nil).Times(1)

				// ------ Insert Entry --------------
				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(nil).Times(1)

				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.ToAccountID, input.Amount).Return(ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail update from acc balance",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				mockUserToAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(mockUserToAcc, nil).Times(1)

				// ------ Start Transaction --------------
				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				// ------ Insert Transfer --------------
				tranRepo.EXPECT().InsertTransferTx(gomock.Any(), gomock.Any(), input).Return(nil).Times(1)

				// ------ Insert Entry --------------
				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(nil).Times(1)

				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.ToAccountID, input.Amount).Return(nil).Times(1)

				// ------ Update Account Balance --------------
				accRepo.EXPECT().UpdateAccountBalanceTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(nil, ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
		{
			name:   "fail update to acc balance",
			userID: mockInputUserID,
			input:  mockInputTransfer,
			mockFn: func(mockTx *database.MockTxManager, tranRepo *transferrepository.MockTransferRepository, tranRedis *transferrepository.MockTransferRedisRepository, accRepo *accrepository.MockAccountRepository, userRepo *userrepository.MockUserRepository, userID string, input *transfer.Transfer) {
				// ------ Redis Check Idem Key -----------
				tranRedis.EXPECT().CheckIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(true, nil).Times(1)

				mockFromAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.FromAccountID).Return(mockFromAccData, nil).Times(1)

				mockToAccData := &account.Account{ID: input.FromAccountID, OwnerID: userID}
				accRepo.EXPECT().FindAccountByID(gomock.Any(), input.ToAccountID).Return(mockToAccData, nil).Times(1)

				mockUserFromAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockFromAccData.OwnerID).Return(mockUserFromAcc, nil).Times(1)

				mockUserToAcc := &user.User{FirstName: "John", LastName: "Cena"}
				userRepo.EXPECT().FindUserByID(gomock.Any(), mockToAccData.OwnerID).Return(mockUserToAcc, nil).Times(1)

				// ------ Start Transaction --------------
				mockTx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(tx *sql.Tx) error) error {
						return fn(nil)
					},
				).Times(1)

				// ------ Insert Transfer --------------
				tranRepo.EXPECT().InsertTransferTx(gomock.Any(), gomock.Any(), input).Return(nil).Times(1)

				// ------ Insert Entry --------------
				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(nil).Times(1)

				accRepo.EXPECT().InsertEntryTx(gomock.Any(), gomock.Any(), input.ToAccountID, input.Amount).Return(nil).Times(1)

				// ------ Update Account Balance --------------
				accRepo.EXPECT().UpdateAccountBalanceTx(gomock.Any(), gomock.Any(), input.FromAccountID, -input.Amount).Return(&account.Account{ID: input.ToAccountID}, nil).Times(1)

				accRepo.EXPECT().UpdateAccountBalanceTx(gomock.Any(), gomock.Any(), input.ToAccountID, input.Amount).Return(nil, ErrDB).Times(1)

				tranRedis.EXPECT().DeleteIdempotencyKey(gomock.Any(), input.IdempotencyKey).Return(nil).Times(1)
			},
			expectedErr: ErrDB,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setup(t)

			tc.mockFn(s.mockTx, s.tranRepo, s.tranRedis, s.accRepo, s.userRepo, tc.userID, tc.input)

			service := s.service
			resp, err := service.TransferMoney(context.Background(), tc.userID, tc.input)

			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// ================  SETUP DI ========================

type transferSetup struct {
	mockTx    *database.MockTxManager
	tranRepo  *transferrepository.MockTransferRepository
	tranRedis *transferrepository.MockTransferRedisRepository
	accRepo   *accrepository.MockAccountRepository
	userRepo  *userrepository.MockUserRepository
	service   transferservice.TransferService
}

func setup(t *testing.T) *transferSetup {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := database.NewMockTxManager(ctrl)
	tranRepo := transferrepository.NewMockTransferRepository(ctrl)
	tranRedis := transferrepository.NewMockTransferRedisRepository(ctrl)
	accRepo := accrepository.NewMockAccountRepository(ctrl)
	userRepo := userrepository.NewMockUserRepository(ctrl)

	service := transferservice.NewTransferService(&transferservice.TransferServiceDeps{
		Tx:            mockTx,
		TransferRepo:  tranRepo,
		TransferRedis: tranRedis,
		AccountRepo:   accRepo,
		UserRepo:      userRepo,
	})

	return &transferSetup{
		mockTx:    mockTx,
		tranRepo:  tranRepo,
		tranRedis: tranRedis,
		accRepo:   accRepo,
		userRepo:  userRepo,
		service:   service,
	}
}
