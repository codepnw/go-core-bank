package acchandler

import (
	"net/http"
	"strconv"

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

	var accounts []*AccountResponse

	for _, item := range data {
		accounts = append(accounts, &AccountResponse{
			ID:             item.ID,
			Title:          item.Title,
			Balance:        item.Balance,
			BalanceDisplay: formatBalanceString(item.Balance),
		})
	}
	
	response.ResponseData(c, http.StatusOK, accounts)
}

func (h *AccountHandler) GetAccountByID(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

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

	resp := AccountResponse{
		ID:             data.ID,
		Title:          data.Title,
		Balance:        data.Balance,
		BalanceDisplay: formatBalanceString(data.Balance),
	}
	
	response.ResponseData(c, http.StatusOK, resp)
}
