package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlInvalidCardID     = NewGraphqlError("card_id", "Invalid card ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidUserID     = NewGraphqlError("card_id", "Invalid user ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidCardNumber = NewGraphqlError("card_id", "Invalid card number", int(http.StatusBadRequest))
	ErrGraphqlInvalidMonth      = NewGraphqlError("month", "Invalid month", int(http.StatusBadRequest))
	ErrGraphqlInvalidYear       = NewGraphqlError("year", "Invalid year", int(http.StatusBadRequest))
)
