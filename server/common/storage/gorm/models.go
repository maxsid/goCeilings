package gorm

import (
	"time"

	"github.com/maxsid/goCeilings/drawing"
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/server/common"
	"gorm.io/gorm"
)

type userModel struct {
	gorm.Model
	Login    string `gorm:"unique"`
	Password string
	Role     common.UserRole
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

func (udp *drawingPermissionModel) ToAPI() *common.DrawingPermission {
	pol := &common.DrawingPermission{
		Get:    udp.Get,
		Change: udp.Change,
		Delete: udp.Delete,
		Share:  udp.Share,
		Owner:  udp.Owner,
	}
	if udp.User != nil {
		pol.User = &udp.User.ToAPI().UserBasic
	}
	if udp.Drawing != nil {
		pol.Drawing = &udp.Drawing.ToAPI().DrawingBasic
	}
	return pol
}

func (udp *drawingPermissionModel) IsFullFalse() bool {
	return !(udp.Get || udp.Change || udp.Delete || udp.Share || udp.Owner)
}

func (udp *drawingPermissionModel) FromAPI(pol *common.DrawingPermission) {
	udp.UserID, udp.DrawingID = pol.User.ID, pol.Drawing.ID
	udp.Get, udp.Change, udp.Delete, udp.Share, udp.Owner = pol.Get, pol.Change, pol.Delete, pol.Share, pol.Owner
}

func (d *drawingModel) ToAPI() *common.Drawing {
	ad := &common.Drawing{
		DrawingBasic: common.DrawingBasic{
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

func (d *drawingModel) FromAPI(ad *common.Drawing) {
	d.ID = ad.ID
	d.Name = ad.Name
	d.Drawing = &ad.GGDrawing
	d.Description = ad.Description
	d.Area = d.Drawing.Area()
	d.Perimeter = d.Drawing.Perimeter()
	d.Height = d.Drawing.Height()
	d.Width = d.Drawing.Width()
}

func (u *userModel) ToAPI() *common.UserConfident {
	return &common.UserConfident{
		UserBasic: common.UserBasic{
			ID:    u.ID,
			Login: u.Login,
			Role:  u.Role,
		},
		Password: u.Password,
	}
}

func (u *userModel) FromAPI(au *common.UserConfident) {
	u.Role = au.Role
	u.Login = au.Login
	u.Password = au.Password
	u.ID = au.ID
}
