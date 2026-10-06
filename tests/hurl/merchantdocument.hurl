# Merchant Document E2E (GraphQL /query) — lifecycle + update-status.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Doc\", lastname: \"Tester\", email: \"doc.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"doc.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
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

# ── Create Merchant (parent) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchant(input: {name: \"Doc Store {{uuid}}\", user_id: {{userId}}}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchant.status" == "success"
[Captures]
merchantId: jsonpath "$.data.createMerchant.data.id"

# ── Find All Merchant Documents ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllMerchantDocuments(input: {page: 1, page_size: 10}) { status message data { document_id merchant_id document_type } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllMerchantDocuments.status" == "success"

# ── Find All Active Merchant Documents ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllActiveMerchantDocuments(input: {page: 1, page_size: 10}) { status message data { document_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllActiveMerchantDocuments.status" == "success"

# ── Find All Trashed Merchant Documents ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllTrashedMerchantDocuments(input: {page: 1, page_size: 10}) { status message data { document_id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllTrashedMerchantDocuments.status" == "success"

# ── Create Merchant Document ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchantDocument(input: {merchant_id: {{merchantId}}, document_type: \"business_license\", document_url: \"https://example.com/doc-{{uuid}}.pdf\"}) { status message data { document_id merchant_id document_type status } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchantDocument.status" == "success"
[Captures]
documentId: jsonpath "$.data.createMerchantDocument.data.document_id"

# ── Find Merchant Document by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findMerchantDocumentById(input: {document_id: {{documentId}}}) { status message data { document_id document_type } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findMerchantDocumentById.status" == "success"

# ── Update Merchant Document ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateMerchantDocument(input: {document_id: {{documentId}}, merchant_id: {{merchantId}}, document_type: \"tax_license\", document_url: \"https://example.com/doc2-{{uuid}}.pdf\", note: \"updated\", status: \"pending\"}) { status message data { document_id document_type status } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateMerchantDocument.status" == "success"

# ── Update Merchant Document Status ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateMerchantDocumentStatus(input: {document_id: {{documentId}}, merchant_id: {{merchantId}}, note: \"approved\", status: \"approved\"}) { status message data { document_id status } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateMerchantDocumentStatus.status" == "success"

# ── Trash Merchant Document ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchantDocument(input: {document_id: {{documentId}}}) { status message data { document_id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchantDocument.status" == "success"

# ── Restore Merchant Document ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreMerchantDocument(input: {document_id: {{documentId}}}) { status message data { document_id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreMerchantDocument.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchantDocument(input: {document_id: {{documentId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchantDocument.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllMerchantDocuments { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllMerchantDocuments.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchantDocument(input: {document_id: {{documentId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchantDocument.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllMerchantDocumentsPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllMerchantDocumentsPermanent.status" == "success"

# ── Document kedua — coverage deleteMerchantDocumentPermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createMerchantDocument(input: {merchant_id: {{merchantId}}, document_type: \"business_license\", document_url: \"https://example.com/perm-{{uuid}}.pdf\"}) { status message data { document_id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createMerchantDocument.status" == "success"
[Captures]
permDocumentId: jsonpath "$.data.createMerchantDocument.data.document_id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedMerchantDocument(input: {document_id: {{permDocumentId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedMerchantDocument.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteMerchantDocumentPermanent(input: {document_id: {{permDocumentId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteMerchantDocumentPermanent.status" == "success"
