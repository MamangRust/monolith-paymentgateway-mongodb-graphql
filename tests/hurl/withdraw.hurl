# Withdraw E2E (GraphQL /query) — lifecycle + queries.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Withdraw\", lastname: \"Tester\", email: \"withdraw.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"withdraw.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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

# ── Create Saldo (parent utk withdraw — atomic withdraw butuh saldo + balance) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{cardNumber}}\", total_balance: 500000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"

# ── Find All Withdraws ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllWithdraw(input: {page: 1, page_size: 10}) { status message data { withdraw_id card_number withdraw_amount } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllWithdraw.status" == "success"

# ── Create Withdraw ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createWithdraw(input: {card_number: \"{{cardNumber}}\", withdraw_amount: 50000, withdraw_time: \"2026-04-09\"}) { status message data { withdraw_id card_number withdraw_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createWithdraw.status" == "success"
[Captures]
withdrawId: jsonpath "$.data.createWithdraw.data.withdraw_id"

# ── Find Withdraw by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdWithdraw(input: {withdraw_id: {{withdrawId}}}) { status message data { withdraw_id card_number withdraw_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdWithdraw.status" == "success"

# ── Find Withdraws by Card Number ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllWithdrawByCardNumber(input: {card_number: \"{{cardNumber}}\", page: 1, page_size: 10}) { status message data { withdraw_id card_number } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllWithdrawByCardNumber.status" == "success"

# ── Find Withdraw by Card Number (single) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByCardNumberWithdraw(input: {card_number: \"{{cardNumber}}\"}) { status message data { withdraw_id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByCardNumberWithdraw.status" == "success"

# ── Update Withdraw ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateWithdraw(input: {withdraw_id: {{withdrawId}}, card_number: \"{{cardNumber}}\", withdraw_amount: 100000, withdraw_time: \"2026-04-09\"}) { status message data { withdraw_id withdraw_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateWithdraw.status" == "success"

# ── Find Active Withdraws ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveWithdraw(input: {page: 1, page_size: 10}) { status message data { withdraw_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveWithdraw.status" == "success"

# ── Trash Withdraw ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedWithdraw(input: {withdraw_id: {{withdrawId}}}) { status message data { withdraw_id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedWithdraw.status" == "success"

# ── Find Trashed Withdraws ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedWithdraw(input: {page: 1, page_size: 10}) { status message data { withdraw_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedWithdraw.status" == "success"

# ── Restore Withdraw ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreWithdraw(input: {withdraw_id: {{withdrawId}}}) { status message data { withdraw_id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreWithdraw.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedWithdraw(input: {withdraw_id: {{withdrawId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedWithdraw.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllWithdraw { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllWithdraw.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedWithdraw(input: {withdraw_id: {{withdrawId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedWithdraw.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllWithdrawPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllWithdrawPermanent.status" == "success"

# ── Withdraw kedua — coverage deleteWithdrawPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createWithdraw(input: {card_number: \"{{cardNumber}}\", withdraw_amount: 60000, withdraw_time: \"2026-04-09\"}) { status message data { withdraw_id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createWithdraw.status" == "success"
[Captures]
permWithdrawId: jsonpath "$.data.createWithdraw.data.withdraw_id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedWithdraw(input: {withdraw_id: {{permWithdrawId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedWithdraw.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteWithdrawPermanent(input: {withdraw_id: {{permWithdrawId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteWithdrawPermanent.status" == "success"
