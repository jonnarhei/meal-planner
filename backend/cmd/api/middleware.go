package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/jonnarhei/meal-planner/backend/internal/auth"
	"github.com/jonnarhei/meal-planner/backend/internal/jsonutil"
)

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			jsonutil.WriteError(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		//splitting header into "Bearer" and <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			jsonutil.WriteError(w, "malformed authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(parts[1], app.config.jwt.secret)
		if err != nil {
			jsonutil.WriteError(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := setUserContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type contextKey string

const userContextKey contextKey = "user"

func setUserContext(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, userContextKey, claims)
}

func getUserFromContext(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(userContextKey).(*auth.Claims)
	return claims
}
