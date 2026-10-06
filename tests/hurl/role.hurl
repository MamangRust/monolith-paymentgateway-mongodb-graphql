# Role E2E (GraphQL /query) — lifecycle + semua query.
# Self-contained: register + login (auto-verify di non-production).

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Role\", lastname: \"Tester\", email: \"role.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"

POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"role.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
accessToken: jsonpath "$.data.loginUser.data.access_token"

# Get Me → userId
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

# ── Find All Roles ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findAllRole(input: {page: 1, page_size: 10}) { status message data { id name } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findAllRole.status" == "success"

# ── Find Active Roles ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByActiveRole(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByActiveRole.status" == "success"

# ── Find Trashed Roles ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByTrashedRole(input: {page: 1, page_size: 10}) { status message data { id } pagination { current_page } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByTrashedRole.status" == "success"

# ── Find Role by User ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByUserIdRole(input: {user_id: {{userId}}}) { status message data { id name } } }"
}
HTTP 200

# ── Create Role ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createRole(input: {name: \"ROLE_TEST_{{uuid}}\"}) { status message data { id name } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createRole.status" == "success"
[Captures]
roleId: jsonpath "$.data.createRole.data.id"

# ── Find Role by ID ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { findByIdRole(input: {role_id: {{roleId}}}) { status message data { id name } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.findByIdRole.status" == "success"

# ── Update Role ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { updateRole(input: {id: {{roleId}}, name: \"ROLE_UPD_{{uuid}}\"}) { status message data { id name } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.updateRole.status" == "success"

# ── Trash Role ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedRole(input: {role_id: {{roleId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedRole.status" == "success"

# ── Restore Role ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreRole(input: {role_id: {{roleId}}}) { status message data { id deleted_at } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreRole.status" == "success"

# ── Trash lagi → restoreAll → trash lagi → deleteAllPermanent ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedRole(input: {role_id: {{roleId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedRole.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { restoreAllRole { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.restoreAllRole.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedRole(input: {role_id: {{roleId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedRole.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteAllRolePermanent { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteAllRolePermanent.status" == "success"

# ── Role kedua — coverage deleteRolePermanent (permanent/:id) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { createRole(input: {name: \"ROLE_PERM_{{uuid}}\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.createRole.status" == "success"
[Captures]
permRoleId: jsonpath "$.data.createRole.data.id"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { trashedRole(input: {role_id: {{permRoleId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.trashedRole.status" == "success"

POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { deleteRolePermanent(input: {role_id: {{permRoleId}}}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.deleteRolePermanent.status" == "success"
