package figure

import (
	"errors"
	"math"
)

type Polygon struct {
	Points []*Point
}

// NewPolygon creates new polygon with points, then calculates coordinates.
// Function can execute panic.
func NewPolygon(points ...*Point) *Polygon {
	pol := &Polygon{Points: points}
	if err := pol.CalculatePoints(); err != nil {
		panic(err)
	}
	return pol
}

// NewPolygonZero returns polygon with one point, which has coordinates {x: 0, y: 0}.
func NewPolygonZero() *Polygon {
	return &Polygon{Points: []*Point{{X: 0, Y: 0}}}
}

// Len returns a number of the points.
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
	n := float64(pol.Len())
	d := (math.Pow(n, 2) - 3*n) / 2
	if d < 0 {
		return 0
	}
	return int(d)
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

// AddPoints adds points and recalculates all points coordinates.
func (pol *Polygon) AddPoints(points ...*Point) error {
	i := pol.Len()
	pol.Points = append(pol.Points, points...)
	for ; i < pol.Len(); i++ {
		if err := pol.calculatePoint(i); err != nil && !errors.Is(err, ErrPointDoesNotHaveCalculator) {
			return err
		}
	}
	return nil
}

// AddPoint adds points with coordinates, without calculating.
func (pol *Polygon) AddPoint(x, y float64) {
	pol.Points = append(pol.Points, &Point{X: x, Y: y})
}

// SetPoint changes point by index and calculates new coordinates of points.
func (pol *Polygon) SetPoint(i int, point *Point) error {
	endOfPoints := pol.Points[i:]
	pol.Points = pol.Points[:i]
	endOfPoints[0] = point
	return pol.AddPoints(endOfPoints...)
}

// AddPointByDirection adds new point with DirectionCalculator and calculates new coordinates.
func (pol *Polygon) AddPointByDirection(distance float64, direction float64) error {
	p := NewCalculatedPoint(&DirectionCalculator{Direction: direction, Distance: distance})
	return pol.AddPoints(p)
}

// AddPointByAngle adds new point with AngleCalculator and calculates new coordinates.
func (pol *Polygon) AddPointByAngle(distance float64, angle float64) error {
	p := NewCalculatedPoint(&AngleCalculator{
		Angle:    angle,
		Distance: distance,
	})
	return pol.AddPoints(p)
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

// Rotation rotates the polygon where o Points is the central point and a is an angle of rotation in radians.
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

// calculatePoint calculates coordinates of the point by index in Polygon.Points.
func (pol *Polygon) calculatePoint(index int) error {
	previous, err := sliceOfForwardElements(index, pol.Points)
	if err != nil {
		return err
	}
	if err = pol.Points[index].CalculateCoordinates(previous.([]*Point)...); err != nil {
		return err
	}
	return nil
}

// CalculatePoints calculates all points coordinates in the polygon.
func (pol *Polygon) CalculatePoints() error {
	for i := range pol.Points {
		if err := pol.calculatePoint(i); err != nil && !errors.Is(err, ErrPointDoesNotHaveCalculator) {
			return err
		}
	}
	return nil
}

// RoundAllPoints rounds coordinates of all points.
func (pol *Polygon) RoundAllPoints(round int) {
	for _, p := range pol.Points {
		p.RoundCoordinates(round)
	}
}

// RoundCalculatedPoints rounds coordinates only points with calculator.
func (pol *Polygon) RoundCalculatedPoints(round int) {
	for _, p := range pol.Points {
		if p.Calculator == nil {
			continue
		}
		p.RoundCoordinates(round)
	}
}
