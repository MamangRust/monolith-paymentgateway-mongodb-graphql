package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlTopupNotFound          = NewGraphqlError("topup", "Topup not found", int(http.StatusNotFound))
	ErrGraphqlTopupInvalidID         = NewGraphqlError("topup", "Invalid Topup ID", int(http.StatusBadRequest))
	ErrGraphqlTopupInvalidMonth      = NewGraphqlError("month", "Invalid Topup Month", int(http.StatusBadRequest))
	ErrGraphqlInvalidTopupCardNumber = NewGraphqlError("card_id", "Invalid card number", int(http.StatusBadRequest))
	ErrGraphqlTopupInvalidYear       = NewGraphqlError("year", "Invalid Topup Year", int(http.StatusBadRequest))

	ErrGraphqlValidateCreateTopup = NewGraphqlError("topup", "Invalid input for create topup", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateTopup = NewGraphqlError("topup", "Invalid input for update topup", int(http.StatusBadRequest))
)
