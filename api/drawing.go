package api

import "github.com/maxsid/goCeilings/drawing/raster"

type DrawingOpen struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Drawing struct {
	DrawingOpen
	raster.GGDrawing
}
