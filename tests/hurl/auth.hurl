# Auth E2E (GraphQL /query) — register → login → me → refresh → forgot/reset/verify.
# Self-contained: register + login (auto-verify di non-production).
# Jalankan: hurl --variable baseUrl=http://localhost:5000 --variable uuid=<uuid> --test tests/hurl/auth.hurl

# ── Register (public) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { registerUser(input: {firstname: \"Auth\", lastname: \"Tester\", email: \"auth.{{uuid}}@example.com\", password: \"password123\", confirm_password: \"password123\"}) { status message data { id } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.registerUser.status" == "success"
[Captures]
userId: jsonpath "$.data.registerUser.data.id"

# ── Login (public) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { loginUser(input: {email: \"auth.{{uuid}}@example.com\", password: \"password123\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.loginUser.status" == "success"
[Captures]
accessToken: jsonpath "$.data.loginUser.data.access_token"
refreshToken: jsonpath "$.data.loginUser.data.refresh_token"

# ── Get Me (authorized) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"{{accessToken}}\"}) { status message data { id email } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.getMe.status" == "success"

# ── Get Me tanpa token → 401 (auth middleware) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "query { getMe(input: {access_token: \"invalid\"}) { status message } }"
}
HTTP 401

# ── Refresh Token (public) ──
POST {{baseUrl}}/query
Content-Type: application/json
{
    "query": "mutation { refreshToken(input: {refresh_token: \"{{refreshToken}}\"}) { status message data { access_token refresh_token } } }"
}
HTTP 200
[Asserts]
jsonpath "$.data.refreshToken.status" == "success"

# ── Forgot Password (email terdaftar; butuh auth middleware) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { forgotPassword(input: {email: \"auth.{{uuid}}@example.com\"}) { status message } }"
}
HTTP 200

# ── Reset Password dengan token tidak valid → error (tidak membocorkan info) ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { resetPassword(input: {reset_token: \"invalid-token\", password: \"newpassword123\", confirm_password: \"newpassword123\"}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.errors[0].message" exists

# ── Verify Code — kode tidak dikenal → error ──
POST {{baseUrl}}/query
Authorization: Bearer {{accessToken}}
Content-Type: application/json
{
    "query": "mutation { verifyCode(input: {code: \"invalid-code\"}) { status message } }"
}
HTTP 200
[Asserts]
jsonpath "$.errors[0].message" exists
