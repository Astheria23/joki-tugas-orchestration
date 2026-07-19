package contracts

// APIResponse represents the standardized JSON structure for all API endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// APIError represents structured error detail returned in the APIResponse.
type APIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}
