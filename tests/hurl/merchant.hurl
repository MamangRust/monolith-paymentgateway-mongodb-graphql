# Merchant E2E (GraphQL /query) — lifecycle + stats + queries.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Merchant\", lastname: \"Tester\", email: \"merchant.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"merchant.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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

# ── Find All Merchants ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllMerchant(input: {page: 1, page_size: 10}) { status message data { id name api_key } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllMerchant.status" == "success"

# ── Create Merchant ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchant(input: {name: \"Global Store {{uuid}}\", user_id: {{userId}}}) { status message data { id name api_key } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchant.status" == "success"
[Captures]
merchantId: jsonpath "$.data.createMerchant.data.id"
apiKey: jsonpath "$.data.createMerchant.data.api_key"

# ── Find Merchant by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdMerchant(input: {merchant_id: {{merchantId}}}) { status message data { id name api_key } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdMerchant.status" == "success"

# ── Find Merchant by API Key ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByApiKey(input: {api_key: \"{{apiKey}}\"}) { status message data { id name } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByApiKey.status" == "success"

# ── Update Merchant ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateMerchant(input: {merchant_id: {{merchantId}}, name: \"Global Store Updated {{uuid}}\", user_id: {{userId}}, status: \"active\"}) { status message data { id name status } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateMerchant.status" == "success"

# ── Update Merchant Status ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateMerchantStatus(input: {merchant_id: {{merchantId}}, status: \"active\"}) { status message data { id status } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateMerchantStatus.status" == "success"

# ── Find Merchant by User ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByMerchantUserId(input: {user_id: {{userId}}}) { status message data { id name } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByMerchantUserId.status" == "success"

# ── Find All Transaction Merchant ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionMerchant(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200

# ── Find All Transaction By Merchant ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionByMerchant(input: {page: 1, page_size: 10, merchant_id: {{merchantId}}}) { status message data { id } pagination { current_page } } }"
}
HTTP 200

# ── Find All Transaction By ApiKey ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTransactionByApikey(input: {page: 1, page_size: 10, api_key: \"{{apiKey}}\"}) { status message data { id } pagination { current_page } } }"
}
HTTP 200

# ── Find Active Merchants ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActive(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActive.status" == "success"

# ── Stats: monthly/yearly amount & total-amount & payment-methods (semua varian) ──
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

# Stats per-merchant (ByMerchants)
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyAmountByMerchants(input: {year: 2026, merchant_id: {{merchantId}}}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyAmountByMerchants.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyAmountByMerchants(input: {year: 2026, merchant_id: {{merchantId}}}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyAmountByMerchants.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTotalAmountByMerchants(input: {year: 2026, merchant_id: {{merchantId}}}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTotalAmountByMerchants.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTotalAmountByMerchants(input: {year: 2026, merchant_id: {{merchantId}}}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTotalAmountByMerchants.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyPaymentMethodByMerchants(input: {year: 2026, merchant_id: {{merchantId}}}) { status message data { month payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyPaymentMethodByMerchants.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyPaymentMethodByMerchants(input: {year: 2026, merchant_id: {{merchantId}}}) { status message data { year payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyPaymentMethodByMerchants.status" == "success"

# Stats per-apikey (ByApikey)
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyAmountByApikey(input: {year: 2026, api_key: \"{{apiKey}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyAmountByApikey.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyAmountByApikey(input: {year: 2026, api_key: \"{{apiKey}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyAmountByApikey.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyTotalAmountByApikey(input: {year: 2026, api_key: \"{{apiKey}}\"}) { status message data { month total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyTotalAmountByApikey.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyTotalAmountByApikey(input: {year: 2026, api_key: \"{{apiKey}}\"}) { status message data { year total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyTotalAmountByApikey.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMonthlyPaymentMethodByApikey(input: {year: 2026, api_key: \"{{apiKey}}\"}) { status message data { month payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMonthlyPaymentMethodByApikey.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findYearlyPaymentMethodByApikey(input: {year: 2026, api_key: \"{{apiKey}}\"}) { status message data { year payment_method total_amount } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findYearlyPaymentMethodByApikey.status" == "success"

# ── Trash Merchant ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchant(input: {merchant_id: {{merchantId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchant.status" == "success"

# ── Find Trashed Merchants ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashed(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashed.status" == "success"

# ── Restore Merchant ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreMerchant(input: {merchant_id: {{merchantId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreMerchant.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchant(input: {merchant_id: {{merchantId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllMerchant { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchant(input: {merchant_id: {{merchantId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllMerchantPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllMerchantPermanent.status" == "success"

# ── Merchant kedua — coverage deleteMerchantPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchant(input: {name: \"Perm Store {{uuid}}\", user_id: {{userId}}}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchant.status" == "success"
[Captures]
permMerchantId: jsonpath "$.data.createMerchant.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchant(input: {merchant_id: {{permMerchantId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchant.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteMerchantPermanent(input: {merchant_id: {{permMerchantId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteMerchantPermanent.status" == "success"
