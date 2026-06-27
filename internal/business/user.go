package business

import (
	"context"
	"fmt"

	"github.com/CpBruceMeena/go-starter/internal/cache"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/models"
	"github.com/CpBruceMeena/go-starter/internal/repository"
)

// UserServiceInterface defines the user service contract
type UserServiceInterface interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetUser(ctx context.Context, id string) (*UserResponse, error)
	UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*UserResponse, error)
}

// UserService is the exported service interface
type UserService = UserServiceInterface

// Request/Response types (separate from models for clean API boundaries)
type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

type UpdateUserRequest struct {
	Email string `json:"email" validate:"omitempty,email"`
	Name  string `json:"name" validate:"omitempty"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// userService implements UserService
type userService struct {
	repo  repository.UserRepository
	cache *cache.TTLCache
	log   *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, c *cache.TTLCache, log *logger.Logger) UserService {
	return &userService{
		repo:  repo,
		cache: c,
		log:   log,
	}
}

// NewUserServiceWithRepo creates a new user service (accepts nil repo for optional database)
func NewUserServiceWithRepo(repo repository.UserRepository, c *cache.TTLCache, log *logger.Logger) UserService {
	return NewUserService(repo, c, log)
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	s.log.InfoContext(ctx, "creating user", "email", req.Email)

	if s.repo == nil {
		return nil, fmt.Errorf("database not configured")
	}

	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.log.ErrorContext(ctx, "error checking existing user", "error", err.Error())
		return nil, err
	}

	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.log.ErrorContext(ctx, "error creating user", "error", err.Error())
		return nil, err
	}

	s.log.InfoContext(ctx, "user created successfully", "user_id", user.ID)
	if s.cache != nil {
		s.cache.Delete("users:list")
	}

	return s.modelToResponse(user), nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*UserResponse, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database not configured")
	}

	cacheKey := fmt.Sprintf("user:%s", id)

	if s.cache != nil {
		if cached, exists := s.cache.Get(cacheKey); exists {
			s.log.DebugContext(ctx, "user found in cache", "user_id", id)
			return cached.(*UserResponse), nil
		}
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.ErrorContext(ctx, "error getting user", "user_id", id, "error", err.Error())
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	response := s.modelToResponse(user)
	if s.cache != nil {
		s.cache.Set(cacheKey, response, 5*60)
	}

	return response, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database not configured")
	}

	s.log.InfoContext(ctx, "updating user", "user_id", id)

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.ErrorContext(ctx, "error getting user", "user_id", id, "error", err.Error())
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.log.ErrorContext(ctx, "error updating user", "user_id", id, "error", err.Error())
		return nil, err
	}

	s.log.InfoContext(ctx, "user updated successfully", "user_id", id)

	cacheKey := fmt.Sprintf("user:%s", id)
	if s.cache != nil {
		s.cache.Delete(cacheKey)
		s.cache.Delete("users:list")
	}

	return s.modelToResponse(user), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if s.repo == nil {
		return fmt.Errorf("database not configured")
	}

	s.log.InfoContext(ctx, "deleting user", "user_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.ErrorContext(ctx, "error deleting user", "user_id", id, "error", err.Error())
		return err
	}

	s.log.InfoContext(ctx, "user deleted successfully", "user_id", id)

	cacheKey := fmt.Sprintf("user:%s", id)
	if s.cache != nil {
		s.cache.Delete(cacheKey)
		s.cache.Delete("users:list")
	}

	return nil
}

// ListUsers lists all users
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*UserResponse, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database not configured")
	}

	cacheKey := fmt.Sprintf("users:list:%d:%d", limit, offset)

	if s.cache != nil {
		if cached, exists := s.cache.Get(cacheKey); exists {
			s.log.DebugContext(ctx, "users list found in cache")
			return cached.([]*UserResponse), nil
		}
	}

	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "error listing users", "error", err.Error())
		return nil, err
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.modelToResponse(user)
	}

	if s.cache != nil {
		s.cache.Set(cacheKey, responses, 5*60)
	}

	return responses, nil
}

// modelToResponse converts models.User to UserResponse
func (s *userService) modelToResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
