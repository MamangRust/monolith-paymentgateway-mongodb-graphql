package errors

import (
	"fmt"
)

func InvalidAccessToken() error {
	return fmt.Errorf("invalid access token")
}

// ErrorResponse is the gateway's JSON error envelope. Writing it is the job of
// service/apigateway/apierror (net/http); this package keeps only the shape.
type ErrorResponse struct {
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Type        ErrorType         `json:"type,omitempty"`
	Code        int               `json:"code"`
	TraceID     string            `json:"trace_id,omitempty"`
	Retryable   bool              `json:"retryable,omitempty"`
	Validations []ValidationError `json:"validations,omitempty"`
}

func NewBadRequestError(message string) *AppError {
	return ErrBadRequest.WithMessage(message)
}

func NewNotFoundError(resource string) *AppError {
	return ErrNotFound.WithMessage(fmt.Sprintf("%s not found", resource))
}

func NewConflictError(message string) *AppError {
	return ErrConflict.WithMessage(message)
}

func NewInternalError(err error) *AppError {
	return ErrInternal.WithInternal(err)
}

func NewServiceUnavailableError(service string) *AppError {
	return ErrServiceUnavailable.WithMessage(fmt.Sprintf("%s is temporarily unavailable", service))
}

func NewForbiddenError(message string) *AppError {
	return ErrForbidden.WithMessage(message)
}
