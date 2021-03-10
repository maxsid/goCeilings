package drawing

// Drawer is object which can draw an image and write it to []byte with Draw method.
// DrawingMIME returns MIME type of the image.
type Drawer interface {
	Draw(drawDescription bool) ([]byte, error)
	DrawingMIME() string
}

// DrawerGetter needs for returning Drawer with GetDrawer method.
type DrawerGetter interface {
	GetDrawer() Drawer
}
