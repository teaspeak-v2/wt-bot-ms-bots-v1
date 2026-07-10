package middleware

import (
	"log/slog"
	"net/http"

	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/apperror"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if logger != nil {
						logger.Error("panic recovered", "error", rec, "request_id", RequestIDFromContext(r.Context()))
					}
					apperror.WriteJSON(w, apperror.Internal("internal server error", nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
