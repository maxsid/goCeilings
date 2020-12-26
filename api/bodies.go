package api

import (
	"github.com/maxsid/goCeilings/figure"
	"github.com/maxsid/goCeilings/value"
)

// Request and responses JSON bodies
type listStatData struct {
	Amount    uint `json:"amount"`
	Page      uint `json:"page"`
	PageLimit uint `json:"page_limit"`
	Pages     uint `json:"pages"`
}

type usersListResponseData struct {
	Users []*UserBasic `json:"users"`
	listStatData
}

type drawingsListResponseData struct {
	Drawings []*DrawingBasic `json:"drawings"`
	listStatData
}

type drawingCalculatedData struct {
	Area        float64 `json:"area"`
	Perimeter   float64 `json:"perimeter"`
	PointsCount int     `json:"points_count"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
}

type drawingGetResponseData struct {
	DrawingBasic
	drawingCalculatedData
	Points   []*figure.Point            `json:"points"`
	Measures *value.FigureMeasuresNames `json:"measures"`
}

type drawingPostPutRequestData struct {
	DrawingBasic
	Points   []*pointCalculating       `json:"points"`
	Measures value.FigureMeasuresNames `json:"measures"`
}

type pointCalculating struct {
	X         float64  `json:"x"`
	Y         float64  `json:"y"`
	Distance  float64  `json:"distance"`
	Direction *float64 `json:"direction"`
	Angle     *float64 `json:"angle"`
}

type pointCalculatingWithMeasures struct {
	Point    pointCalculating          `json:"point"`
	Measures value.FigureMeasuresNames `json:"measures"`
}

type pointWithMeasure struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Measure string  `json:"measure"`
}

type drawingPointsGettingResponseData struct {
	DrawingBasic
	Points  []*figure.Point `json:"points"`
	Measure string          `json:"measure"`
}

type pointsCalculatingWithMeasures struct {
	Points   []*pointCalculating       `json:"points"`
	Measures value.FigureMeasuresNames `json:"measures"`
}

type drawingPermissionCreating struct {
	UserID    uint `json:"user_id"`
	DrawingID uint `json:"drawing_id"`
	Get       bool `json:"get"`
	Change    bool `json:"change"`
	Delete    bool `json:"delete"`
	Share     bool `json:"share"`
}
