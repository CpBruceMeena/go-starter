package business

import (
	"context"
	"fmt"

	"github.com/CpBruceMeena/go-starter/internal/cache"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/models"
	"github.com/CpBruceMeena/go-starter/internal/repository"
)

// UserService defines user business logic
type UserService interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error)
	GetUser(ctx context.Context, id string) (*models.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.UserResponse, error)
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

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	s.log.InfoContext(ctx, "creating user", "email", req.Email)

	// Check if user already exists
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.log.ErrorContext(ctx, "error checking existing user", "error", err.Error())
		return nil, err
	}

	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create user
	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.log.ErrorContext(ctx, "error creating user", "error", err.Error())
		return nil, err
	}

	s.log.InfoContext(ctx, "user created successfully", "user_id", user.ID)

	// Invalidate list cache
	s.cache.Delete("users:list")

	return user.ToResponse(), nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*models.UserResponse, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	// Try to get from cache
	if cached, exists := s.cache.Get(cacheKey); exists {
		s.log.DebugContext(ctx, "user found in cache", "user_id", id)
		return cached.(*models.UserResponse), nil
	}

	// Get from database
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.ErrorContext(ctx, "error getting user", "user_id", id, "error", err.Error())
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	response := user.ToResponse()

	// Cache the result (5 minute TTL)
	s.cache.Set(cacheKey, response, 5*60)

	return response, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	s.log.InfoContext(ctx, "updating user", "user_id", id)

	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.ErrorContext(ctx, "error getting user", "user_id", id, "error", err.Error())
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}

	// Save to database
	if err := s.repo.Update(ctx, user); err != nil {
		s.log.ErrorContext(ctx, "error updating user", "user_id", id, "error", err.Error())
		return nil, err
	}

	s.log.InfoContext(ctx, "user updated successfully", "user_id", id)

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", id)
	s.cache.Delete(cacheKey)
	s.cache.Delete("users:list")

	return user.ToResponse(), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	s.log.InfoContext(ctx, "deleting user", "user_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.ErrorContext(ctx, "error deleting user", "user_id", id, "error", err.Error())
		return err
	}

	s.log.InfoContext(ctx, "user deleted successfully", "user_id", id)

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", id)
	s.cache.Delete(cacheKey)
	s.cache.Delete("users:list")

	return nil
}

// ListUsers lists all users
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*models.UserResponse, error) {
	cacheKey := fmt.Sprintf("users:list:%d:%d", limit, offset)

	// Try to get from cache
	if cached, exists := s.cache.Get(cacheKey); exists {
		s.log.DebugContext(ctx, "users list found in cache")
		return cached.([]*models.UserResponse), nil
	}

	// Get from database
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "error listing users", "error", err.Error())
		return nil, err
	}

	// Convert to responses
	responses := make([]*models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	// Cache the result (5 minute TTL)
	s.cache.Set(cacheKey, responses, 5*60)

	return responses, nil
}
