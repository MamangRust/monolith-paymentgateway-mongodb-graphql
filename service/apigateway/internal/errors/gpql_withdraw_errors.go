package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlWithdrawNotFound          = NewGraphqlError("withdraw", "Withdraw not found", int(http.StatusNotFound))
	ErrGraphqlWithdrawInvalidID         = NewGraphqlError("withdraw", "Invalid Withdraw ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidWithdrawUserID     = NewGraphqlError("card_id", "Invalid user ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidWithdrawCardNumber = NewGraphqlError("card_id", "Invalid card number", int(http.StatusBadRequest))
	ErrGraphqlInvalidWithdrawMonth      = NewGraphqlError("month", "Invalid month", int(http.StatusBadRequest))
	ErrGraphqlInvalidWithdrawYear       = NewGraphqlError("year", "Invalid year", int(http.StatusBadRequest))

	ErrGraphqlValidateCreateWithdrawRequest = NewGraphqlError("withdraw", "Invalid input for create withdraw", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateWithdrawRequest = NewGraphqlError("withdraw", "Invalid input for update withdraw", int(http.StatusBadRequest))
)
