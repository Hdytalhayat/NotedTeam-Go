// middlewares/ws_auth_middleware.go
package middlewares

import (
	"net/http"
	"os"
	"strings"

	"notedteam.backend/config"
	"notedteam.backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// WsAuthMiddleware adalah middleware otentikasi yang fleksibel untuk WebSocket.
// Ia mencari token di query param 'token' ATAU di header 'Authorization'.
func WsAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Coba dapatkan token dari query parameter "?token="
		tokenString = c.Query("token")

		// 2. Jika tidak ada di query, coba dapatkan dari header "Authorization"
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		// Jika token tetap tidak ditemukan, batalkan permintaan.
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication token is required"})
			return
		}

		// Proses validasi token (sama seperti AuthMiddleware)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			userID := uint(claims["user_id"].(float64))
			var user models.User
			if err := config.DB.First(&user, userID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User associated with token not found"})
				return
			}

			// Set user_id dan objek user ke context
			c.Set("user_id", user.ID)
			c.Set("user", user)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}
