package helper

// Response represents the standard API response structure
type Response struct {
	Message string      `json:"message"        example:"Success"`
	Data    interface{} `json:"data,omitempty"`
}

// NewResponse creates a new response object
func NewResponse(message string, data interface{}) Response {
	return Response{
		Message: message,
		Data:    data,
	}
}
