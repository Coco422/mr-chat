package httpx

import "github.com/gin-gonic/gin"

type Envelope struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Error     *ErrorBody  `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Success(c *gin.Context, status int, data interface{}) {
	SuccessWithMeta(c, status, data, nil)
}

func SuccessWithMeta(c *gin.Context, status int, data interface{}, meta interface{}) {
	c.JSON(status, Envelope{
		Success:   true,
		Data:      data,
		Meta:      meta,
		RequestID: RequestIDFromContext(c),
	})
}

func Failure(c *gin.Context, status int, code, message string, details interface{}) {
	c.JSON(status, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: RequestIDFromContext(c),
	})
}

func RequestIDFromContext(c *gin.Context) string {
	requestID, ok := c.Get("request_id")
	if !ok {
		return ""
	}

	id, ok := requestID.(string)
	if !ok {
		return ""
	}

	return id
}
