package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlMerchantNotFound       = NewGraphqlError("merchant", "Merchant not found", int(http.StatusNotFound))
	ErrGraphqlMerchantInvalidID      = NewGraphqlError("merchant", "Invalid Merchant ID", int(http.StatusBadRequest))
	ErrGraphqlMerchantInvalidUserID  = NewGraphqlError("merchant", "Invalid Merchant User ID", int(http.StatusBadRequest))
	ErrGraphqlMerchantInvalidApiKey  = NewGraphqlError("merchant", "Invalid Merchant Api Key", int(http.StatusBadRequest))
	ErrGraphqlMerchantInvalidMonth   = NewGraphqlError("month", "Invalid Merchant Month", int(http.StatusBadRequest))
	ErrGraphqlMerchantInvalidYear    = NewGraphqlError("year", "Invalid Merchant Year", int(http.StatusBadRequest))
	ErrGraphqlValidateCreateMerchant = NewGraphqlError("merchant", "Invalid input for create merchant", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateMerchant = NewGraphqlError("merchant", "Invalid input for update merchant", int(http.StatusBadRequest))
)
