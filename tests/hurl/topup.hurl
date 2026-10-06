# Topup E2E (GraphQL /query) — lifecycle + queries.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Topup\", lastname: \"Tester\", email: \"topup.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"topup.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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

# ── Create Saldo (parent utk topup — atomic topup butuh saldo row) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{cardNumber}}\", total_balance: 50000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"

# ── Find All Topups ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTopup(input: {page: 1, page_size: 10}) { status message data { id card_number topup_amount } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTopup.status" == "success"

# ── Create Topup ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createTopup(input: {card_number: \"{{cardNumber}}\", topup_no: \"TP-{{uuid}}\", topup_amount: 50000, topup_method: \"visa\"}) { status message data { id card_number topup_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createTopup.status" == "success"
[Captures]
topupId: jsonpath "$.data.createTopup.data.id"

# ── Find Topup by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdTopup(input: {topup_id: {{topupId}}}) { status message data { id card_number topup_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdTopup.status" == "success"

# ── Find Topups by Card Number ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTopupByCardNumber(input: {card_number: \"{{cardNumber}}\", page: 1, page_size: 10}) { status message data { id card_number } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTopupByCardNumber.status" == "success"

# ── Find Topup by Card Number (single) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByCardNumberTopup(input: {card_number: \"{{cardNumber}}\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByCardNumberTopup.status" == "success"

# ── Update Topup ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateTopup(input: {topup_id: {{topupId}}, card_number: \"{{cardNumber}}\", topup_amount: 100000, topup_method: \"visa\"}) { status message data { id topup_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateTopup.status" == "success"

# ── Find Active Topups ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveTopup(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveTopup.status" == "success"

# ── Trash Topup ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTopup(input: {topup_id: {{topupId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTopup.status" == "success"

# ── Find Trashed Topups ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedTopup(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedTopup.status" == "success"

# ── Restore Topup ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreTopup(input: {topup_id: {{topupId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreTopup.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTopup(input: {topup_id: {{topupId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTopup.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllTopup { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllTopup.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTopup(input: {topup_id: {{topupId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTopup.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllTopupPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllTopupPermanent.status" == "success"

# ── Topup kedua — coverage deleteTopupPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createTopup(input: {card_number: \"{{cardNumber}}\", topup_no: \"TP-PERM-{{uuid}}\", topup_amount: 60000, topup_method: \"mastercard\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createTopup.status" == "success"
[Captures]
permTopupId: jsonpath "$.data.createTopup.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTopup(input: {topup_id: {{permTopupId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTopup.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteTopupPermanent(input: {topup_id: {{permTopupId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteTopupPermanent.status" == "success"
