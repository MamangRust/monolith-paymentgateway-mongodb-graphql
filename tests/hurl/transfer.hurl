# Transfer E2E (GraphQL /query) — lifecycle + queries.
# Self-contained: dua user (sender + receiver), masing-masing card + saldo.
# Atomic transfer butuh saldo row di kedua card (sender balance >= amount).
# Jalankan: hurl --variable baseUrl=http://localhost:5000 --variable uuid=<uuid> --test tests/hurl/transfer.hurl

# ── User A (sender) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Transfer\", lastname: \"Sender\", email: \"transfer.a.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"transfer.a.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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
senderUserId: jsonpath "$.data.getMe.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{senderUserId}}, card_type: \"debit\", expire_date: \"2030-12-31\", cvv: \"123\", card_provider: \"visa\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
senderCardNumber: jsonpath "$.data.createCard.data.card_number"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{senderCardNumber}}\", total_balance: 500000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"

# ── User B (receiver) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Transfer\", lastname: \"Receiver\", email: \"transfer.b.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"transfer.b.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
receiverToken: jsonpath "$.data.loginUser.data.access_token"

POST {{baseUrl}}/query
Authorization: Bearer {{receiverToken}}
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"{{receiverToken}}\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.getMe.status" == "success"
[Captures]
receiverUserId: jsonpath "$.data.getMe.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{receiverToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{receiverUserId}}, card_type: \"debit\", expire_date: \"2030-12-31\", cvv: \"456\", card_provider: \"bni\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
receiverCardNumber: jsonpath "$.data.createCard.data.card_number"

POST {{baseUrl}}/query
Authorization: Bearer {{receiverToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{receiverCardNumber}}\", total_balance: 100000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"

# ── Find All Transfers ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransfer(input: {page: 1, page_size: 10}) { status message data { id transfer_from transfer_to } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTransfer.status" == "success"

# ── Create Transfer ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createTransfer(input: {transfer_from: \"{{senderCardNumber}}\", transfer_to: \"{{receiverCardNumber}}\", transfer_amount: 50000}) { status message data { id transfer_from transfer_to transfer_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createTransfer.status" == "success"
[Captures]
transferId: jsonpath "$.data.createTransfer.data.id"

# ── Find Transfer by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdTransfer(input: {transfer_id: {{transferId}}}) { status message data { id transfer_from transfer_to } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdTransfer.status" == "success"

# ── Find Transfers by Transfer From ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findTransferByTransferFrom(input: {transfer_from: \"{{senderCardNumber}}\"}) { status message data { id transfer_from } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findTransferByTransferFrom.status" == "success"

# ── Find Transfers by Transfer To ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findTransferByTransferTo(input: {transfer_to: \"{{receiverCardNumber}}\"}) { status message data { id transfer_to } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findTransferByTransferTo.status" == "success"

# ── Update Transfer ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateTransfer(input: {transfer_id: {{transferId}}, transfer_from: \"{{senderCardNumber}}\", transfer_to: \"{{receiverCardNumber}}\", transfer_amount: 100000}) { status message data { id transfer_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateTransfer.status" == "success"

# ── Find Active Transfers ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveTransfer(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveTransfer.status" == "success"

# ── Trash Transfer ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransfer(input: {transfer_id: {{transferId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransfer.status" == "success"

# ── Find Trashed Transfers ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedTransfer(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedTransfer.status" == "success"

# ── Restore Transfer ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreTransfer(input: {transfer_id: {{transferId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreTransfer.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransfer(input: {transfer_id: {{transferId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransfer.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllTransfer { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllTransfer.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransfer(input: {transfer_id: {{transferId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransfer.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllTransferPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllTransferPermanent.status" == "success"

# ── Transfer kedua — coverage deleteTransferPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createTransfer(input: {transfer_from: \"{{senderCardNumber}}\", transfer_to: \"{{receiverCardNumber}}\", transfer_amount: 60000}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createTransfer.status" == "success"
[Captures]
permTransferId: jsonpath "$.data.createTransfer.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransfer(input: {transfer_id: {{permTransferId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransfer.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteTransferPermanent(input: {transfer_id: {{permTransferId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteTransferPermanent.status" == "success"
