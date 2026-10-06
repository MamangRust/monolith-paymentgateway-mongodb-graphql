# User E2E (GraphQL /query) — lifecycle + semua query.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"User\", lastname: \"Tester\", email: \"user.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"user.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
accessToken: jsonpath "$.data.loginUser.data.access_token"
refreshToken: jsonpath "$.data.loginUser.data.refresh_token"

# Get Me → ambil userId untuk input create
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"{{accessToken}}\"}) { status message data { id email } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.getMe.status" == "success"
[Captures]
userId: jsonpath "$.data.getMe.data.id"

# ── Find All Users ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllUser(input: {page: 1, page_size: 10}) { status message data { id email } pagination { current_page page_size total_pages total_records } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllUser.status" == "success"

# ── Create User ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createUser(input: {firstname: \"Alice\", lastname: \"Smith\", email: \"alice.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id email } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createUser.status" == "success"
[Captures]
newUserId: jsonpath "$.data.createUser.data.id"

# ── Find User by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdUser(input: {id: {{newUserId}}}) { status message data { id email } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdUser.status" == "success"

# ── Update User ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateUser(input: {id: {{newUserId}}, firstname: \"Alicia\", lastname: \"Smith\", email: \"alice.{{uuid}}@example.com\", password: \"password123456\", confirm_password: \"password123456\"}) { status message data { id firstname } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateUser.status" == "success"

# ── Find Active Users ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveUser(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveUser.status" == "success"

# ── Invalid id → error ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdUser(input: {id: 99999999}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.errors[0].message" exists

# ── Trash User ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedUser(input: {id: {{newUserId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedUser.status" == "success"

# ── Find Trashed Users ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedUser(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedUser.status" == "success"

# ── Restore User ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreUser(input: {id: {{newUserId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreUser.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedUser(input: {id: {{newUserId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedUser.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllUser { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllUser.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedUser(input: {id: {{newUserId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedUser.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllUserPermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllUserPermanent.status" == "success"

# ── User kedua — coverage DELETE /permanent/:id ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createUser(input: {firstname: \"Perm\", lastname: \"User\", email: \"perm.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createUser.status" == "success"
[Captures]
permUserId: jsonpath "$.data.createUser.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedUser(input: {id: {{permUserId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedUser.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteUserPermanent(input: {id: {{permUserId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteUserPermanent.status" == "success"
