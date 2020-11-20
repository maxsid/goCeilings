package figure

import (
	"math"
)

type Polygon struct {
	Points []*Point
}

func NewPolygon(points ...*Point) *Polygon {
	return &Polygon{Points: points}
}

func NewPolygonZero() *Polygon {
	return &Polygon{Points: []*Point{{0, 0}}}
}

func (pol *Polygon) Len() int {
	return len(pol.Points)
}

func (pol *Polygon) LastPoint() (*Point, error) {
	if pol.Len() == 0 {
		return nil, ErrNotEnoughPoints
	}
	return pol.Points[pol.Len()-1], nil
}

func (pol *Polygon) LastSide() (*Segment, error) {
	n := pol.Len()
	if n < 2 {
		return nil, ErrNotEnoughPoints
	}
	return &Segment{A: pol.Points[n-2], B: pol.Points[n-1]}, nil
}

func (pol *Polygon) DiagonalsLen() int {
	n := pol.Len()
	return n * (n - 3) / 2
}

func (pol *Polygon) Area() float64 {
	n := pol.Len()
	if n < 3 {
		return 0
	}
	var sum1, sum2 float64
	var p1, p2 *Point
	for i := 0; i < n; i++ {
		p1, p2 = pol.Points[i], pol.Points[(i+1)%n]
		sum1 += p1.X * p2.Y
		sum2 += p1.Y * p2.X
	}
	sum := sum1 - sum2
	return 0.5 * math.Abs(sum)
}

func (pol *Polygon) Perimeter() float64 {
	if pol.Len() < 2 {
		return 0
	}
	var sum float64
	for _, s := range pol.Sides() {
		sum += s.Distance()
	}
	return sum
}

func (pol *Polygon) Sides() []*Segment {
	lines := make([]*Segment, 0)
	n := pol.Len()
	if n <= 1 {
		return lines
	}
	for i := 0; i < n; i++ {
		lines = append(lines, &Segment{
			A: pol.Points[i],
			B: pol.Points[(i+1)%n],
		})
	}
	return lines
}

func (pol *Polygon) AddPoints(points ...*Point) {
	pol.Points = append(pol.Points, points...)
}

func (pol *Polygon) AddPoint(x, y float64) {
	pol.Points = append(pol.Points, &Point{X: x, Y: y})
}

func (pol *Polygon) AddPointByDirection(distance float64, direction float64) error {
	l, err := pol.LastPoint()
	if err != nil {
		return err
	}
	pol.Points = append(pol.Points, NewPointByDirection(l, distance, direction))
	return nil
}

func (pol *Polygon) AddPointByAngle(distance float64, angle float64) error {
	s, err := pol.LastSide()
	if err != nil {
		return err
	}
	pol.Points = append(pol.Points, NewPointByDirectionFromSegment(s, distance, angle))
	return nil
}

func (pol *Polygon) Width() float64 {
	if pol.Len() < 2 {
		return 0
	}
	minX, maxX := pol.Points[0].X, pol.Points[0].X
	for _, p := range pol.Points[1:] {
		if minX > p.X {
			minX = p.X
		}
		if maxX < p.X {
			maxX = p.X
		}
	}
	return math.Abs(maxX - minX)
}

func (pol *Polygon) Height() float64 {
	if pol.Len() < 2 {
		return 0
	}
	minY, maxY := pol.Points[0].Y, pol.Points[0].Y
	for _, p := range pol.Points[1:] {
		if minY > p.Y {
			minY = p.Y
		}
		if maxY < p.Y {
			maxY = p.Y
		}
	}
	return math.Abs(maxY - minY)
}

func (pol *Polygon) Rotation(o *Point, a float64) {
	for _, p := range pol.Points {
		if o == p {
			continue
		}
		x, y := p.X, p.Y
		p.X = x*math.Cos(a) - y*math.Sin(a)
		p.Y = x*math.Sin(a) + y*math.Cos(a)
	}
}
