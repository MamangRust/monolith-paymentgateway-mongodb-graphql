package graphqlerror

import (
	"net/http"
)

var (
	ErrGraphqlSaldoNotFound          = NewGraphqlError("saldo", "Saldo not found", int(http.StatusNotFound))
	ErrGraphqlSaldoInvalidID         = NewGraphqlError("saldo", "Invalid Saldo ID", int(http.StatusBadRequest))
	ErrGraphqlSaldoInvalidCardNumber = NewGraphqlError("saldo", "Invalid Saldo Card Number", int(http.StatusBadRequest))
	ErrGraphqlSaldoInvalidMonth      = NewGraphqlError("saldo", "Invalid Saldo Month", int(http.StatusBadRequest))
	ErrGraphqlSaldoInvalidYear       = NewGraphqlError("saldo", "Invalid Saldo Year", int(http.StatusBadRequest))

	ErrGraphqlValidateCreateSaldo = NewGraphqlError("saldo", "Invalid input for create saldo", int(http.StatusBadRequest))
	ErrGraphqlValidateUpdateSaldo = NewGraphqlError("saldo", "Invalid input for update saldo", int(http.StatusBadRequest))
)
