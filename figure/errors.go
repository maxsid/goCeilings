package figure

import "errors"

var (
	ErrTooMuchPoints   = errors.New("too much points for this shape")
	ErrNotEnoughPoints = errors.New("figure hasn't enough number of points for this operation")
)
