package storage

import (
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/drawing"
	"github.com/maxsid/goCeilings/drawing/raster"
	"gorm.io/gorm"
	"time"
)

type userModel struct {
	gorm.Model
	Login    string `gorm:"unique"`
	Password string
	Role     api.UserRole
}

type drawingModel struct {
	gorm.Model
	Name                           string
	Description                    *drawing.Description
	Area, Perimeter, Height, Width float64
	Drawing                        *raster.GGDrawing
}

type drawingPermissionModel struct {
	UserID                            uint          `gorm:"primaryKey;autoIncrement:false"`
	DrawingID                         uint          `gorm:"primaryKey;autoIncrement:false"`
	User                              *userModel    `gorm:"foreignKey:UserID"`
	Drawing                           *drawingModel `gorm:"foreignKey:DrawingID"`
	Get, Change, Delete, Share, Owner bool
	CreatedAt                         time.Time
	UpdatedAt                         time.Time
	DeletedAt                         gorm.DeletedAt `gorm:"index"`
}

func (udp *drawingPermissionModel) ToApi() *api.DrawingPermission {
	pol := &api.DrawingPermission{
		Get:    udp.Get,
		Change: udp.Change,
		Delete: udp.Delete,
		Share:  udp.Share,
		Owner:  udp.Owner,
	}
	if udp.User != nil {
		pol.User = &udp.User.ToApi().UserBasic
	}
	if udp.Drawing != nil {
		pol.Drawing = &udp.Drawing.ToApi().DrawingBasic
	}
	return pol
}

func (udp *drawingPermissionModel) IsFullFalse() bool {
	return !(udp.Get || udp.Change || udp.Delete || udp.Share || udp.Owner)
}

func (udp *drawingPermissionModel) UpdateFromApi(pol *api.DrawingPermission) {
	udp.UserID, udp.DrawingID = pol.User.ID, pol.Drawing.ID
	udp.Get, udp.Change, udp.Delete, udp.Share, udp.Owner = pol.Get, pol.Change, pol.Delete, pol.Share, pol.Owner
}

func (d *drawingModel) ToApi() *api.Drawing {
	ad := &api.Drawing{
		DrawingBasic: api.DrawingBasic{
			ID:   d.ID,
			Name: d.Name,
		},
		GGDrawing: *d.Drawing,
	}
	ad.GGDrawing.Description = d.Description
	if ad.GGDrawing.Description == nil {
		ad.GGDrawing.Description = drawing.NewDescription()
	}
	return ad
}

func (d *drawingModel) UpdateFromApi(ad *api.Drawing) {
	d.ID = ad.ID
	d.Name = ad.Name
	d.Drawing = &ad.GGDrawing
	d.Description = ad.Description
	d.Area = d.Drawing.Area()
	d.Perimeter = d.Drawing.Perimeter()
	d.Height = d.Drawing.Height()
	d.Width = d.Drawing.Width()
}

func (u *userModel) ToApi() *api.UserConfident {
	return &api.UserConfident{
		UserBasic: api.UserBasic{
			ID:    u.ID,
			Login: u.Login,
			Role:  u.Role,
		},
		Password: u.Password,
	}
}

func (u *userModel) UpdateFromApi(au *api.UserConfident) {
	u.Role = au.Role
	u.Login = au.Login
	u.Password = au.Password
	u.ID = au.ID
}
