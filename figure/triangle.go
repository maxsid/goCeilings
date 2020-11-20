package figure

type Triangle struct {
	*Polygon
}

func NewTriangle(points ...*Point) (*Triangle, error) {
	t := Triangle{&Polygon{[]*Point{}}}
	if err := t.AddPoints(points...); err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *Triangle) AddPoints(points ...*Point) error {
	if t.Len()+len(points) > 3 {
		return ErrTooMuchPoints
	}
	t.Polygon.AddPoints(points...)
	return nil
}

func (t *Triangle) AddPointByDirection(d float64, a float64) error {
	if t.Len() >= 3 {
		return ErrTooMuchPoints
	}
	return t.Polygon.AddPointByDirection(d, a)
}
