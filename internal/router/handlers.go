package router

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/CpBruceMeena/go-starter/internal/business"
	"github.com/CpBruceMeena/go-starter/internal/models"
	"github.com/CpBruceMeena/go-starter/internal/response"
)

// CreateUser creates a new user
func CreateUser(c echo.Context, svc business.UserService) error {
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("INVALID_REQUEST", err.Error()))
	}

	user, err := svc.CreateUser(c.Request().Context(), &business.CreateUserRequest{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("CREATE_USER_ERROR", err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(user))
}

// GetUser retrieves a user by ID
func GetUser(c echo.Context, svc business.UserService) error {
	id := c.Param("id")

	user, err := svc.GetUser(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("USER_NOT_FOUND", err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(user))
}

// UpdateUser updates an existing user
func UpdateUser(c echo.Context, svc business.UserService) error {
	id := c.Param("id")

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("INVALID_REQUEST", err.Error()))
	}

	user, err := svc.UpdateUser(c.Request().Context(), id, &business.UpdateUserRequest{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("UPDATE_USER_ERROR", err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(user))
}

// DeleteUser deletes a user
func DeleteUser(c echo.Context, svc business.UserService) error {
	id := c.Param("id")

	if err := svc.DeleteUser(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusNotFound, response.Error("DELETE_USER_ERROR", err.Error()))
	}

	return c.NoContent(http.StatusNoContent)
}

// ListUsers lists all users
func ListUsers(c echo.Context, svc business.UserService) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, err := svc.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("LIST_USERS_ERROR", err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(users))
}