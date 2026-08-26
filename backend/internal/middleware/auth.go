package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lfhonda/metappstwo.git/backend/internal/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format. Use: Bearer {token}",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("token_claims", claims)

		log.Printf("[AUTH] User %d (%s) accessing %s", claims.UserID, claims.Email, c.Request.URL.Path)

		c.Next()
	}
}

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal error",
				"message": "Invalid role type",
			})
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userRoleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Insufficient permissions. Required: " + strings.Join(allowedRoles, ", "),
		})
		c.Abort()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		executionTime := time.Since(startTime)

		log.Printf(
			"[%s] %s %s | Status: %d | Duration: %v | IP: %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			c.Writer.Status(),
			executionTime,
			c.ClientIP(),
		)
	}
}

func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	var visitors = make(map[string][]time.Time)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		var recentVisits []time.Time
		for _, t := range visitors[ip] {
			if now.Sub(t) < time.Minute {
				recentVisits = append(recentVisits, t)
			}
		}
		visitors[ip] = recentVisits

		if len(recentVisits) >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		visitors[ip] = append(visitors[ip], now)
		c.Next()
	}
}
