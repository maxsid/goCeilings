package storage

import (
	"errors"
	"fmt"
	"github.com/maxsid/goCeilings/api"
)

type UserStorage struct {
	*Storage
	user *api.UserBasic
}

func (u *UserStorage) checkPermission(drawingID uint, f func(p *api.DrawingPermission) bool) error {
	if u.user.Role == api.RoleAdmin {
		return nil
	}
	p, err := u.Storage.GetDrawingPermission(u.user.ID, drawingID)
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			return api.ErrOperationNotAllowed
		}
		return err
	}
	if !f(p) {
		return api.ErrOperationNotAllowed
	}
	return nil
}

func (u *UserStorage) GetCurrentUser() *api.UserBasic {
	return u.user
}

func (u *UserStorage) GetStorage() api.Storage {
	return u.Storage
}

func (u *UserStorage) GetUserStorage(user *api.UserBasic) (api.UserStorage, error) {
	if u.user.Role != api.RoleAdmin {
		return nil, fmt.Errorf("%w: only for admins", api.ErrOperationNotAllowed)
	}
	return u.Storage.GetUserStorage(user)
}

func (u *UserStorage) GetUser(_, _ string) (*api.UserConfident, error) {
	return nil, fmt.Errorf("%w: get user by login and password", api.ErrOperationNotAllowed)
}

func (u *UserStorage) GetUserByID(id uint) (*api.UserConfident, error) {
	if u.user.Role == api.RoleAdmin || u.user.ID == id {
		return u.Storage.GetUserByID(id)
	}
	return nil, api.ErrOperationNotAllowed
}

func (u *UserStorage) CreateUsers(users ...*api.UserConfident) error {
	if u.user.Role != api.RoleAdmin {
		return api.ErrOperationNotAllowed
	}
	return u.Storage.CreateUsers(users...)
}

func (u *UserStorage) UpdateUser(user *api.UserConfident) error {
	if u.user.Role == api.RoleAdmin {
		return u.Storage.UpdateUser(user)
	}
	if user.ID == u.user.ID {
		user.Role = u.user.Role
		return u.Storage.UpdateUser(user)
	}
	return api.ErrOperationNotAllowed
}

func (u *UserStorage) RemoveUser(id uint) error {
	if u.user.Role != api.RoleAdmin {
		return api.ErrOperationNotAllowed
	}
	return u.Storage.RemoveUser(id)
}

func (u *UserStorage) GetUsersList(page, pageLimit uint) ([]*api.UserBasic, error) {
	if u.user.Role != api.RoleAdmin {
		return nil, api.ErrOperationNotAllowed
	}
	return u.Storage.GetUsersList(page, pageLimit)
}

func (u *UserStorage) UsersAmount() (uint, error) {
	if u.user.Role != api.RoleAdmin {
		return 0, api.ErrOperationNotAllowed
	}
	return u.Storage.UsersAmount()
}

func (u *UserStorage) CreateDrawings(userID uint, drawings ...*api.Drawing) error {
	if u.user.Role != api.RoleAdmin && u.user.ID != userID {
		return api.ErrOperationNotAllowed
	}
	return u.Storage.CreateDrawings(userID, drawings...)
}

func (u *UserStorage) GetDrawing(id uint) (*api.Drawing, error) {
	if u.user.Role == api.RoleAdmin {
		return u.Storage.GetDrawing(id)
	}
	return u.getDrawingOfUser(u.user.ID, id)
}

func (u *UserStorage) getDrawingOfUser(userID, drawingID uint) (*api.Drawing, error) {
	if err := u.checkPermission(drawingID, func(p *api.DrawingPermission) bool { return p.Get || p.Owner }); err != nil {
		return nil, err
	}
	return u.Storage.getDrawingOfUser(userID, drawingID)
}

func (u *UserStorage) UpdateDrawing(drawing *api.Drawing) error {
	if u.user.Role == api.RoleAdmin {
		return u.Storage.UpdateDrawing(drawing)
	}
	return u.updateDrawingOfUser(u.user.ID, drawing)
}

func (u *UserStorage) updateDrawingOfUser(userID uint, drawing *api.Drawing) error {
	if err := u.checkPermission(drawing.ID, func(p *api.DrawingPermission) bool { return p.Change || p.Owner }); err != nil {
		return err
	}
	return u.Storage.updateDrawingOfUser(userID, drawing)
}

func (u *UserStorage) RemoveDrawing(id uint) error {
	if u.user.Role == api.RoleAdmin {
		return u.Storage.RemoveDrawing(id)
	}
	return u.removeDrawingOfUser(u.user.ID, id)
}

func (u *UserStorage) removeDrawingOfUser(userID, drawingID uint) error {
	if err := u.checkPermission(drawingID, func(p *api.DrawingPermission) bool { return p.Delete || p.Owner }); err != nil {
		return err
	}
	return u.Storage.removeDrawingOfUser(userID, drawingID)
}

func (u *UserStorage) GetDrawingsList(userID, page, pageLimit uint) ([]*api.DrawingBasic, error) {
	if u.user.Role == api.RoleAdmin || u.user.ID == userID {
		return u.Storage.GetDrawingsList(userID, page, pageLimit)
	}
	return nil, api.ErrOperationNotAllowed
}

func (u *UserStorage) DrawingsAmount(userID uint) (uint, error) {
	if u.user.Role == api.RoleAdmin || u.user.ID == userID {
		return u.Storage.DrawingsAmount(userID)
	}
	return 0, api.ErrOperationNotAllowed
}

func (u *UserStorage) GetDrawingPermission(userID, drawingID uint) (*api.DrawingPermission, error) {
	if err := u.checkPermission(drawingID, func(p *api.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return nil, err
	}
	return u.Storage.GetDrawingPermission(userID, drawingID)
}

func (u *UserStorage) GetDrawingsPermissionsOfDrawing(drawingID uint) ([]*api.DrawingPermission, error) {
	if err := u.checkPermission(drawingID, func(p *api.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return nil, err
	}
	return u.Storage.GetDrawingsPermissionsOfDrawing(drawingID)
}

func (u *UserStorage) GetDrawingsPermissionsOfUser(userID uint) ([]*api.DrawingPermission, error) {
	if u.user.Role == api.RoleAdmin || userID == u.user.ID {
		return u.Storage.GetDrawingsPermissionsOfUser(userID)
	}
	return nil, api.ErrOperationNotAllowed
}

func (u *UserStorage) CreateDrawingPermission(permission *api.DrawingPermission) error {
	if err := u.checkPermission(permission.Drawing.ID, func(p *api.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return err
	}
	return u.Storage.CreateDrawingPermission(permission)
}

func (u *UserStorage) UpdateDrawingPermission(permission *api.DrawingPermission) error {
	if err := u.checkPermission(permission.Drawing.ID, func(p *api.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return err
	}
	return u.Storage.UpdateDrawingPermission(permission)
}

func (u *UserStorage) RemoveDrawingPermission(userID, drawingID uint) error {
	if err := u.checkPermission(drawingID, func(p *api.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return err
	}
	return u.Storage.RemoveDrawingPermission(userID, drawingID)
}
