package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/jonnarhei/meal-planner/backend/internal/auth"
)

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		//splitting header into "Bearer" and <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "malformed authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(parts[1], app.config.jwt.secret)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := setUserContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type contextKey string

const userContextKet contextKey = "user"

func setUserContext(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, userContextKet, claims)
}

func getUserFromContext(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(userContextKet).(*auth.Claims)
	return claims
}