package figure

import (
	"math"
)

func PointDirection(cp, p *Point) float64 {
	x, y := p.X-cp.X, p.Y-cp.Y
	r := math.Atan2(y, x)
	if r < 0 {
		return math.Pi*2 + r
	}
	return r
}
