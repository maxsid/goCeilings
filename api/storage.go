package api

// Storage interfaces
type UserCreator interface {
	CreateUsers(users ...*User) error
}

type UsersListGetter interface {
	GetUsersList(page, pageLimit uint) ([]*UserOpen, error)
	UsersAmount() (uint, error)
}

type UserGetter interface {
	GetUser(login, pass string) (*User, error)
	GetUserByID(id uint) (*User, error)
}

type UserUpdater interface {
	UpdateUser(user *User) error
}

type UserRemover interface {
	RemoveUser(id uint) error
}

type UserManager interface {
	UserGetter
	UserCreator
	UserUpdater
	UserRemover
	UsersListGetter
}

type DrawingCreator interface {
	CreateDrawings(userID uint, drawings ...*Drawing) error
}

type DrawingsListGetter interface {
	GetDrawingsList(userID, page, pageLimit uint) ([]*DrawingOpen, error)
	DrawingsAmount(userID uint) (uint, error)
}

type DrawingGetter interface {
	GetDrawing(id uint) (*Drawing, error)
	GetDrawingOfUser(userID, drawingID uint) (*Drawing, error)
}

type DrawingUpdater interface {
	UpdateDrawing(drawing *Drawing) error
	UpdateDrawingOfUser(userID uint, drawing *Drawing) error
}

type DrawingRemover interface {
	RemoveDrawing(id uint) error
	RemoveDrawingOfUser(userID, drawingID uint) error
}

type DrawingManager interface {
	DrawingCreator
	DrawingGetter
	DrawingUpdater
	DrawingRemover
	DrawingsListGetter
}

type Storage interface {
	UserManager
	DrawingManager
}
