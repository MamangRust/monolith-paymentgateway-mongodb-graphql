package graphqlerror

import (
	"net/http"
)

var ErrGraphqlRefreshToken = NewGraphqlError("error", "refresh token failed", int(http.StatusUnauthorized))
