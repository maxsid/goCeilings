package figure

import (
	"errors"
	"fmt"
)

var (
	ErrTooMuchPoints   = errors.New("too much points")
	ErrNotEnoughPoints = errors.New("not enough points")

	ErrInvalidType        = errors.New("invalid type")
	ErrVariableIsNotSlice = fmt.Errorf("%w: variable is not a slice", ErrInvalidType)

	ErrPointDoesNotHaveCalculator = errors.New("point does not have a calculator")
)
