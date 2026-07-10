package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/token"
)

type authKey struct{}

type serviceKey struct{}

func Auth(manager *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				apperror.WriteJSON(w, apperror.Unauthorized("missing bearer token", nil))
				return
			}
			claims, err := manager.Parse(parts[1])
			if err != nil {
				apperror.WriteJSON(w, apperror.Unauthorized("invalid access token", err))
				return
			}
			ctx := context.WithValue(r.Context(), authKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ServiceKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				apperror.WriteJSON(w, apperror.Unauthorized("service key not configured", nil))
				return
			}
			if r.Header.Get("X-Service-Key") != key {
				apperror.WriteJSON(w, apperror.Unauthorized("invalid service key", nil))
				return
			}
			ctx := context.WithValue(r.Context(), serviceKey{}, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Claims(ctx context.Context) *token.Claims {
	if claims, ok := ctx.Value(authKey{}).(*token.Claims); ok {
		return claims
	}
	return nil
}

func UserID(ctx context.Context) string {
	if claims := Claims(ctx); claims != nil {
		return claims.UserID
	}
	return ""
}

func Role(ctx context.Context) models.UserRole {
	if claims := Claims(ctx); claims != nil {
		return claims.Role
	}
	return ""
}

func RequireRole(role models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if Role(r.Context()) != role {
				apperror.WriteJSON(w, apperror.Forbidden("insufficient permissions", nil))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
