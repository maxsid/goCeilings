package figure

import (
	"math"
	"reflect"
)

// pointDirection returns direction in radians from cp to p.
func pointDirection(cp, p *Point) float64 {
	x, y := p.X-cp.X, p.Y-cp.Y
	r := math.Atan2(y, x)
	if r < 0 {
		return math.Pi*2 + r
	}
	return r
}

// sliceOfForwardElements returns slice of all forward elements of i.
// The first element of the result slice is i+1 and the last is i-1 (i+1, i+2, ..., i-2, i-1).
// For example, sliceOfForwardElements(3, []int{0,1,2,3,4,5,6}) returns []int{4,5,6,0,1,2}
func sliceOfForwardElements(i int, slice interface{}) (interface{}, error) {
	sliceOf := reflect.ValueOf(slice)
	if sliceOf.Kind() != reflect.Slice {
		return nil, ErrVariableIsNotSlice
	}
	endSlice := sliceOf.Slice(0, i)
	startSlice := sliceOf.Slice(i+1, sliceOf.Len())
	return reflect.AppendSlice(startSlice, endSlice).Interface(), nil
}
