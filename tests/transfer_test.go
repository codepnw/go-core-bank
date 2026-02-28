package tests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/codepnw/go-starter-kit/internal/config"
	accrepository "github.com/codepnw/go-starter-kit/internal/features/account/repository"
	"github.com/codepnw/go-starter-kit/internal/features/transfer"
	transferrepository "github.com/codepnw/go-starter-kit/internal/features/transfer/repository"
	transferservice "github.com/codepnw/go-starter-kit/internal/features/transfer/service"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	"github.com/codepnw/go-starter-kit/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestTransfer_Deadlock(t *testing.T) {
	// Arrange
	cleanDB(testDB)

	userA := createTestUser(t, testDB, "test1@mail.com", "test", "one")
	userB := createTestUser(t, testDB, "example1@mail.com", "test", "example")

	accA := createTestAccount(t, testDB, userA, "account a", 100)
	accB := createTestAccount(t, testDB, userB, "account b", 100)

	// DI
	tx := database.NewDBTransaction(testDB)
	tranRepo := transferrepository.NewTransferRepository(testDB)
	accRepo := accrepository.NewAccountRepository(testDB)
	userRepo := userrepository.NewUserRepository(testDB)

	service := transferservice.NewTransferService(tx, tranRepo, accRepo, userRepo)

	// Act
	n := 10
	amount := 10
	errs := make(chan error, n)

	for i := range 10 {
		fromAccountID := accA
		toAccountID := accB

		if i%2 == 1 {
			fromAccountID = accB
			toAccountID = accA
		}

		go func() {
			// Set UserID
			ctx := context.Background()
			ownerID := userA
			if fromAccountID == accB {
				ownerID = userB
			}
			ctxAuth := context.WithValue(ctx, config.ContextUserIDKey, ownerID)

			_, err := service.TransferMoney(ctxAuth, ownerID, &transfer.Transfer{
				FromAccountID: int64(fromAccountID),
				ToAccountID:   int64(toAccountID),
				Amount:        int64(amount),
			})

			errs <- err
		}()
	}

	// Assert
	for i := range n {
		_ = i
		err := <-errs
		require.NoError(t, err, "found error (deadlock)")
	}

	balanceAccA, err := getAccountBalance(testDB, accA)
	require.NoError(t, err)

	balanceAccB, err := getAccountBalance(testDB, accB)
	require.NoError(t, err)

	require.Equal(t, 100, balanceAccA, "Balance A Invalid")
	require.Equal(t, 100, balanceAccB, "Balance B Invalid")
}

// -------------- HELPER ----------------------

func cleanDB(db *sql.DB) {
	query := `
		TRUNCATE TABLE users, accounts, entries, transfers
		RESTART IDENTITY CASCADE;
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Clear Test DB failed: %v", err)
	}
	log.Println("Clear Test DB Successfully")
}

func createTestUser(t *testing.T, db *sql.DB, email, firstName, lastName string) string {
	var userID string
	query := `
		INSERT INTO users (email, password, first_name, last_name)
		VALUES ($1, 'mypassword', $2, $3) RETURNING id
	`
	err := db.QueryRow(query, email, firstName, lastName).Scan(&userID)
	if err != nil {
		t.Fatalf("Create Test User failed (%s): %v", email, err)
	}
	return userID
}

func createTestAccount(t *testing.T, db *sql.DB, userID, title string, initialBalance float64) int {
	var accountID int
	query := `
		INSERT INTO accounts (owner_id, title, balance)
		VALUES ($1, $2, $3) RETURNING id
	`
	err := db.QueryRow(query, userID, title, initialBalance).Scan(&accountID)
	if err != nil {
		t.Fatalf("Create Test Account failed: %v", err)
	}
	return accountID
}

func getAccountBalance(db *sql.DB, id int) (int, error) {
	var balance int
	err := db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("Get Account Balance failed: %v", err)
	}
	return balance, nil
}
