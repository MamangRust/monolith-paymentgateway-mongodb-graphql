# Saldo E2E (GraphQL /query) — lifecycle + stats balance.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Saldo\", lastname: \"Tester\", email: \"saldo.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"saldo.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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

# ── Find All Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllSaldo(input: {page: 1, page_size: 10}) { status message data { saldo_id card_number total_balance } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllSaldo.status" == "success"

# ── Create Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{cardNumber}}\", total_balance: 100000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"
[Captures]
saldoId: jsonpath "$.data.createSaldo.data.saldo_id"

# ── Find Saldo by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdSaldo(input: {saldo_id: {{saldoId}}}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdSaldo.status" == "success"

# ── Find Saldo by Card Number ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByCardNumberSaldo(card_number: \"{{cardNumber}}\") { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByCardNumberSaldo.status" == "success"

# ── Update Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateSaldo(input: {saldo_id: {{saldoId}}, card_number: \"{{cardNumber}}\", total_balance: 150000}) { status message data { saldo_id total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateSaldo.status" == "success"

# ── Find Active Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveSaldo(input: {page: 1, page_size: 10}) { status message data { saldo_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveSaldo.status" == "success"

# ── Stats: monthly/yearly total balance & balances ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTotalSaldoBalance(input: {year: 2026, month: 6}) { status message data { month year total_balance } } }"
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

# ── Trash Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedSaldo(input: {saldo_id: {{saldoId}}}) { status message data { saldo_id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedSaldo.status" == "success"

# ── Find Trashed Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedSaldo(input: {page: 1, page_size: 10}) { status message data { saldo_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedSaldo.status" == "success"

# ── Restore Saldo ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreSaldo(input: {saldo_id: {{saldoId}}}) { status message data { saldo_id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreSaldo.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedSaldo(input: {saldo_id: {{saldoId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedSaldo.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllSaldo { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllSaldo.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedSaldo(input: {saldo_id: {{saldoId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedSaldo.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllSaldoPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllSaldoPermanent.status" == "success"

# ── Saldo kedua — coverage deleteSaldoPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{cardNumber}}\", total_balance: 50000}) { status message data { saldo_id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"
[Captures]
permSaldoId: jsonpath "$.data.createSaldo.data.saldo_id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedSaldo(input: {saldo_id: {{permSaldoId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedSaldo.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteSaldoPermanent(input: {saldo_id: {{permSaldoId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteSaldoPermanent.status" == "success"
