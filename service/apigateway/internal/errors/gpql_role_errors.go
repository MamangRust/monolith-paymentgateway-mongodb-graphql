package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlRoleNotFound      = NewGraphqlError("error", "Role not found", int(http.StatusNotFound))
	ErrGraphqlRoleInvalidId     = NewGraphqlError("error", "Invalid Role ID", int(http.StatusNotFound))
	ErrGraphqlRoleInvalidUserId = NewGraphqlError("error", "Invalid Role User ID", int(http.StatusNotFound))

	ErrGraphqlValidateCreateRole = NewGraphqlError("error", "validation failed: invalid create Role request", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateRole = NewGraphqlError("error", "validation failed: invalid update Role request", int(http.StatusBadRequest))
)
