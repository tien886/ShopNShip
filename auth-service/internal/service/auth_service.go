package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tien886/ShopNShip/auth-service/internal/model"
	"github.com/tien886/ShopNShip/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService interface {
	Register(email, password, fullName string) error
	Login(email, password string) (string, string, error)
	ValidateToken(tokenStr string) (*jwt.MapClaims, error)
	RefreshToken(tokenStr string) (string, string, error)
}

type authService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo repository.UserRepository, secret string) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

func (s *authService) Register(email, password, fullName string) error {
	existingUser, _ := s.repo.FindByEmail(email)
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Email:    email,
		Password: string(hashedPassword),
		FullName: fullName,
	}

	return s.repo.Create(user)
}

func (s *authService) generateTokens(user *model.User) (string, string, error) {
	// Access Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // 1 day
		"role": user.Role,
		"type": "access",
	})
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"type": "refresh",
	})
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (s *authService) Login(email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

func (s *authService) ValidateToken(tokenStr string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) RefreshToken(tokenStr string) (string, string, error) {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	tokenType, ok := (*claims)["type"].(string)
	if !ok || tokenType != "refresh" {
		return "", "", errors.New("invalid token type")
	}

	// Extract user ID
	sub, ok := (*claims)["sub"].(float64)
	if !ok {
		return "", "", errors.New("invalid token subject")
	}
	userID := uint(sub)

	// Fetch user to ensure they still exist and get updated role
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	return s.generateTokens(user)
}

