package auth

type User struct {
	ID       int
	Email    string
	Password string
	Role     string // office_admin | logpoint_admin
}
