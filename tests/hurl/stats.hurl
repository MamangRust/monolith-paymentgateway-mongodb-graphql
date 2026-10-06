# =============================================================================
# Stats E2E (GraphQL /query) — mencakup semua query stats per resource.
# Param mengikuti dataset seed (year=2026, month=4).
# Register auto-verify (APP_ENV != production) sehingga login langsung berhasil.
# =============================================================================

# Register user (self-contained)
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Stats\", lastname: \"Tester\", email: \"stats.tester.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"
[Captures]
userId: jsonpath "$.data.registerUser.data.id"

# Login (capture tokens for authenticated requests)
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"stats.tester.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
accessToken: jsonpath "$.data.loginUser.data.access_token"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"{{accessToken}}\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.getMe.status" == "success"

# ── Create Card (parent utk card_number) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{userId}}, card_type: \"debit\", expire_date: \"2030-12-31\", cvv: \"123\", card_provider: \"visa\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
cardNumber: jsonpath "$.data.createCard.data.card_number"

# ── Create Merchant (parent utk api_key / merchant_id queries) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchant(input: {name: \"Stats Store {{uuid}}\", user_id: {{userId}}}) { status message data { id name api_key } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchant.status" == "success"
[Captures]
merchantId: jsonpath "$.data.createMerchant.data.id"
apiKey: jsonpath "$.data.createMerchant.data.api_key"

# -----------------------------------------------------------------------------
# Saldo stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTotalSaldoBalance(input: {year: 2026, month: 4}) { status message data { month year total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTotalSaldoBalance.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearTotalSaldoBalance(input: {year: 2026}) { status message data { year total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearTotalSaldoBalance.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlySaldoBalances(input: {year: 2026}) { status message data { month total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlySaldoBalances.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlySaldoBalances(input: {year: 2026}) { status message data { year total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlySaldoBalances.status" == "success"

# -----------------------------------------------------------------------------
# Card stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyBalance(input: {year: 2026}) { status message data { month total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyBalance.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyBalance(input: {year: 2026}) { status message data { year total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyBalance.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyBalanceByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyBalanceByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyBalanceByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyBalanceByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupAmount(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupAmount(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransactionAmount(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransactionAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransactionAmount(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransactionAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransactionAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransactionAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransactionAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransactionAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawAmount(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawAmount(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferSenderAmount(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferSenderAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferSenderAmount(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferSenderAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferReceiverAmount(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferReceiverAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferReceiverAmount(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferReceiverAmount.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferSenderAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferSenderAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferSenderAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferSenderAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferReceiverAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferReceiverAmountByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferReceiverAmountByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferReceiverAmountByCardNumber.status" == "success"

# -----------------------------------------------------------------------------
# Topup stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupAmounts(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupAmounts.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupAmounts(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupAmounts.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupAmountsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupAmountsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupAmountsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupAmountsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupStatusSuccess(input: {year: 2026, month: 4}) { status message data { month total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupStatusSuccess(input: {year: 2026}) { status message data { year total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupStatusFailed(input: {year: 2026, month: 4}) { status message data { month total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupStatusFailed(input: {year: 2026}) { status message data { year total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupStatusSuccessByCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupStatusSuccessByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupStatusSuccessByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupStatusSuccessByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupStatusFailedByCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupStatusFailedByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupStatusFailedByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupStatusFailedByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupMethods(input: {year: 2026}) { status message data { month topup_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupMethods.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupMethods(input: {year: 2026}) { status message data { year topup_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupMethods.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTopupMethodsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month topup_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTopupMethodsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTopupMethodsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year topup_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTopupMethodsByCardNumber.status" == "success"

# -----------------------------------------------------------------------------
# Transaction stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyAmounts(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyAmounts.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyAmounts(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyAmounts.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyAmountsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyAmountsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyAmountsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyAmountsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransactionStatusSuccess(input: {year: 2026, month: 4}) { status message data { month total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransactionStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransactionStatusSuccess(input: {year: 2026}) { status message data { year total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransactionStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransactionStatusFailed(input: {year: 2026, month: 4}) { status message data { month total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransactionStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransactionStatusFailed(input: {year: 2026}) { status message data { year total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransactionStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransactionStatusSuccessByCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransactionStatusSuccessByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransactionStatusSuccessByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransactionStatusSuccessByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransactionStatusFailedByCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransactionStatusFailedByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransactionStatusFailedByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransactionStatusFailedByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyPaymentMethods(input: {year: 2026}) { status message data { month payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyPaymentMethods.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyPaymentMethods(input: {year: 2026}) { status message data { year payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyPaymentMethods.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyPaymentMethodsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyPaymentMethodsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyPaymentMethodsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyPaymentMethodsByCardNumber.status" == "success"

# -----------------------------------------------------------------------------
# Transfer stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferAmounts(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferAmounts.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferAmounts(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferAmounts.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferAmountsBySenderCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferAmountsBySenderCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferAmountsBySenderCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferAmountsBySenderCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferAmountsByReceiverCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferAmountsByReceiverCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferAmountsByReceiverCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferAmountsByReceiverCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferStatusSuccess(input: {year: 2026, month: 4}) { status message data { month total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferStatusSuccess(input: {year: 2026}) { status message data { year total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferStatusFailed(input: {year: 2026, month: 4}) { status message data { month total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferStatusFailed(input: {year: 2026}) { status message data { year total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferStatusSuccessByCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferStatusSuccessByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferStatusSuccessByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferStatusSuccessByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTransferStatusFailedByCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTransferStatusFailedByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTransferStatusFailedByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTransferStatusFailedByCardNumber.status" == "success"

# -----------------------------------------------------------------------------
# Withdraw stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdraws(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdraws.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdraws(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdraws.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawsByCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawsByCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawStatusSuccess(input: {year: 2026, month: 4}) { status message data { month total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawStatusSuccess(input: {year: 2026}) { status message data { year total_success total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawStatusSuccess.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawStatusFailed(input: {year: 2026, month: 4}) { status message data { month total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawStatusFailed(input: {year: 2026}) { status message data { year total_failed total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawStatusFailed.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawStatusSuccessCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawStatusSuccessCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawStatusSuccessCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_success } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawStatusSuccessCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyWithdrawStatusFailedCardNumber(input: {year: 2026, month: 4, card_number: \"{{cardNumber}}\"}) { status message data { month total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyWithdrawStatusFailedCardNumber.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyWithdrawStatusFailedCardNumber(input: {year: 2026, card_number: \"{{cardNumber}}\"}) { status message data { year total_failed } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyWithdrawStatusFailedCardNumber.status" == "success"

# -----------------------------------------------------------------------------
# Merchant stats
# -----------------------------------------------------------------------------
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyAmountMerchant(input: {year: 2026}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyAmountMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyAmountMerchant(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyAmountMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTotalAmountMerchant(input: {year: 2026}) { status message data { month year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTotalAmountMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTotalAmountMerchant(input: {year: 2026}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTotalAmountMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyPaymentMethodsMerchant(input: {year: 2026}) { status message data { month payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyPaymentMethodsMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyPaymentMethodMerchant(input: {year: 2026}) { status message data { year payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyPaymentMethodMerchant.status" == "success"

# Merchant transactions (reporting)
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionMerchant(input: {page: 1, page_size: 10}) { status message data { id merchant_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTransactionMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionByMerchant(input: {page: 1, page_size: 10, merchant_id: {{merchantId}}}) { status message data { id merchant_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTransactionByMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionByApikey(input: {page: 1, page_size: 10, api_key: \"{{apiKey}}\"}) { status message data { id merchant_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTransactionByApikey.status" == "success"
