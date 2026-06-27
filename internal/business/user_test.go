package business

import (
	"context"
	"testing"

	"github.com/CpBruceMeena/go-starter/internal/cache"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/models"
)

// MockUserRepository implements repository.UserRepository for testing
type MockUserRepository struct {
	CreateFunc    func(ctx context.Context, user *models.User) error
	GetByIDFunc   func(ctx context.Context, id string) (*models.User, error)
	GetByEmailFunc func(ctx context.Context, email string) (*models.User, error)
	UpdateFunc    func(ctx context.Context, user *models.User) error
	DeleteFunc    func(ctx context.Context, id string) error
	ListFunc      func(ctx context.Context, limit, offset int) ([]*models.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return []*models.User{}, nil
}

// MockLogger implements logger.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, args ...any)                              {}
func (m *MockLogger) Error(msg string, args ...any)                             {}
func (m *MockLogger) Warn(msg string, args ...any)                              {}
func (m *MockLogger) Debug(msg string, args ...any)                             {}
func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...any) {}
func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...any)    {}
func (m *MockLogger) DebugContext(ctx context.Context, msg string, args ...any)  {}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateUserRequest
		setup   func(*MockUserRepository)
		wantErr bool
	}{
		{
			name: "valid user creation",
			req: &CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setup: func(m *MockUserRepository) {
				m.GetByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
					return nil, nil
				}
				m.CreateFunc = func(ctx context.Context, user *models.User) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			req: &CreateUserRequest{
				Email: "existing@example.com",
				Name:  "Test User",
			},
			setup: func(m *MockUserRepository) {
				m.GetByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
					return &models.User{ID: "1", Email: email}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := cache.NewTTLCache()
			mockLog := &MockLogger{}
			tt.setup(mockRepo)

			service := NewUserService(mockRepo, mockCache, mockLog)
			_, err := service.CreateUser(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(*MockUserRepository)
		wantErr bool
	}{
		{
			name: "user found",
			id:   "1",
			setup: func(m *MockUserRepository) {
				m.GetByIDFunc = func(ctx context.Context, id string) (*models.User, error) {
					return &models.User{ID: id, Email: "test@example.com", Name: "Test"}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "999",
			setup: func(m *MockUserRepository) {
				m.GetByIDFunc = func(ctx context.Context, id string) (*models.User, error) {
					return nil, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := cache.NewTTLCache()
			mockLog := &MockLogger{}
			tt.setup(mockRepo)

			service := NewUserService(mockRepo, mockCache, mockLog)
			_, err := service.GetUser(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}