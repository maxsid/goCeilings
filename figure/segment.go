package figure

import (
	"math"
)

type Segment struct {
	A, B *Point
}

func (l *Segment) Distance() float64 {
	ax, ay := l.A.X, l.A.Y
	bx, by := l.B.X, l.B.Y
	dist := math.Sqrt(math.Pow(ax-bx, 2) + math.Pow(ay-by, 2))
	return dist
}
