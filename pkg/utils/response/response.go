package response

import (
	"github.com/gin-gonic/gin"
)

func ResponseData(c *gin.Context, code int, data any) {
	c.JSON(code, gin.H{
		"success": true,
		"code":    code,
		"data":    data,
	})
}

func ResponseMessage(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": true,
		"code":    code,
		"message": message,
	})
}

func ResponseError(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{
		"success": false,
		"code":    code,
		"error":   err.Error(),
	})
}
