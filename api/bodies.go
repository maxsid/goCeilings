package api

import (
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
	Point    PointCalculating          `json:"point"`
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
