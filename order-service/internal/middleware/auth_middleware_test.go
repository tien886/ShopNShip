package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestRemoveBearer(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantToken string
		wantErr   string
	}{
		{
			name:    "empty header",
			input:   "",
			wantErr: "authorization header is required",
		},
		{
			name:      "with Bearer prefix",
			input:     "Bearer valid-token",
			wantToken: "valid-token",
		},
		{
			name:      "without Bearer prefix",
			input:     "valid-token",
			wantToken: "valid-token",
		},
		{
			name:      "Bearer with empty token",
			input:     "Bearer ",
			wantToken: "",
		},
		{
			name:      "token with spaces",
			input:     "Bearer token with spaces",
			wantToken: "token with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := RemoveBearer(tt.input)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func generateTestToken(secret string, sub float64, role string) string {
	claims := jwt.MapClaims{
		"sub":  sub,
		"role": role,
		"exp":  float64(time.Now().Add(time.Hour).Unix()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	userID := uint(42)
	tokenStr := generateTestToken(secret, float64(userID), "user")

	tests := []struct {
		name         string
		authHeader   string
		validateUser func(uint) error
		wantStatus   int
		wantBody     string
	}{
		{
			name:       "no authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"authorization header is required"}`,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"invalid or expired token"}`,
		},
		{
			name:       "valid Bearer token and valid user",
			authHeader: "Bearer " + tokenStr,
			validateUser: func(uid uint) error {
				assert.Equal(t, userID, uid)
				return nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"ok"}`,
		},
		{
			name:       "valid token without Bearer prefix",
			authHeader: tokenStr,
			validateUser: func(uid uint) error {
				assert.Equal(t, userID, uid)
				return nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"ok"}`,
		},
		{
			name:       "valid token but user not registered",
			authHeader: "Bearer " + tokenStr,
			validateUser: func(uid uint) error {
				assert.Equal(t, userID, uid)
				return errors.New("user not found")
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"user not registered"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			r.Use(AuthMiddleware(secret, tt.validateUser))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
