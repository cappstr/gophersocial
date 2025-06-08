package apiserver

import (
	"context"
	"github.com/cappstr/GopherSocial/internal/store"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("http request", "method", r.Method, "url", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddleware(jwtManager *JwtManager, userStore *store.UsersStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/v1/auth/") {
				next.ServeHTTP(w, r)
			}
			var token string
			authHeader := r.Header.Get("Authorization")
			if parts := strings.Split(authHeader, "Bearer "); len(parts) == 2 {
				token = parts[1]
			}
			if token == "" {
				slog.Error("auth header not found", "token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parsedToken, err := jwtManager.Parse(token)
			if err != nil {
				slog.Error("auth token parse error", "error", err, "token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !jwtManager.IsAccessToken(parsedToken) {
				slog.Error("invalid token", "token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			userIdStr, err := parsedToken.Claims.GetSubject()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
			}
			userId, err := strconv.Atoi(userIdStr)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user, err := userStore.GetUserById(r.Context(), userId)
			if err != nil {
				slog.Error("failed to get user", "userId", userId, "err", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "user", user)))
		})
	}
}
