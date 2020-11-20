package api

type Permission byte

const (
	AdminPermission = Permission(iota + 1)
	UserPermission
)

type UserOpen struct {
	ID         uint       `json:"id"`
	Login      string     `json:"login"`
	Permission Permission `json:"permission"`
}

type User struct {
	UserOpen
	Password string `json:"password"`
}
