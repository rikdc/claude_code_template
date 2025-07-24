package handlers

// HookData represents the structure of hook data from Claude Code
type HookData struct {
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	SessionID string                 `json:"session_id"`
	Filename  string                 `json:"filename"`
	Data      map[string]interface{} `json:"data"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}