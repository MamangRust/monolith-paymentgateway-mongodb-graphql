package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlTransferNotFound          = NewGraphqlError("transfer", "Transfer not found", int(http.StatusNotFound))
	ErrGraphqlTransferInvalidID         = NewGraphqlError("transfer", "Invalid Transfer ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransferUserID     = NewGraphqlError("card_id", "Invalid user ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransferCardNumber = NewGraphqlError("card_id", "Invalid card number", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransferMonth      = NewGraphqlError("month", "Invalid month", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransferYear       = NewGraphqlError("year", "Invalid year", int(http.StatusBadRequest))

	ErrGraphqlValidateCreateTransferRequest = NewGraphqlError("transfer", "Invalid input for create transfer", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateTransferRequest = NewGraphqlError("transfer", "Invalid input for update transfer", int(http.StatusBadRequest))
)
