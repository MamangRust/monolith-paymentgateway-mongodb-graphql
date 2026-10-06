package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	mycontext "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/context"
	"github.com/MamangRust/monolith-payment-gateway-pkg/auth"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"go.uber.org/zap"
)

func AuthMiddleware(tm auth.TokenManager, logger logger.LoggerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logger.Debug("Incoming GraphQL request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
			)

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("Failed to read request body", zap.Error(err))
				writeJSONError(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			var bodyMap map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
				if q, ok := bodyMap["query"].(string); ok {
					queryLower := strings.ToLower(q)
					if strings.Contains(queryLower, "loginuser") ||
						strings.Contains(queryLower, "registeruser") ||
						strings.Contains(queryLower, "refreshtoken") {

						logger.Debug("Skipping auth for public operation",
							zap.String("operation", q),
						)
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Debug("Missing Authorization header")
				writeJSONError(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				logger.Debug("Invalid Authorization format",
					zap.String("header", authHeader),
				)
				writeJSONError(w, "invalid Authorization format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			userIDStr, err := tm.ValidateToken(tokenString)
			if err != nil {
				logger.Debug("Token validation failed",
					zap.Error(err),
					zap.String("token", tokenString),
				)
				writeJSONError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				logger.Error("Failed to parse user ID from token", zap.Error(err))
				writeJSONError(w, "invalid user ID in token", http.StatusUnauthorized)
				return
			}

			logger.Debug("Token validated successfully",
				zap.Int("user_id", userID),
			)

			ctx := mycontext.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))

			logger.Debug("Request completed",
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
				zap.Int("user_id", userID),
			)
		})
	}
}

func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]interface{}{
			{"message": message},
		},
	})
}
