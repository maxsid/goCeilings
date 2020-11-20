package api

import (
	"github.com/maxsid/goCeilings/drawing/raster"
	"github.com/maxsid/goCeilings/figure"
	"github.com/maxsid/goCeilings/value"
)

// Request and responses JSON bodies
type ListStatData struct {
	Amount    uint `json:"amount"`
	Page      uint `json:"page"`
	PageLimit uint `json:"page_limit"`
	Pages     uint `json:"pages"`
}

type UsersListResponseData struct {
	Users []*UserOpen `json:"users"`
	ListStatData
}

type DrawingsListResponseData struct {
	Drawings []*DrawingOpen `json:"drawings"`
	ListStatData
}

type DrawingCalculatingData struct {
	Area        float64 `json:"area"`
	Perimeter   float64 `json:"perimeter"`
	PointsCount int     `json:"points_count"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
}

type DrawingGetResponseData struct {
	DrawingOpen
	DrawingCalculatingData
	Points   []*figure.Point            `json:"points"`
	Measures *value.FigureMeasuresNames `json:"measures"`
}

type DrawingPostPutRequestData struct {
	DrawingOpen
	Points   []*PointCalculating       `json:"points"`
	Measures value.FigureMeasuresNames `json:"measures"`
}

type PointCalculating struct {
	X         float64  `json:"x"`
	Y         float64  `json:"y"`
	Distance  float64  `json:"distance"`
	Direction *float64 `json:"direction"`
	Angle     *float64 `json:"angle"`
}

type PointCalculatingWithMeasures struct {
	PointCalculating
	Measures value.FigureMeasuresNames `json:"measures"`
}

type PointWithMeasure struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Measure string  `json:"measure"`
}

type DrawingPointsGettingResponseData struct {
	DrawingOpen
	Points  []*figure.Point `json:"points"`
	Measure string          `json:"measure"`
}

type PointsCalculatingWithMeasures struct {
	Points   []*PointCalculating       `json:"points"`
	Measures value.FigureMeasuresNames `json:"measures"`
}

// Storage models
type DrawingOpen struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Drawing struct {
	DrawingOpen
	raster.GGDrawing
}

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

type DataKeeper interface {
	UserManager
	DrawingManager
}
