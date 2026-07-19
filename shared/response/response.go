package response

import (
	"github.com/Astheria23/jokiOrchestrator/shared/contracts"
	"github.com/gin-gonic/gin"
)

// JSON sends a success JSON response using the standard API response structure.
func JSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, contracts.APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends a standard error JSON response.
func Error(c *gin.Context, statusCode int, errMsg string) {
	c.JSON(statusCode, contracts.APIResponse{
		Success: false,
		Error: contracts.APIError{
			Message: errMsg,
		},
	})
}

// ErrorWithCode sends an error JSON response containing a custom application-level error code.
func ErrorWithCode(c *gin.Context, statusCode int, errorCode int, errMsg string) {
	c.JSON(statusCode, contracts.APIResponse{
		Success: false,
		Error: contracts.APIError{
			Code:    errorCode,
			Message: errMsg,
		},
	})
}
