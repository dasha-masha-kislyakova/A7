package auth

type Manager struct {
	ID       int64
	Username string
	Password string // для моков — храним в открытую
	Role     string // "office_manager" | "logistics_manager"
}
