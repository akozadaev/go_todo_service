package handler

// ErrorResponse представляет стандартный формат ошибки
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewErrorResponse создает новый ответ с ошибкой
func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Error: message,
	}
}


