package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	fbAuth "firebase.google.com/go/v4/auth"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UserRoleKey contextKey = "userRole"
)

// AuthMiddleware verifies Firebase ID token and adds user info to context
func AuthMiddleware(firebaseAuth *fbAuth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("[Auth] ❌ Missing Authorization header")
				http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
				return
			}
			
			// Expected format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Println("[Auth] ❌ Invalid Authorization header format")
				http.Error(w, "unauthorized: invalid token format", http.StatusUnauthorized)
				return
			}
			
			idToken := parts[1]
			
			// Verify token with Firebase
			if firebaseAuth == nil {
				log.Println("[Auth] ⚠️  Firebase auth not available, allowing request")
				// In development without Firebase, allow through
				ctx := context.WithValue(r.Context(), UserIDKey, "dev-user-123")
				ctx = context.WithValue(ctx, UserRoleKey, "user")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			
			token, err := firebaseAuth.VerifyIDToken(context.Background(), idToken)
			if err != nil {
				log.Printf("[Auth] ❌ Token verification failed: %v\n", err)
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}
			
			// Extract user role from custom claims or default to "user"
			role := "user"
			if claims, ok := token.Claims["role"]; ok {
				if roleStr, ok := claims.(string); ok {
					role = roleStr
				}
			}
			
			log.Printf("[Auth] ✅ Authenticated user: %s (role: %s)\n", token.UID, role)
			
			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, token.UID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin middleware ensures user has admin role
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(string)
		if !ok || role != "admin" {
			log.Printf("[Auth] ❌ Admin access denied for role: %s\n", role)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "forbidden: admin access required",
			})
			return
		}
		
		log.Println("[Auth] ✅ Admin access granted")
		next.ServeHTTP(w, r)
	})
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}

// GetUserRole extracts user role from request context
func GetUserRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(UserRoleKey).(string)
	return role, ok
}
