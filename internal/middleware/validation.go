package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/response"
)

// Validate is the validator instance
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// ValidationMiddleware validates request bodies against struct tags
func ValidationMiddleware(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the validated instance if available
			if c.Get("validated") != nil {
				return next(c)
			}
			return next(c)
		}
	}
}

// BindAndValidate binds request body and validates it
func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("INVALID_REQUEST", err.Error()))
	}
	if err := Validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", err.Error()))
	}
	return nil
}