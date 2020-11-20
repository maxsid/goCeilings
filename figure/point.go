package figure

import (
	"github.com/maxsid/goCeilings/value"
	"math"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func NewPointByDirection(cp *Point, distance, direction float64) *Point {
	return &Point{
		X: cp.X + distance*math.Cos(direction),
		Y: cp.Y + distance*math.Sin(direction),
	}
}

func NewPointByDirectionFromSegment(s *Segment, distance, direction float64) *Point {
	dir := PointDirection(s.B, s.A) + direction
	if d := int(dir / (math.Pi * 2)); d > 0 {
		dir -= (math.Pi * 2) * float64(d)
	}
	return NewPointByDirection(s.B, distance, dir)
}

func (p *Point) GetRoundedPoint(round int) *Point {
	return &Point{X: value.Round(p.X, round), Y: value.Round(p.Y, round)}
}
