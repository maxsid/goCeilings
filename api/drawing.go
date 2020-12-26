package api

import "github.com/maxsid/goCeilings/drawing/raster"

// DrawingBasic contains basic drawing information.
type DrawingBasic struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// Drawing contains full drawing information, including DrawingBasic.
type Drawing struct {
	DrawingBasic
	raster.GGDrawing
}

// DrawingPermission contains possible user operations.
type DrawingPermission struct {
	User    *UserBasic    `json:"user"`
	Drawing *DrawingBasic `json:"drawing"`
	Get     bool          `json:"get,omitempty"`
	Change  bool          `json:"change,omitempty"`
	Delete  bool          `json:"delete,omitempty"`
	Share   bool          `json:"share,omitempty"`
	Owner   bool          `json:"owner,omitempty"`
}
