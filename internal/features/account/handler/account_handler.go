package acchandler

import (
	"net/http"

	"github.com/codepnw/go-starter-kit/internal/auth"
	"github.com/codepnw/go-starter-kit/internal/errs"
	"github.com/codepnw/go-starter-kit/internal/features/account"
	accservice "github.com/codepnw/go-starter-kit/internal/features/account/service"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	service accservice.AccountService
}

func NewAccountHandler(service accservice.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	req := new(CreateAccountReq)
	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	err = h.service.CreateAccount(c.Request.Context(), &account.Account{
		OwnerID: userID,
		Title:   req.Title,
		Balance: req.Balance,
	})
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	response.ResponseMessage(c, http.StatusCreated, "new account created.")
}

func (h *AccountHandler) GetAllAccounts(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	data, err := h.service.GetAllAccounts(c.Request.Context(), userID)
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	var accounts []*accountResponse

	for _, item := range data {
		accounts = append(accounts, formatAccountResponse(item))
	}

	response.ResponseData(c, http.StatusOK, accounts)
}

func (h *AccountHandler) GetAccountByID(c *gin.Context) {
	accountID := getAccountID(c)

	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	data, err := h.service.GetAccountByID(c.Request.Context(), userID, accountID)
	if err != nil {
		switch err {
		case errs.ErrAccountNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseData(c, http.StatusOK, formatAccountResponse(data))
}

func (h *AccountHandler) DepositMoney(c *gin.Context) {
	accountID := getAccountID(c)

	req := new(AccountDepositReq)
	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	resp, err := h.service.DepositMoney(c.Request.Context(), accountID, req.Amount)
	if err != nil {
		switch err {
		case errs.ErrAccountNotFound:
			response.ResponseError(c, http.StatusNotFound, err)
		case errs.ErrAmountGeaterThanZero:
			response.ResponseError(c, http.StatusBadRequest, err)
		default:
			response.ResponseError(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.ResponseData(c, http.StatusOK, formatAccountResponse(resp))
}
