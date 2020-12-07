package figure

import (
	"encoding/json"
	"fmt"
	"github.com/maxsid/goCeilings/value"
	"math"
)

// Point is a minimal element of any figure. Contains Coordinates on the drawing and Calculator, which calculates
// coordinates via CalculateCoordinates function and previous points of the figure.
type Point struct {
	X          float64                    `json:"x"`
	Y          float64                    `json:"y"`
	Calculator PointCoordinatesCalculator `json:"calculator,omitempty"`
}

// NewPoint creates new Point object only with Coordinates without Calculator.
func NewPoint(x, y float64) *Point {
	return &Point{
		X:          x,
		Y:          y,
		Calculator: nil,
	}
}

// NewCalculatedPoint creates new Point object with pcc Calculator, without calculating.
func NewCalculatedPoint(pcc PointCoordinatesCalculator) *Point {
	return &Point{
		Calculator: pcc,
	}
}

// CalculateCoordinates calculates coordinates for point, if Calculator is not nil.
func (p *Point) CalculateCoordinates(previousPoints ...*Point) error {
	if p.Calculator != nil {
		return p.Calculator.Calculate(p, previousPoints...)
	}
	return ErrPointDoesNotHaveCalculator
}

func (p *Point) RoundCoordinates(round int) {
	p.X, p.Y = value.Round(p.X, round), value.Round(p.Y, round)
}

func (p *Point) NewRoundedPoint(round int) *Point {
	np := &Point{p.X, p.Y, p.Calculator}
	np.RoundCoordinates(round)
	return np
}

type PointCoordinatesCalculator interface {
	Calculate(point *Point, previousPoints ...*Point) error
	ConvertToOne(measures *value.FigureMeasures)
	ConvertFromOne(measures *value.FigureMeasures)
	json.Unmarshaler
}

// DirectionCalculator is a calculator for calculating coordinates with distance and direction from previous point.
type DirectionCalculator struct {
	Direction float64 `json:"direction"`
	Distance  float64 `json:"distance"`
}

func (d *DirectionCalculator) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	_, ok1 := m["direction"]
	_, ok2 := m["distance"]
	if !ok1 || !ok2 {
		return fmt.Errorf("%w. Expected %T", ErrInvalidType, d)
	}
	d.Distance, d.Direction = m["distance"].(float64), m["direction"].(float64)
	return nil
}

func (d *DirectionCalculator) ConvertToOne(measures *value.FigureMeasures) {
	d.Distance = value.ConvertToOne(measures.Length, d.Distance)
	d.Direction = value.ConvertToOne(measures.Angle, d.Direction)
}

func (d *DirectionCalculator) ConvertFromOne(measures *value.FigureMeasures) {
	d.Distance = value.ConvertFromOne(measures.Length, d.Distance)
	d.Direction = value.ConvertFromOne(measures.Angle, d.Direction)
}

// Calculate calculates point coordinates by distance in metres and direction in radians.
// previousPoints parameter has to have at least one point.
func (d *DirectionCalculator) Calculate(point *Point, previousPoints ...*Point) error {
	if len(previousPoints) < 1 {
		return fmt.Errorf("%w for calculating direction (must have at least 1)", ErrNotEnoughPoints)
	}
	pp := previousPoints[len(previousPoints)-1]
	point.X, point.Y = pp.X+d.Distance*math.Cos(d.Direction), pp.Y+d.Distance*math.Sin(d.Direction)
	return nil
}

// AngleCalculator is a calculator for calculating coordinates with Distance and Angle from previous two points
// (segment), creating angle with specified Angle value.
type AngleCalculator struct {
	Angle    float64 `json:"angle"`
	Distance float64 `json:"distance"`
}

func (ac *AngleCalculator) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	_, ok1 := m["angle"]
	_, ok2 := m["distance"]
	if !ok1 || !ok2 {
		return fmt.Errorf("%w. Expected %T", ErrInvalidType, ac)
	}
	ac.Distance, ac.Angle = m["distance"].(float64), m["angle"].(float64)
	return nil
}

func (ac *AngleCalculator) ConvertToOne(measures *value.FigureMeasures) {
	ac.Distance = value.ConvertToOne(measures.Length, ac.Distance)
	ac.Angle = value.ConvertToOne(measures.Angle, ac.Angle)
}

func (ac *AngleCalculator) ConvertFromOne(measures *value.FigureMeasures) {
	ac.Distance = value.ConvertFromOne(measures.Length, ac.Distance)
	ac.Angle = value.ConvertFromOne(measures.Angle, ac.Angle)
}

func (ac AngleCalculator) Calculate(point *Point, previousPoints ...*Point) error {
	if len(previousPoints) < 2 {
		return fmt.Errorf("%w for calculating an angle (must have at least 2)", ErrNotEnoughPoints)
	}
	a, b := previousPoints[len(previousPoints)-2], previousPoints[len(previousPoints)-1]
	dir := pointDirection(b, a) + ac.Angle
	if d := int(dir / (math.Pi * 2)); d > 0 {
		dir -= (math.Pi * 2) * float64(d)
	}
	point.X, point.Y = b.X+ac.Distance*math.Cos(dir), b.Y+ac.Distance*math.Sin(dir)
	return nil
}

func (p *Point) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	_, ok1 := m["x"]
	_, ok2 := m["y"]
	if !ok1 || !ok2 {
		return fmt.Errorf("%w, got %v, must be Point", ErrInvalidType, m)
	}
	p.X, p.Y = m["x"].(float64), m["y"].(float64)
	if _, ok := m["calculator"]; !ok {
		return nil
	}
	calcBytes, err := json.Marshal(m["calculator"])
	if err != nil {
		return err
	}
	calculators := []PointCoordinatesCalculator{
		&DirectionCalculator{},
		&AngleCalculator{},
	}
	for _, calc := range calculators {
		if err := json.Unmarshal(calcBytes, calc); err == nil {
			p.Calculator = calc
			return nil
		}
	}
	return fmt.Errorf("%w of calculator: got %v", ErrInvalidType, m["calculator"])
}
