package storage

import (
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/drawing"
	"github.com/maxsid/goCeilings/drawing/raster"
	"gorm.io/gorm"
	"time"
)

type Storage struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	api.User
}

type Drawing struct {
	gorm.Model
	Name                           string
	Description                    drawing.Description
	Area, Perimeter, Height, Width float64
	Drawing                        *raster.GGDrawing
}

type UserDrawingRelation struct {
	UserID    uint     `gorm:"primaryKey;autoIncrement:false"`
	DrawingID uint     `gorm:"primaryKey;autoIncrement:false"`
	User      *User    `gorm:"foreignKey:UserID"`
	Drawing   *Drawing `gorm:"foreignKey:DrawingID"`
	CreatedAt time.Time
}

func (d *Drawing) ToApiDrawing(ad *api.Drawing) *api.Drawing {
	if ad == nil {
		ad = &api.Drawing{}
	}
	ad.ID = d.ID
	ad.Name = d.Name
	ad.GGDrawing = *d.Drawing
	ad.GGDrawing.Description = d.Description
	if ad.GGDrawing.Description == nil {
		ad.GGDrawing.Description = drawing.Description{}
	}
	return ad
}

func (d *Drawing) UpdateFromApiDrawing(ad *api.Drawing) {
	d.ID = ad.ID
	d.Name = ad.Name
	d.Drawing = &ad.GGDrawing
	d.Area = d.Drawing.Area()
	d.Perimeter = d.Drawing.Perimeter()
	d.Height = d.Drawing.Height()
	d.Width = d.Drawing.Width()
}

func (u *User) ToApiUser(au *api.User) *api.User {
	if au == nil {
		au = &u.User
	} else {
		au.Login = u.Login
		au.Password = u.Password
		au.Permission = u.Permission
	}
	au.ID = u.ID
	return au
}

func (u *User) UpdateFromApiUser(au *api.User) {
	id := au.ID
	u.User = *au
	u.ID = id
}
