package transferhandler

type TransferMoneyRequest struct {
	FromAccountID  int64  `json:"from_account_id" binding:"required"`
	ToAccountID    int64  `json:"to_account_id" binding:"required"`
	Amount         int64  `json:"amount" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}
