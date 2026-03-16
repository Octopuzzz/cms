package contextutils

type ContextKey string

const (
	ContextUserID       ContextKey = "user_id"
	ContextUsername     ContextKey = "username"
	ContextRoles        ContextKey = "roles"
	ContextIsSuperAdmin ContextKey = "is_super_admin"
	ContextClaims       ContextKey = "claims"
)
