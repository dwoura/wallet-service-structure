package response

import (
	"net/http"
	"wallet-core/pkg/errno"

	"github.com/gin-gonic/gin"
)

// Response defines the standard JSON structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

// Success returns a success response with data
func Success(c *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{} // Return empty object instead of null
	}
	c.JSON(http.StatusOK, Response{
		Code:    errno.OK.Code,
		Message: errno.OK.Message,
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, err error) {
	code, msg := errno.Decode(err)
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    gin.H{},
	})
}
