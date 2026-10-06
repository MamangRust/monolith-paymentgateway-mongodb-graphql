package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/MamangRust/monolith-payment-gateway-apigateway/testhelper"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

// NewGraphQLHandler wires the apigateway GraphQL gateway (via its public
// testhelper package) to the given in-process gRPC connections, keyed by
// service name (auth, user, role, card, merchant, saldo, topup, transaction,
// transfer, withdraw).
func NewGraphQLHandler(conns map[string]*grpc.ClientConn, redisClient *redis.Client, log logger.LoggerInterface) http.Handler {
	return testhelper.NewTestHandler(conns, log, redisClient)
}

// NewAuthenticatedGraphQLHandler behaves like NewGraphQLHandler but wraps the
// handler with the production auth middleware, so protected operations (e.g.
// getMe) resolve the caller from the Authorization: Bearer header.
func NewAuthenticatedGraphQLHandler(conns map[string]*grpc.ClientConn, redisClient *redis.Client, log logger.LoggerInterface, tm auth.TokenManager) http.Handler {
	return testhelper.NewAuthenticatedTestHandler(conns, log, redisClient, tm)
}

// DoGraphQL executes a GraphQL document (no auth header) and returns the
// decoded JSON body. A non-200 response or undecodable body is an error.
func DoGraphQL(h http.Handler, query string) (map[string]any, error) {
	return doGraphQLBody(h, func() (*http.Request, error) {
		payload, err := json.Marshal(map[string]any{"query": query})
		if err != nil {
			return nil, err
		}
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
}

// DoGraphQLWithAuth executes a GraphQL document with an Authorization: Bearer
// header and returns the decoded JSON body.
func DoGraphQLWithAuth(h http.Handler, query string, bearerToken string) (map[string]any, error) {
	return doGraphQLBody(h, func() (*http.Request, error) {
		payload, err := json.Marshal(map[string]any{"query": query})
		if err != nil {
			return nil, err
		}
		req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if bearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+bearerToken)
		}
		return req, nil
	})
}

func doGraphQLBody(h http.Handler, build func() (*http.Request, error)) (map[string]any, error) {
	req, err := build()
	if err != nil {
		return nil, err
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("graphql request returned HTTP %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		return nil, fmt.Errorf("failed to decode graphql response: %w: %s", err, rec.Body.String())
	}
	return body, nil
}

// GraphQLErrorMessages extracts the messages of any top-level GraphQL errors.
func GraphQLErrorMessages(body map[string]any) []string {
	raw, ok := body["errors"].([]any)
	if !ok {
		return nil
	}
	messages := make([]string, 0, len(raw))
	for _, e := range raw {
		em, ok := e.(map[string]any)
		if !ok {
			continue
		}
		if msg, ok := em["message"].(string); ok {
			messages = append(messages, msg)
		}
	}
	return messages
}

// GraphQLOpData returns the `data.<op>` object of a GraphQL response.
func GraphQLOpData(body map[string]any, op string) map[string]any {
	data, ok := body["data"].(map[string]any)
	if !ok {
		return nil
	}
	res, _ := data[op].(map[string]any)
	return res
}

// GraphQLOpEnvelopeData returns the `data.<op>.data` payload, i.e. the inner
// domain payload of the standard {status, message, data} API envelope.
func GraphQLOpEnvelopeData(body map[string]any, op string) map[string]any {
	opRes := GraphQLOpData(body, op)
	if opRes == nil {
		return nil
	}
	payload, _ := opRes["data"].(map[string]any)
	return payload
}
