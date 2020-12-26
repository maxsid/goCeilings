package api

// Storage interfaces
type UserCreator interface {
	CreateUsers(users ...*UserConfident) error
}

type UsersListGetter interface {
	GetUsersList(page, pageLimit uint) ([]*UserBasic, error)
	UsersAmount() (uint, error)
}

type UserGetter interface {
	GetUser(login, pass string) (*UserConfident, error)
	GetUserByID(id uint) (*UserConfident, error)
}

type UserUpdater interface {
	UpdateUser(user *UserConfident) error
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
	GetDrawingsList(userID, page, pageLimit uint) ([]*DrawingBasic, error)
	DrawingsAmount(userID uint) (uint, error)
}

type DrawingGetter interface {
	GetDrawing(id uint) (*Drawing, error)
}

type DrawingUpdater interface {
	UpdateDrawing(drawing *Drawing) error
}

type DrawingRemover interface {
	RemoveDrawing(id uint) error
}

type DrawingManager interface {
	DrawingCreator
	DrawingGetter
	DrawingUpdater
	DrawingRemover
	DrawingsListGetter
}

type DrawingPermissionCreator interface {
	CreateDrawingPermission(permission *DrawingPermission) error
}

type DrawingPermissionGetter interface {
	GetDrawingPermission(userID, drawingID uint) (*DrawingPermission, error)
	GetDrawingsPermissionsOfDrawing(drawingID uint) ([]*DrawingPermission, error)
	GetDrawingsPermissionsOfUser(userID uint) ([]*DrawingPermission, error)
}

type DrawingPermissionUpdater interface {
	UpdateDrawingPermission(permission *DrawingPermission) error
}

type DrawingPermissionRemover interface {
	RemoveDrawingPermission(userID, drawingID uint) error
}

type DrawingPermissionManager interface {
	DrawingPermissionGetter
	DrawingPermissionCreator
	DrawingPermissionUpdater
	DrawingPermissionRemover
}

// Storage executes all operations with database.
type Storage interface {
	GetUserStorage(user *UserBasic) (UserStorage, error)
	UserManager
	DrawingManager
	DrawingPermissionManager
}

// UserStorage executes operations with database allowed only for the user.
type UserStorage interface {
	GetCurrentUser() *UserBasic
	GetStorage() Storage
	Storage
}
