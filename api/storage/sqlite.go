package storage

import (
	"errors"
	"fmt"
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/api/storage/generator"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

func newDatabase(dialect gorm.Dialector) (*Storage, error) {
	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&UserDrawingRelation{}, &User{}, &Drawing{}); err != nil {
		return nil, err
	}
	return &Storage{db}, nil
}

func NewSQLiteStorage(name string) (*Storage, error) {
	return newDatabase(sqlite.Open(name))
}

func (s *Storage) CreateAdmin(force bool) error {
	const (
		lowercaseCount = 10
		uppercaseCount = 6
		digitsCount    = 4
		symbolsCount   = 3
	)
	return s.db.Transaction(func(db *gorm.DB) error {
		if !force {
			var adminsCount int64
			if err := db.Find(&User{}, "permission = ?", api.AdminPermission).Count(&adminsCount).Error; err != nil {
				return err
			}
			if adminsCount > 0 {
				log.Printf("Found %d admins in database", adminsCount)
				return nil
			}
		}
		pass := generator.GeneratePassword(lowercaseCount, uppercaseCount, digitsCount, symbolsCount)

		admin := User{}
		admin.Login = "admin" + fmt.Sprint(time.Now().Unix())
		admin.Password = getHexHash(pass, HashSalt)
		admin.Permission = api.AdminPermission
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("Not found admins. New admin login = %s password = %s", admin.Login, pass)
		return nil
	})
}

func (s *Storage) GetUser(login, pass string) (*api.User, error) {
	pass = getHexHash(pass, HashSalt)
	var user User
	if err := s.db.First(&user, "login = ? and password = ?", login, pass).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToApiUser(nil), nil
}

func (s *Storage) GetUserByID(id uint) (*api.User, error) {
	var user User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrUserNotFound
		}
		return nil, err
	}
	return user.ToApiUser(nil), nil
}

func (s *Storage) CreateUsers(users ...*api.User) error {
	err := s.db.Transaction(func(db *gorm.DB) error {
		for _, u := range users {
			var record User
			if err := db.First(&record, "login = ?", u.Login).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
				if err != nil {
					return err
				}
				return api.ErrUserAlreadyExist
			}
			record = User{User: *u}
			record.Password = getHexHash(record.Password, HashSalt)
			if err := db.Create(&record).Error; err != nil {
				return err
			}
			u.ID, u.Password = record.Model.ID, record.Password
		}
		return nil
	})
	return err
}

func (s *Storage) UpdateUser(user *api.User) error {
	if user.Password != "" {
		user.Password = getHexHash(user.Password, HashSalt)
	}
	record := User{}
	record.UpdateFromApiUser(user)
	if tx := s.db.Model(&record).Where("id = ?", record.User.ID).Updates(record); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrUserNotFound
	}
	return nil
}

func (s *Storage) RemoveUser(id uint) error {
	if tx := s.db.Delete(&User{}, "id = ?", id); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrUserNotFound
	}
	return nil
}

func (s *Storage) GetUsersList(page, pageLimit uint) ([]*api.UserOpen, error) {
	if pageLimit == 0 {
		return make([]*api.UserOpen, 0), nil
	}
	offset := (page - 1) * pageLimit
	users := make([]*User, 0)
	if err := s.db.Offset(int(offset)).Limit(int(pageLimit)).Find(&users).Error; err != nil {
		return nil, err
	}
	usersOpen := make([]*api.UserOpen, len(users))
	for i, u := range users {
		u.UserOpen.ID = u.ID
		usersOpen[i] = &u.UserOpen
	}
	return usersOpen, nil
}

func (s *Storage) UsersAmount() (uint, error) {
	var count int64
	if err := s.db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint(count), nil
}

func (s *Storage) CreateDrawings(userID uint, drawings ...*api.Drawing) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var user User
		if err := db.First(&user, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return api.ErrUserNotFound
			}
			return err
		}
		for _, d := range drawings {
			drawingRecord := Drawing{}
			drawingRecord.UpdateFromApiDrawing(d)
			if err := db.Create(&drawingRecord).Error; err != nil {
				return err
			}
			relation := UserDrawingRelation{User: &user, Drawing: &drawingRecord}
			if err := db.Create(&relation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Storage) GetDrawing(id uint) (*api.Drawing, error) {
	var drawing Drawing
	if err := s.db.First(&drawing, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrDrawingNotFound
		}
		return nil, err
	}
	return drawing.ToApiDrawing(nil), nil
}

func (s *Storage) GetDrawingOfUser(userID, drawingID uint) (*api.Drawing, error) {
	var relation UserDrawingRelation
	err := s.db.Preload("Drawing").First(&relation, "user_id = ? and drawing_id = ?", userID, drawingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.ErrDrawingNotFound
		}
		return nil, err
	}
	return relation.Drawing.ToApiDrawing(nil), nil
}

func (s *Storage) UpdateDrawing(drawing *api.Drawing) error {
	sDrawing := Drawing{}
	sDrawing.UpdateFromApiDrawing(drawing)
	if tx := s.db.Model(&Drawing{}).Where("id = ?", sDrawing.ID).Updates(&sDrawing); tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return api.ErrDrawingNotFound
	}
	return nil
}

func (s *Storage) UpdateDrawingOfUser(userID uint, drawing *api.Drawing) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var relation UserDrawingRelation
		tx := db.Preload("Drawing").First(&relation, "user_id = ? and drawing_id = ?", userID, drawing.ID)
		if tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return api.ErrDrawingNotFound
			}
			return tx.Error
		}
		relation.Drawing.UpdateFromApiDrawing(drawing)
		return db.Model(&Drawing{}).Where("id = ?", drawing.ID).Updates(&relation.Drawing).Error
	})
}

func (s *Storage) RemoveDrawing(id uint) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		if tx := db.Delete(&Drawing{}, "id = ?", id); tx.Error != nil {
			return tx.Error
		} else if tx.RowsAffected == 0 {
			return api.ErrDrawingNotFound
		}
		return db.Delete(&UserDrawingRelation{}, "drawing_id = ?", id).Error
	})
}

func (s *Storage) RemoveDrawingOfUser(userID, drawingID uint) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		tx := db.Delete(&UserDrawingRelation{}, "user_id = ? and drawing_id = ?", userID, drawingID)
		if tx.Error != nil {
			return tx.Error
		}
		if tx.RowsAffected == 0 {
			return api.ErrDrawingNotFound
		}
		if err := db.Delete(&Drawing{}, "id = ?", drawingID).Error; err != nil {
			return err
		}
		return db.Delete(&UserDrawingRelation{}, "drawing_id = ?", drawingID).Error
	})
}

func (s *Storage) GetDrawingsList(userID, page, pageLimit uint) ([]*api.DrawingOpen, error) {
	if pageLimit == 0 {
		return make([]*api.DrawingOpen, 0), nil
	}
	relations := make([]*UserDrawingRelation, 0)
	offset := int((page - 1) * pageLimit)
	err := s.db.Model(&UserDrawingRelation{}).Preload("Drawing").
		Offset(offset).Limit(int(pageLimit)).Find(&relations, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	if len(relations) == 0 {
		return nil, api.ErrDrawingNotFound
	}
	drawings := make([]*api.DrawingOpen, len(relations))
	for i, r := range relations {
		d := r.Drawing.ToApiDrawing(nil)
		drawings[i] = &d.DrawingOpen
	}
	return drawings, nil
}

func (s *Storage) DrawingsAmount(userID uint) (uint, error) {
	var count int64
	if err := s.db.Find(&UserDrawingRelation{}, "user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint(count), nil
}
