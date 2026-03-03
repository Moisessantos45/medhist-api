package middleware

import (
	"api_citas/internal/pkg"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(maker *pkg.PasetoMaker, rd *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// if c.Request.Method == http.MethodOptions {
		// 	log.Println("OPTIONS preflight - skipping auth")
		// 	c.Next() // Deja pasar OPTIONS sin auth
		// }

		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"msg": "missing or invalid Authorization header"})
			return
		}

		token := strings.TrimSpace(authHeader[7:])
		log.Printf("Received token: %s", token[:8]+"...")

		isRevoked, err := rd.Exists(c, token).Result()
		if err != nil {
			log.Printf("Redis EXISTS error: %v", err)
			c.AbortWithStatusJSON(500, gin.H{"msg": "internal error"})
			return
		}

		if isRevoked == 0 {
			log.Printf("Token expired or logged out: %s", token[:8]+"...")
			c.AbortWithStatusJSON(401, gin.H{"msg": "token revoked or expired"})
			return
		}

		payload, err := maker.VerifyToken(token)
		if err != nil {
			log.Printf("Token verification failed: %v", err)
			c.AbortWithStatusJSON(401, gin.H{"msg": "invalid token"})
			return
		}

		userID, err := strconv.ParseUint(payload.UserID, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"msg": "invalid user"})
			return
		}

		c.Set("userID", userID)
		c.Set("payload", payload)
		c.Set("token", token)
		c.Next()
	}

}
