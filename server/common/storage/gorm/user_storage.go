package gorm

import (
	"errors"
	"fmt"

	"github.com/maxsid/goCeilings/server/api"
	"github.com/maxsid/goCeilings/server/common"
)

type UserStorage struct {
	*Storage
	user *common.UserBasic
}

func (u *UserStorage) checkPermission(drawingID uint, f func(p *common.DrawingPermission) bool) error {
	if u.user.Role == common.RoleAdmin {
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

func (u *UserStorage) GetCurrentUser() *common.UserBasic {
	return u.user
}

func (u *UserStorage) GetStorage() common.Storage {
	return u.Storage
}

func (u *UserStorage) GetUserStorage(user *common.UserBasic) (common.UserStorage, error) {
	if u.user.Role != common.RoleAdmin {
		return nil, fmt.Errorf("%w: only for admins", api.ErrOperationNotAllowed)
	}
	return u.Storage.GetUserStorage(user)
}

func (u *UserStorage) GetUser(_, _ string) (*common.UserConfident, error) {
	return nil, fmt.Errorf("%w: get user by login and password", api.ErrOperationNotAllowed)
}

func (u *UserStorage) GetUserByID(id uint) (*common.UserConfident, error) {
	if u.user.Role == common.RoleAdmin || u.user.ID == id {
		return u.Storage.GetUserByID(id)
	}
	return nil, api.ErrOperationNotAllowed
}

func (u *UserStorage) CreateUsers(users ...*common.UserConfident) error {
	if u.user.Role != common.RoleAdmin {
		return api.ErrOperationNotAllowed
	}
	return u.Storage.CreateUsers(users...)
}

func (u *UserStorage) UpdateUser(user *common.UserConfident) error {
	if u.user.Role == common.RoleAdmin {
		return u.Storage.UpdateUser(user)
	}
	if user.ID == u.user.ID {
		user.Role = u.user.Role
		return u.Storage.UpdateUser(user)
	}
	return api.ErrOperationNotAllowed
}

func (u *UserStorage) RemoveUser(id uint) error {
	if u.user.Role != common.RoleAdmin {
		return api.ErrOperationNotAllowed
	}
	return u.Storage.RemoveUser(id)
}

func (u *UserStorage) GetUsersList(page, pageLimit uint) ([]*common.UserBasic, error) {
	if u.user.Role != common.RoleAdmin {
		return nil, api.ErrOperationNotAllowed
	}
	return u.Storage.GetUsersList(page, pageLimit)
}

func (u *UserStorage) UsersAmount() (uint, error) {
	if u.user.Role != common.RoleAdmin {
		return 0, api.ErrOperationNotAllowed
	}
	return u.Storage.UsersAmount()
}

func (u *UserStorage) CreateDrawings(userID uint, drawings ...*common.Drawing) error {
	if u.user.Role != common.RoleAdmin && u.user.ID != userID {
		return api.ErrOperationNotAllowed
	}
	return u.Storage.CreateDrawings(userID, drawings...)
}

func (u *UserStorage) GetDrawing(id uint) (*common.Drawing, error) {
	if u.user.Role == common.RoleAdmin {
		return u.Storage.GetDrawing(id)
	}
	return u.getDrawingOfUser(u.user.ID, id)
}

func (u *UserStorage) getDrawingOfUser(userID, drawingID uint) (*common.Drawing, error) {
	if err := u.checkPermission(drawingID, func(p *common.DrawingPermission) bool { return p.Get || p.Owner }); err != nil {
		return nil, err
	}
	return u.Storage.getDrawingOfUser(userID, drawingID)
}

func (u *UserStorage) UpdateDrawing(drawing *common.Drawing) error {
	if u.user.Role == common.RoleAdmin {
		return u.Storage.UpdateDrawing(drawing)
	}
	return u.updateDrawingOfUser(u.user.ID, drawing)
}

func (u *UserStorage) updateDrawingOfUser(userID uint, drawing *common.Drawing) error {
	if err := u.checkPermission(drawing.ID, func(p *common.DrawingPermission) bool { return p.Change || p.Owner }); err != nil {
		return err
	}
	return u.Storage.updateDrawingOfUser(userID, drawing)
}

func (u *UserStorage) RemoveDrawing(id uint) error {
	if u.user.Role == common.RoleAdmin {
		return u.Storage.RemoveDrawing(id)
	}
	return u.removeDrawingOfUser(u.user.ID, id)
}

func (u *UserStorage) removeDrawingOfUser(userID, drawingID uint) error {
	if err := u.checkPermission(drawingID, func(p *common.DrawingPermission) bool { return p.Delete || p.Owner }); err != nil {
		return err
	}
	return u.Storage.removeDrawingOfUser(userID, drawingID)
}

func (u *UserStorage) GetDrawingsList(userID, page, pageLimit uint) ([]*common.DrawingBasic, error) {
	if u.user.Role == common.RoleAdmin || u.user.ID == userID {
		return u.Storage.GetDrawingsList(userID, page, pageLimit)
	}
	return nil, api.ErrOperationNotAllowed
}

func (u *UserStorage) DrawingsAmount(userID uint) (uint, error) {
	if u.user.Role == common.RoleAdmin || u.user.ID == userID {
		return u.Storage.DrawingsAmount(userID)
	}
	return 0, api.ErrOperationNotAllowed
}

func (u *UserStorage) GetDrawingPermission(userID, drawingID uint) (*common.DrawingPermission, error) {
	if err := u.checkPermission(drawingID, func(p *common.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return nil, err
	}
	return u.Storage.GetDrawingPermission(userID, drawingID)
}

func (u *UserStorage) GetDrawingsPermissionsOfDrawing(drawingID uint) ([]*common.DrawingPermission, error) {
	if err := u.checkPermission(drawingID, func(p *common.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return nil, err
	}
	return u.Storage.GetDrawingsPermissionsOfDrawing(drawingID)
}

func (u *UserStorage) GetDrawingsPermissionsOfUser(userID uint) ([]*common.DrawingPermission, error) {
	if u.user.Role == common.RoleAdmin || userID == u.user.ID {
		return u.Storage.GetDrawingsPermissionsOfUser(userID)
	}
	return nil, api.ErrOperationNotAllowed
}

func (u *UserStorage) CreateDrawingPermission(permission *common.DrawingPermission) error {
	if err := u.checkPermission(permission.Drawing.ID, func(p *common.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return err
	}
	return u.Storage.CreateDrawingPermission(permission)
}

func (u *UserStorage) UpdateDrawingPermission(permission *common.DrawingPermission) error {
	if err := u.checkPermission(permission.Drawing.ID, func(p *common.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return err
	}
	return u.Storage.UpdateDrawingPermission(permission)
}

func (u *UserStorage) RemoveDrawingPermission(userID, drawingID uint) error {
	if err := u.checkPermission(drawingID, func(p *common.DrawingPermission) bool { return p.Share || p.Owner }); err != nil {
		return err
	}
	return u.Storage.RemoveDrawingPermission(userID, drawingID)
}
