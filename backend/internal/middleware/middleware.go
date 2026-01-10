package middleware

import (
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT middleware validates JWT tokens
func JWT(secret string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.Render(401, nil)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.Render(401, nil)
			}

			tokenString := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return c.Render(401, nil)
			}

			// Set user info in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			userRole, ok := c.Value("role").(string)
			if !ok {
				return c.Render(403, nil)
			}

			for _, role := range roles {
				if userRole == role {
					return next(c)
				}
			}

			return c.Render(403, nil)
		}
	}
}

// RequestID adds a unique request ID to each request
func RequestID() buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateID()
			}
			c.Response().Header().Set("X-Request-ID", requestID)
			c.Set("request_id", requestID)
			return next(c)
		}
	}
}

func generateID() string {
	// Simple ID generation - use UUID in production
	return "req_" + randomString(16)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
