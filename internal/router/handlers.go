package router

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/your-org/go-starter/internal/business"
	"github.com/your-org/go-starter/internal/models"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// @Summary Create a new user
// @Description Create a new user in the system
// @Tags Users
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User data"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [post]
// CreateUser creates a new user
func CreateUser(c echo.Context, svc business.UserService) error {
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()
	user, err := svc.CreateUser(ctx, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "CREATE_USER_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, user)
}

// @Summary Get user by ID
// @Description Get a specific user by ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [get]
// GetUser retrieves a user by ID
func GetUser(c echo.Context, svc business.UserService) error {
	id := c.Param("id")

	ctx := c.Request().Context()
	user, err := svc.GetUser(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "USER_NOT_FOUND",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

// @Summary Update user
// @Description Update an existing user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateUserRequest true "Updated user data"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [put]
// UpdateUser updates an existing user
func UpdateUser(c echo.Context, svc business.UserService) error {
	id := c.Param("id")

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: err.Error(),
		})
	}

	ctx := c.Request().Context()
	user, err := svc.UpdateUser(ctx, id, &req)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "UPDATE_USER_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func DeleteUser(c echo.Context, svc business.UserService) error {
	id := c.Param("id")

	ctx := c.Request().Context()
	if err := svc.DeleteUser(ctx, id); err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "DELETE_USER_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Delete user
// @Description Delete a user by ID
// @Tags Users
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [delete]

// @Summary List users
// @Description Get a paginated list of users
// @Tags Users
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [get]
// ListUsers lists all users
func ListUsers(c echo.Context, svc business.UserService) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	ctx := c.Request().Context()
	users, err := svc.ListUsers(ctx, limit, offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "LIST_USERS_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, users)
}
