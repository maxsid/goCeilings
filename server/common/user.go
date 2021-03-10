package common

type UserRole byte

const (
	RoleAdmin = UserRole(iota + 1)
	RoleUser
)

// UserBasic contains basic public information of the user.
type UserBasic struct {
	ID    uint     `json:"id"`
	Login string   `json:"login"`
	Role  UserRole `json:"role"`
}

// UserConfident contains UserBasic and confident user information.
type UserConfident struct {
	UserBasic
	Password string `json:"password"`
}
