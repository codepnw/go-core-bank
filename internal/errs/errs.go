package errs

import "errors"

// Error User
var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenRevoked           = errors.New("token revoked")
	ErrTokenExpires           = errors.New("token expires")
)

// Err Account
var (
	ErrAccountNotFound                 = errors.New("account not found")
	ErrAccountNotFoundOrMoneyNotEnough = errors.New("account not found or money not enough")
	ErrAmountGeaterThanZero            = errors.New("amount must be greater than zero")
)

// Err Transfer
var (
	ErrTransferSameAccount = errors.New("cannot transfer to the same account")
	ErrForbidden           = errors.New("forbidden: you do not have permission to access this resource")
	ErrDuplicateTransfer   = errors.New("duplicate transfer")
)
