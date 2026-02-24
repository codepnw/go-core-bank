package transferhandler

import (
	"net/http"

	"github.com/codepnw/go-starter-kit/internal/auth"
	"github.com/codepnw/go-starter-kit/internal/features/transfer"
	transferservice "github.com/codepnw/go-starter-kit/internal/features/transfer/service"
	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	service transferservice.TransferService
}

func NewTransferHandler(service transferservice.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

func (h *TransferHandler) TransferMoney(c *gin.Context) {
	userID, err := auth.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		response.ResponseError(c, http.StatusUnauthorized, err)
		return
	}

	req := new(TransferMoneyRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	input := &transfer.Transfer{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	resp, err := h.service.TransferMoney(c.Request.Context(), userID, input)
	if err != nil {
		response.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	response.ResponseData(c, http.StatusOK, resp)
}
