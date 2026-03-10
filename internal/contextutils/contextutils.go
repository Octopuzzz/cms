package contextutils

// ContextKey is a custom type for context keys to prevent collisions
type ContextKey string

const (
	// UserIDKey is the context key for storing the authenticated user's ID
	UserIDKey ContextKey = "userID"
)
