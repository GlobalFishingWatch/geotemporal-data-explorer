package types

import "fmt"

const (
	UnauthorizedMessage        = "Unauthorized"
	ForbiddenMessage           = "Forbidden"
	NotFoundMessage            = "Not Found"
	UnprocessableEntityMessage = "Unprocessable Entity"
	ServiceUnavailableMessage  = "Service Unavailable"
)
const (
	NoContent               = 204
	UnauthorizedCode        = 401
	ForbiddenCode           = 403
	NotFoundCode            = 404
	UnprocessableEntityCode = 422
	ServiceUnavailableCode  = 503
)

type MessageError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}
type AppError struct {
	Code         int            `json:"statusCode"`
	ErrorMessage string         `json:"error"`
	Messages     []MessageError `json:"messages"`
}

func NewAppError(code int, errorMessage string, messages []MessageError) *AppError {
	return &AppError{
		Code:         code,
		ErrorMessage: errorMessage,
		Messages:     messages,
	}
}

func (m *AppError) Error() string {
	return fmt.Sprintf("%v", m.Messages)
}

func NewUnauthorizedStandard(message string) *AppError {
	return NewAppError(UnauthorizedCode, UnauthorizedMessage, []MessageError{{
		Title:  UnauthorizedMessage,
		Detail: message,
	}})
}

func NewForbiddenStandard(message string) *AppError {
	return NewAppError(ForbiddenCode, ForbiddenMessage, []MessageError{{
		Title:  ForbiddenMessage,
		Detail: message,
	}})
}

func NewNotFoundStandard(message string) *AppError {
	return NewAppError(NotFoundCode, NotFoundMessage, []MessageError{{
		Title:  NotFoundMessage,
		Detail: message,
	}})
}

func NewUnprocessableEntityStandard(messages []MessageError) *AppError {
	return NewAppError(UnprocessableEntityCode, UnprocessableEntityMessage, messages)
}

func NewServerUnavailableStandard(message string) *AppError {
	return NewAppError(ServiceUnavailableCode, ServiceUnavailableMessage, []MessageError{{
		Title:  ServiceUnavailableMessage,
		Detail: message,
	}})
}
