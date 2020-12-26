package storage

import (
	"errors"
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/api/storage/generator"
	"gorm.io/gorm"
	"log"
)

type Storage struct {
	db *gorm.DB
}

func newDatabase(dialect gorm.Dialector) (st *Storage, err error) {
	st = &Storage{}
	st.db, err = gorm.Open(dialect, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return
	}
	if err = st.db.AutoMigrate(&drawingPermissionModel{}, &userModel{}, &drawingModel{}); err != nil {
		return
	}
	return
}

func (s *Storage) GetUserStorage(user *api.UserBasic) (api.UserStorage, error) {
	return &UserStorage{
		Storage: s,
		user:    user,
	}, nil
}

func (s *Storage) CreateAdmin(force bool) error {
	const (
		lowercaseCount = 10
		uppercaseCount = 6
		digitsCount    = 4
		symbolsCount   = 3
	)
	if !force {
		var adminsCount int64
		if err := s.db.Find(&userModel{}, "role = ?", api.RoleAdmin).Count(&adminsCount).Error; err != nil {
			return err
		}
		if adminsCount > 0 {
			log.Printf("Found %d admins in database", adminsCount)
			return nil
		}
	}
	pass := generator.GeneratePassword(lowercaseCount, uppercaseCount, digitsCount, symbolsCount)

	admin := userModel{}
	admin.Login = "admin-" + generator.GeneratePassword(8, 0, 0, 0)
	admin.Password = getHexHash(pass, HashSalt)
	admin.Role = api.RoleAdmin
	if err := s.db.Create(&admin).Error; err != nil {
		return err
	}
	log.Printf("Not found admins. New admin login = %s password = %s", admin.Login, pass)
	return nil
}

func (s *Storage) GetUser(login, pass string) (*api.UserConfident, error) {
	pass = getHexHash(pass, HashSalt)
	var user userModel
	if err := s.db.First(&user, "login = ? and password = ?", login, pass).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToApi(), nil
}

func (s *Storage) GetUserByID(id uint) (*api.UserConfident, error) {
	var user userModel
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToApi(), nil
}

func (s *Storage) CreateUsers(users ...*api.UserConfident) error {
	records := make([]*userModel, len(users))
	for i, u := range users {
		ru := &userModel{}
		records[i] = ru
		ru.UpdateFromApi(u)
		ru.Password = getHexHash(ru.Password, HashSalt)
	}
	if err := s.db.Create(&records).Error; err != nil {
		return err
	}
	return nil
}

func (s *Storage) UpdateUser(user *api.UserConfident) error {
	if user.Password != "" {
		user.Password = getHexHash(user.Password, HashSalt)
	}
	record := userModel{}
	record.UpdateFromApi(user)
	if tx := s.db.Model(&record).Where("id = ?", user.ID).Updates(record); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrUserNotFound
	}
	return nil
}

func (s *Storage) RemoveUser(id uint) error {
	if tx := s.db.Delete(&userModel{}, "id = ?", id); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrUserNotFound
	}
	return nil
}

func (s *Storage) GetUsersList(page, pageLimit uint) ([]*api.UserBasic, error) {
	if pageLimit == 0 {
		return make([]*api.UserBasic, 0), nil
	}
	offset := (page - 1) * pageLimit
	rows, err := s.db.Model(&userModel{}).Offset(int(offset)).Limit(int(pageLimit)).Order("id asc").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	user, out := new(userModel), make([]*api.UserBasic, 0)
	for rows.Next() {
		if err := s.db.ScanRows(rows, user); err != nil {
			return nil, err
		}
		out = append(out, &user.ToApi().UserBasic)
	}
	return out, nil
}

func (s *Storage) UsersAmount() (uint, error) {
	var count int64
	if err := s.db.Model(&userModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint(count), nil
}

func (s *Storage) CreateDrawings(userID uint, drawings ...*api.Drawing) error {
	rDrawings, rDrawingPermissions := make([]*drawingModel, len(drawings)), make([]*drawingPermissionModel, len(drawings))
	for i, d := range drawings {
		dRecord := &drawingModel{}
		dRecord.UpdateFromApi(d)
		rDrawings[i] = dRecord
	}
	return s.db.Transaction(func(db *gorm.DB) error {
		if err := db.Create(&rDrawings).Error; err != nil {
			return err
		}
		for i, d := range rDrawings {
			rDrawingPermissions[i] = &drawingPermissionModel{UserID: userID, DrawingID: d.Model.ID, Owner: true}
		}
		if err := db.Create(&rDrawingPermissions).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Storage) GetDrawing(id uint) (*api.Drawing, error) {
	var drawing drawingModel
	if err := s.db.First(&drawing, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrDrawingNotFound
		}
		return nil, err
	}
	return drawing.ToApi(), nil
}

func (s *Storage) getDrawingOfUser(userID, drawingID uint) (*api.Drawing, error) {
	var permission drawingPermissionModel
	err := s.db.Preload("Drawing").First(&permission, "drawing_id = ? AND user_id = ?", drawingID, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrDrawingNotFound
		}
		return nil, err
	}
	return permission.Drawing.ToApi(), nil
}

func (s *Storage) UpdateDrawing(drawing *api.Drawing) error {
	sDrawing := drawingModel{}
	sDrawing.UpdateFromApi(drawing)
	if tx := s.db.Model(&drawingModel{}).Where("id = ?", sDrawing.ID).Updates(&sDrawing); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrDrawingNotFound
	}
	return nil
}

func (s *Storage) updateDrawingOfUser(userID uint, drawing *api.Drawing) error {
	if err := s.db.First(&drawingPermissionModel{}, "drawing_id = ? AND user_id = ?", drawing.ID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.ErrDrawingNotFound
		}
		return err
	}
	d := drawingModel{}
	d.UpdateFromApi(drawing)
	if tx := s.db.Model(d).Where("id = ?", drawing.ID).Updates(d); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrDrawingNotFound
	}
	return nil
}

func (s *Storage) RemoveDrawing(id uint) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		if tx := db.Delete(&drawingModel{}, "id = ?", id); tx.Error != nil {
			return tx.Error
		} else if tx.RowsAffected == 0 {
			return api.ErrDrawingNotFound
		}
		if tx := db.Delete(&drawingPermissionModel{}, "drawing_id = ?", id); tx.Error != nil {
			return tx.Error
		} else if tx.RowsAffected == 0 {
			return api.ErrDrawingNotFound
		}
		return nil
	})
}

func (s *Storage) removeDrawingOfUser(userID, drawingID uint) error {
	if tx := s.db.Select("drawing_id").First(&drawingPermissionModel{}, "drawing_id = ? AND user_id = ?", drawingID, userID); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return api.ErrDrawingNotFound
		}
	}
	return s.RemoveDrawing(drawingID)
}

func (s *Storage) GetDrawingsList(userID, page, pageLimit uint) ([]*api.DrawingBasic, error) {
	if pageLimit == 0 {
		return make([]*api.DrawingBasic, 0), nil
	}
	offset := int((page - 1) * pageLimit)
	var permissions []*drawingPermissionModel
	err := s.db.Preload("Drawing").Offset(offset).Limit(int(pageLimit)).Order("created_at desc").Find(&permissions, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	} else if permissions == nil || len(permissions) == 0 {
		return nil, api.ErrDrawingNotFound
	}
	out := make([]*api.DrawingBasic, len(permissions))
	for i, p := range permissions {
		out[i] = &p.Drawing.ToApi().DrawingBasic
	}
	return out, nil
}

func (s *Storage) DrawingsAmount(userID uint) (uint, error) {
	var count int64
	if err := s.db.Find(&drawingPermissionModel{}, "user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint(count), nil
}

func (s *Storage) GetDrawingPermission(userID, drawingID uint) (*api.DrawingPermission, error) {
	var permission drawingPermissionModel
	tx := s.db.Preload("User").Preload("Drawing").First(&permission, "user_id = ? AND drawing_id = ?", userID, drawingID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, api.ErrNotFound
		}
		return nil, tx.Error
	}
	return permission.ToApi(), nil
}

func (s *Storage) GetDrawingsPermissionsOfDrawing(drawingID uint) ([]*api.DrawingPermission, error) {
	var permissions []*drawingPermissionModel
	err := s.db.Preload("User").Preload("Drawing").Order("created_at DESC").Find(&permissions, "drawing_id = ?", drawingID).Error
	if err != nil {
		return nil, err
	}
	out := make([]*api.DrawingPermission, len(permissions))
	for i, p := range permissions {
		out[i] = p.ToApi()
	}
	return out, nil
}

func (s *Storage) GetDrawingsPermissionsOfUser(userID uint) ([]*api.DrawingPermission, error) {
	var permissions []*drawingPermissionModel
	err := s.db.Preload("User").Preload("Drawing").Order("created_at DESC").Find(&permissions, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	out := make([]*api.DrawingPermission, len(permissions))
	for i, p := range permissions {
		out[i] = p.ToApi()
	}
	return out, nil
}

func (s *Storage) CreateDrawingPermission(permission *api.DrawingPermission) error {
	p := drawingPermissionModel{}
	p.UpdateFromApi(permission)
	p.Owner = false
	if !p.IsFullFalse() {
		if err := s.db.Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) UpdateDrawingPermission(permission *api.DrawingPermission) error {
	p := drawingPermissionModel{}
	p.UpdateFromApi(permission)
	p.Owner = false
	tx := s.db.Model(&p).Select("Get", "Change", "Delete", "Share").Where("user_id = ? AND drawing_id = ? AND owner = ?", p.UserID, p.DrawingID, false).Updates(&p)
	if tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrNotFound
	}
	if p.IsFullFalse() {
		return s.RemoveDrawingPermission(p.UserID, p.DrawingID)
	}
	return nil
}

func (s *Storage) RemoveDrawingPermission(userID, drawingID uint) error {
	if tx := s.db.Delete(&drawingPermissionModel{}, "user_id = ? AND drawing_id = ? AND owner = ?", userID, drawingID, false); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrNotFound
	}
	return nil
}
