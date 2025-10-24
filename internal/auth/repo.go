package auth

type Repo interface {
	ByUsername(username string) (Manager, bool)
}

type memRepo struct {
	users map[string]Manager
}

func NewMemRepo() Repo {
	return &memRepo{
		users: map[string]Manager{
			"office": {ID: 1, Username: "office", Password: "office", Role: "office_manager"},
			"logi":   {ID: 2, Username: "logi", Password: "logi", Role: "logistics_manager"},
		},
	}
}

func (r *memRepo) ByUsername(username string) (Manager, bool) {
	m, ok := r.users[username]
	return m, ok
}
