package raster

import "errors"

var (
	ErrWrongValueScanType = errors.New("wrong value scan type")
	ErrTooFewPoints       = errors.New("too few points")
)
