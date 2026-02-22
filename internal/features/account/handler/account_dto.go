package acchandler

import "fmt"

type CreateAccountReq struct {
	Title   string `json:"title"`
	Balance int64  `json:"balance"`
}

type AccountResponse struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Balance        int64  `json:"balance"`
	BalanceDisplay string `json:"balance_display"`
}

func formatBalanceString(balance int64) string {
	return fmt.Sprintf("%.2f", float64(balance)/100)
}
