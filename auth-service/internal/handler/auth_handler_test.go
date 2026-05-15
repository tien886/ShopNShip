package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tien886/ShopNShip/auth-service/internal/dto"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(email, password, fullName string) error {
	args := m.Called(email, password, fullName)
	return args.Error(0)
}

func (m *MockAuthService) Login(email, password string) (string, string, error) {
	args := m.Called(email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateToken(tokenStr string) (*jwt.MapClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.MapClaims), args.Error(1)
}

func (m *MockAuthService) RefreshToken(tokenStr string) (string, string, error) {
	args := m.Called(tokenStr)
	return args.String(0), args.String(1), args.Error(2)
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)
	h := NewAuthHandler(mockSvc)

	r := gin.Default()
	r.POST("/register", h.Register)

	t.Run("successful registration", func(t *testing.T) {
		reqBody, _ := json.Marshal(dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			FullName: "Test User",
		})

		mockSvc.On("Register", "test@example.com", "password123", "Test User").Return(nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
