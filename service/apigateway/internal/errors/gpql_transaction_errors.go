package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlTransactionNotFound          = NewGraphqlError("transaction", "Transaction not found", int(http.StatusNotFound))
	ErrGraphqlTransactionInvalidID         = NewGraphqlError("transaction", "Invalid Transaction ID", int(http.StatusBadRequest))
	ErrGraphqlTransactionInvalidMerchantID = NewGraphqlError("transaction", "Invalid Transaction Merchant ID", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransactionCardNumber = NewGraphqlError("card_id", "Invalid card number", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransactionMonth      = NewGraphqlError("month", "Invalid month", int(http.StatusBadRequest))
	ErrGraphqlInvalidTransactionYear       = NewGraphqlError("year", "Invalid year", int(http.StatusBadRequest))

	ErrGraphqlValidateCreateTransactionRequest = NewGraphqlError("transaction", "Invalid input for create card", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateTransactionRequest = NewGraphqlError("transaction", "Invalid input for update card", int(http.StatusBadRequest))
)
