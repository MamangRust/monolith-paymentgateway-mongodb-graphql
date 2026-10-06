# Card E2E (GraphQL /query) — lifecycle + dashboard + semua stats.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Card\", lastname: \"Tester\", email: \"card.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"card.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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
[Captures]
userId: jsonpath "$.data.getMe.data.id"

# ── Find All Cards ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllCard(input: {page: 1, page_size: 10}) { status message data { id card_number card_type } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllCard.status" == "success"

# ── Create Card ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{userId}}, card_type: \"debit\", expire_date: \"2030-12-31\", cvv: \"123\", card_provider: \"visa\"}) { status message data { id card_number card_type } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
cardId: jsonpath "$.data.createCard.data.id"
cardNumber: jsonpath "$.data.createCard.data.card_number"

# ── Find Card by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdCard(input: {card_id: {{cardId}}}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdCard.status" == "success"

# ── Find Card by User ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByUserIdCard(input: {user_id: {{userId}}}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByUserIdCard.status" == "success"

# ── Find Card by Card Number ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByCardNumber(input: {card_number: \"{{cardNumber}}\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByCardNumber.status" == "success"

# ── Dashboard Card ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { dashboardCard { status message data { total_balance total_topup total_withdraw total_transaction total_transfer } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.dashboardCard.status" == "success"

# ── Dashboard Card by Number ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { dashboardCardNumber(input: {card_number: \"{{cardNumber}}\"}) { status message data { total_balance total_topup total_withdraw total_transaction total_transfer_send total_transfer_receiver } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.dashboardCardNumber.status" == "success"

# ── Find Active Cards ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveCard(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveCard.status" == "success"

# ── Stats: balance / topup / withdraw / transaction / transfer (monthly & yearly) ──
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

# ── Stats by card number ──
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

# ── Update Card ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateCard(input: {card_id: {{cardId}}, user_id: {{userId}}, card_type: \"credit\", expire_date: \"2031-12-31\", cvv: \"321\", card_provider: \"mastercard\"}) { status message data { id card_type } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateCard.status" == "success"

# ── Trash Card ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedCard(input: {card_id: {{cardId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedCard.status" == "success"

# ── Find Trashed Cards ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedCard(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedCard.status" == "success"

# ── Restore Card ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreCard(input: {card_id: {{cardId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreCard.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedCard(input: {card_id: {{cardId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedCard.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllCard { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllCard.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedCard(input: {card_id: {{cardId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedCard.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllCardPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllCardPermanent.status" == "success"

# ── Card kedua — coverage deleteCardPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{userId}}, card_type: \"debit\", expire_date: \"2032-12-31\", cvv: \"999\", card_provider: \"visa\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
permCardId: jsonpath "$.data.createCard.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedCard(input: {card_id: {{permCardId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedCard.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteCardPermanent(input: {card_id: {{permCardId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteCardPermanent.status" == "success"
