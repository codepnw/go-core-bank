package acchandler

import (
	"fmt"
	"log"
	"strconv"

	"github.com/codepnw/go-starter-kit/internal/features/account"
	"github.com/gin-gonic/gin"
)

type CreateAccountReq struct {
	Title   string `json:"title"`
	Balance int64  `json:"balance"`
}

type AccountDepositReq struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

type accountResponse struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Balance        int64  `json:"balance"`
	BalanceDisplay string `json:"balance_display"`
}

func formatAccountResponse(acc *account.Account) *accountResponse {
	return &accountResponse{
		ID:             acc.ID,
		Title:          acc.Title,
		Balance:        acc.Balance,
		BalanceDisplay: fmt.Sprintf("%.2f", float64(acc.Balance)/100),
	}
}

func getAccountID(c *gin.Context) int64 {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("get accountID failed: %v\n", err)
		return 0
	}
	return accountID
}