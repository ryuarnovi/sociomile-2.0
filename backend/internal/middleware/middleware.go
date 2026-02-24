package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// Enhanced debug: log header length, token length (not value), and proxy source info
		hdrLen := len(authHeader)
		xff := c.GetHeader("X-Forwarded-For")
		remote := c.Request.RemoteAddr
		if authHeader == "" {
			fmt.Printf("[AuthMiddleware] Authorization header missing. method=%s path=%s remote=%s xff=%s hdrLen=%d\n", c.Request.Method, c.Request.RequestURI, remote, xff, hdrLen)
		} else {
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenLen := len(strings.TrimPrefix(authHeader, "Bearer "))
				fmt.Printf("[AuthMiddleware] Authorization present. method=%s path=%s remote=%s xff=%s hdrLen=%d tokenLen=%d\n", c.Request.Method, c.Request.RequestURI, remote, xff, hdrLen, tokenLen)
			} else {
				fmt.Printf("[AuthMiddleware] Authorization present but missing 'Bearer ' prefix. method=%s path=%s remote=%s xff=%s hdrLen=%d\n", c.Request.Method, c.Request.RequestURI, remote, xff, hdrLen)
			}
		}
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Bearer token required"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserFromContext extracts user info from gin context
func GetUserFromContext(c *gin.Context) (userID, tenantID, role string) {
	if val, exists := c.Get("user_id"); exists {
		userID, _ = val.(string)
	}
	if val, exists := c.Get("tenant_id"); exists {
		tenantID, _ = val.(string)
	}
	if val, exists := c.Get("role"); exists {
		role, _ = val.(string)
	}
	return userID, tenantID, role
}
