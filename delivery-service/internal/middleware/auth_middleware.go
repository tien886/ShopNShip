package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RemoveBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	return authHeader, nil
}

func AuthMiddleware(secret string, validateUser func(uint) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		tokenStr, err := RemoveBearer(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		fmt.Println(err)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			sub, ok := claims["sub"].(float64)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
				c.Abort()
				return
			}

			userID := uint(sub)
			if validateUser != nil {
				if err := validateUser(userID); err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "user not registered"})
					c.Abort()
					return
				}
			}

			c.Set("user_id", userID)
			c.Set("role", claims["role"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
	}
}
