package graphqlerror

import (
	"fmt"
	"net/http"

	"github.com/MamangRust/monolith-payment-gateway-shared/domain/response"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGraphqlErrorFromErrorResponse(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("graphql error: %v", err)
	}

	httpCode := grpcToHttpCode(st.Code())
	statusText := http.StatusText(httpCode)
	message := st.Message()

	return NewGraphqlError(statusText, message, httpCode)
}

func NewGraphqlError(statusText string, message string, code int) error {
	errResp := &response.ErrorResponse{
		Status:  statusText,
		Message: message,
		Code:    code,
	}

	return fmt.Errorf("graphql error: [%d] %s - %s", errResp.Code, errResp.Status, errResp.Message)
}

func grpcToHttpCode(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
