# Transaction E2E (GraphQL /query) — lifecycle + queries.
# Self-contained: dua user — payer (card + saldo) & merchant owner (merchant + api_key + settlement card + saldo).
# Atomic transaction butuh saldo row di kedua card & merchant settlement card != card payer.
# Jalankan: hurl --variable baseUrl=http://localhost:5000 --variable uuid=<uuid> --test tests/hurl/transaction.hurl

# ── User A (payer) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Tx\", lastname: \"Payer\", email: \"tx.payer.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"tx.payer.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
payerToken: jsonpath "$.data.loginUser.data.access_token"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"{{payerToken}}\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.getMe.status" == "success"
[Captures]
payerUserId: jsonpath "$.data.getMe.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{payerUserId}}, card_type: \"debit\", expire_date: \"2030-12-31\", cvv: \"123\", card_provider: \"visa\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
payerCardNumber: jsonpath "$.data.createCard.data.card_number"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{payerCardNumber}}\", total_balance: 500000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"

# ── User B (merchant owner) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Tx\", lastname: \"Merchant\", email: \"tx.merchant.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"tx.merchant.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
merchantToken: jsonpath "$.data.loginUser.data.access_token"

POST {{baseUrl}}/query
Authorization: Bearer {{merchantToken}}
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"{{merchantToken}}\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.getMe.status" == "success"
[Captures]
merchantUserId: jsonpath "$.data.getMe.data.id"

# Settlement card merchant (dijamin beda dari card payer — user berbeda)
POST {{baseUrl}}/query
Authorization: Bearer {{merchantToken}}
Content-Type: application/json
{
    "query": "mutation { createCard(input: {user_id: {{merchantUserId}}, card_type: \"credit\", expire_date: \"2031-12-31\", cvv: \"456\", card_provider: \"bni\"}) { status message data { id card_number } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createCard.status" == "success"
[Captures]
merchantCardNumber: jsonpath "$.data.createCard.data.card_number"

POST {{baseUrl}}/query
Authorization: Bearer {{merchantToken}}
Content-Type: application/json
{
    "query": "mutation { createSaldo(input: {card_number: \"{{merchantCardNumber}}\", total_balance: 100000}) { status message data { saldo_id card_number total_balance } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createSaldo.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{merchantToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchant(input: {name: \"Tx Store {{uuid}}\", user_id: {{merchantUserId}}}) { status message data { id api_key } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchant.status" == "success"
[Captures]
merchantId: jsonpath "$.data.createMerchant.data.id"
apiKey: jsonpath "$.data.createMerchant.data.api_key"

# ── Find All Transactions ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { findAllTransaction(input: {page: 1, page_size: 10}) { status message data { id card_number amount } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTransaction.status" == "success"

# ── Create Transaction ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { createTransaction(input: {card_number: \"{{payerCardNumber}}\", amount: 50000, payment_method: \"visa\", merchant_id: {{merchantId}}, transaction_time: \"2026-04-09\", api_key: \"{{apiKey}}\"}) { status message data { id card_number amount merchant_id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createTransaction.status" == "success"
[Captures]
transactionId: jsonpath "$.data.createTransaction.data.id"

# ── Find Transaction by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { findByIdTransaction(input: {transaction_id: {{transactionId}}}) { status message data { id card_number amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdTransaction.status" == "success"

# ── Find Transactions by Card Number ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionByCardNumber(input: {card_number: \"{{payerCardNumber}}\", page: 1, page_size: 10}) { status message data { id card_number } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTransactionByCardNumber.status" == "success"

# ── Find Transactions by Merchant ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { findTransactionByMerchantId(input: {merchant_id: {{merchantId}}}) { status message data { id merchant_id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findTransactionByMerchantId.status" == "success"

# ── Update Transaction ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { updateTransaction(input: {transaction_id: {{transactionId}}, card_number: \"{{payerCardNumber}}\", amount: 100000, payment_method: \"visa\", merchant_id: {{merchantId}}, transaction_time: \"2026-04-09\", api_key: \"{{apiKey}}\"}) { status message data { id amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateTransaction.status" == "success"

# ── Find Active Transactions ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { findByActiveTransaction(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveTransaction.status" == "success"

# ── Trash Transaction ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransaction(input: {transaction_id: {{transactionId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransaction.status" == "success"

# ── Find Trashed Transactions ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedTransaction(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedTransaction.status" == "success"

# ── Restore Transaction ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { restoreTransaction(input: {transaction_id: {{transactionId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreTransaction.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransaction(input: {transaction_id: {{transactionId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransaction.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllTransaction { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllTransaction.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransaction(input: {transaction_id: {{transactionId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransaction.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllTransactionPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllTransactionPermanent.status" == "success"

# ── Transaction kedua — coverage deleteTransactionPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { createTransaction(input: {card_number: \"{{payerCardNumber}}\", amount: 75000, payment_method: \"mastercard\", merchant_id: {{merchantId}}, transaction_time: \"2026-04-09\", api_key: \"{{apiKey}}\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createTransaction.status" == "success"
[Captures]
permTransactionId: jsonpath "$.data.createTransaction.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { trashedTransaction(input: {transaction_id: {{permTransactionId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedTransaction.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{payerToken}}
Content-Type: application/json
{
    "query": "mutation { deleteTransactionPermanent(input: {transaction_id: {{permTransactionId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteTransactionPermanent.status" == "success"
