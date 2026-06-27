package response

// APIResponse is the standard response wrapper for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents an error in API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta contains metadata for paginated responses
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Success returns a success response
func Success(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

// SuccessWithMeta returns a success response with pagination metadata
func SuccessWithMeta(data interface{}, page, perPage, total int) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: (total + perPage - 1) / perPage,
		},
	}
}

// Error returns an error response
func Error(code, message string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}

// List returns a paginated list response
type List struct {
	Items interface{} `json:"items"`
	Meta  *Meta       `json:"meta"`
}
