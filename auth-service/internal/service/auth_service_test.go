package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tien886/ShopNShip/auth-service/internal/model"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authSvc := NewAuthService(mockRepo, "secret")

	t.Run("successful registration", func(t *testing.T) {
		mockRepo.On("FindByEmail", "test@example.com").Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything).Return(nil).Once()

		err := authSvc.Register("test@example.com", "password123", "Test User")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockRepo.On("FindByEmail", "exists@example.com").Return(&model.User{Email: "exists@example.com"}, nil).Once()

		err := authSvc.Register("exists@example.com", "password123", "Existing User")
		assert.Equal(t, ErrUserAlreadyExists, err)
		mockRepo.AssertExpectations(t)
	})
}
