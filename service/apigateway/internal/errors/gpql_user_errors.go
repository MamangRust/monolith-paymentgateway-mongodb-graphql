package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlUserNotFound  = NewGraphqlError("error", "User not found", int(http.StatusNotFound))
	ErrGraphqlUserInvalidId = NewGraphqlError("error", "Invalid User ID", int(http.StatusNotFound))
)
