package raster

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/maxsid/goCeilings/drawing"
	"github.com/maxsid/goCeilings/drawing/naming"
	"github.com/maxsid/goCeilings/figure"
	"github.com/maxsid/goCeilings/value"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	numbersPrecision                                             = 2
	drawingWidth, drawingHeight                                  = 1600, 1600
	descriptionWidth                                             = 320
	fontSizeSideTitle, fontSizePointTitle, fontSizeNotes         = 20, 20, 20
	lineWidth                                            float64 = 3
	marginLeft, marginRight, marginTop, marginDown               = 35, 35, 35, 35
	marginHorizontal, marginVertical                             = marginLeft + marginRight, marginTop + marginDown
	pointSize                                            float64 = 3
	marginLetterX, marginLetterY                         float64 = 4, 20
)

type GGDrawing struct {
	figure.Polygon
	Description      *drawing.Description  `json:"description"`
	Measures         *value.FigureMeasures `json:"measures"`
	offsetX, offsetY float64
}

func NewEmptyGGDrawing() *GGDrawing {
	d := &GGDrawing{
		Polygon:     *figure.NewPolygon(),
		Description: drawing.NewDescription(),
		Measures:    value.NewFigureMeasures(),
	}
	return d
}

func NewGGDrawing() *GGDrawing {
	d := &GGDrawing{
		Polygon:     *figure.NewPolygonZero(),
		Description: drawing.NewDescription(),
		Measures:    value.NewFigureMeasures(),
	}
	return d
}

func (d *GGDrawing) Draw(drawDesc bool) ([]byte, error) {
	if d.Len() < 3 {
		return nil, fmt.Errorf("%w for drawing (%d), have to be at least 3", ErrTooFewPoints, d.Len())
	}
	scale := calcDrawScale(d.Polygon.Width(), d.Polygon.Height())
	d.updateOffset(scale)
	imageWidth, imageHeight := calcImageSize(scale, d.Polygon.Width(), d.Polygon.Height(), drawDesc)
	ggCtx := gg.NewContext(imageWidth, imageHeight)
	setBackground(ggCtx)
	d.drawLines(ggCtx, scale)
	d.drawPoints(ggCtx, scale)
	if err := setFontSize(ggCtx, fontSizeSideTitle); err != nil {
		return nil, err
	}
	d.drawPointsTitles(ggCtx, imageHeight, scale)
	if err := setFontSize(ggCtx, fontSizePointTitle); err != nil {
		return nil, err
	}
	d.drawLinesTitles(ggCtx, imageHeight, scale)
	if drawDesc {
		desc := drawing.NewDescription()
		d.addPolygonInfoToDescription(desc)
		d.addSidesToDescription(desc)
		d.addPointsToDescription(desc)
		if err := setFontSize(ggCtx, fontSizeNotes); err != nil {
			return nil, err
		}
		d.drawDescription(ggCtx, scale, drawing.NewUnionDescription(d.Description, desc))
	}
	return contextToPNGBytes(ggCtx)
}

func (d *GGDrawing) DrawingMIME() string {
	return "image/png"
}

func (d *GGDrawing) GetDrawer() drawing.Drawer {
	return d
}

func (d *GGDrawing) updateOffset(scale float64) {
	left, _ := d.Polygon.LeftPoint()
	low, _ := d.Polygon.LowPoint()
	d.offsetX, d.offsetY = -left.X*scale, -low.Y*scale
}

func (d *GGDrawing) drawPoints(ggCtx *gg.Context, scale float64) {
	ggCtx.InvertY()
	defer ggCtx.InvertY()
	ggCtx.SetColor(colornames.Black)
	for _, p := range d.Points {
		x, y := getXYOnDrawing(p, d.offsetX, d.offsetY, scale)
		ggCtx.DrawPoint(x, y, pointSize)
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawLines(ggCtx *gg.Context, scale float64) {
	ggCtx.InvertY()
	defer ggCtx.InvertY()
	pol := figure.Polygon{Points: d.Points}
	ggCtx.SetColor(colornames.Red)
	ggCtx.SetLineWidth(lineWidth)
	for _, s := range pol.Sides() {
		x1, y1 := getXYOnDrawing(s.A, d.offsetX, d.offsetY, scale)
		x2, y2 := getXYOnDrawing(s.B, d.offsetX, d.offsetY, scale)
		ggCtx.DrawLine(x1, y1, x2, y2)
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawLinesTitles(ggCtx *gg.Context, imageHeight int, scale float64) {
	pol := d.Polygon
	for _, l := range pol.Sides() {
		dist := fmt.Sprint(value.ConvertFromOneRound(d.Measures.Length, l.Distance(), 2))
		w, h := ggCtx.MeasureString(dist)
		x1, y1 := getXYOnDrawing(l.A, d.offsetX, d.offsetY, scale)
		x2, y2 := getXYOnDrawing(l.B, d.offsetX, d.offsetY, scale)
		x, y := (x1+x2)/2-(w/2), float64(imageHeight)-((y1+y2)/2-(h/2))
		ggCtx.SetColor(colornames.White)
		ggCtx.DrawRectangle(x+2, y-h, w-2, h)
		ggCtx.Fill()
		ggCtx.SetColor(colornames.Black)
		ggCtx.DrawString(dist, x, y)
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawPointsTitles(ggCtx *gg.Context, imageHeight int, scale float64) {
	ni := naming.NewNameIterator('A', 'Z')
	for _, p := range d.Points {
		x, y := getXYOnDrawing(p, d.offsetX, d.offsetY, scale)
		ggCtx.DrawString(ni.Next(), marginLetterX+x, float64(imageHeight)-(y-marginLetterY))
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawDescription(ggCtx *gg.Context, drawingScale float64, desc *drawing.Description) {
	s := strings.Join(desc.ToStringSlice(), "\n")
	sx := value.Round(d.Polygon.Width()*drawingScale, 2)
	sx += marginHorizontal + marginLeft
	ggCtx.DrawStringWrapped(s, sx, marginTop, 0, 0, descriptionWidth, 1.5, gg.AlignLeft)
	ggCtx.Stroke()
}

func (d *GGDrawing) addPointsToDescription(desc *drawing.Description) {
	ni := naming.NewNameIterator('A', 'Z')
	ps := make([]string, len(d.Points))
	for i, p := range d.Points {
		x := value.ConvertFromOneRound(d.Measures.Length, p.X, 2)
		y := value.ConvertFromOneRound(d.Measures.Length, p.Y, 2)
		ps[i] = fmt.Sprintf("%s=(%v;%v)", ni.Next(), x, y)
	}
	desc.PushBack("Points", strings.Join(ps, ", "))
}

func (d *GGDrawing) addSidesToDescription(desc *drawing.Description) {
	pol := d.Polygon
	ni := naming.NewNameIterator('A', 'Z')
	sides, l1, l2 := pol.Sides(), ni.Next(), ni.Next()
	ss := make([]string, len(sides))
	for i, s := range sides {
		dist := value.ConvertFromOneRound(d.Measures.Length, s.Distance(), 2)
		ss[i] = fmt.Sprintf("%s%s=%v", l1, l2, dist)
		l1, l2 = l2, ni.Next()
	}
	desc.PushBack("Sides", strings.Join(ss, ", "))
}

func (d *GGDrawing) addPolygonInfoToDescription(desc *drawing.Description) {
	desc.PushBack("Area", fmt.Sprintf("%.2f", d.Area()))
	desc.PushBack("Perimeter", fmt.Sprintf("%.2f", d.Perimeter()))
	desc.PushBack("Width", fmt.Sprintf("%.2f", d.Width()))
	desc.PushBack("Height", fmt.Sprintf("%.2f", d.Height()))
	desc.PushBack("Points", fmt.Sprintf("%d", d.Len()))
}

func (d *GGDrawing) AddPoints(points ...*figure.Point) error {
	for _, p := range points {
		if p.Calculator == nil {
			p.X, p.Y = value.ConvertToOne(d.Measures.Length, p.X), value.ConvertToOne(d.Measures.Length, p.Y)
		} else {
			p.Calculator.ConvertToOne(d.Measures)
		}
	}
	return d.Polygon.AddPoints(points...)
}

func (d *GGDrawing) AddPoint(x, y float64) {
	x, y = value.ConvertToOne(d.Measures.Length, x), value.ConvertToOne(d.Measures.Length, y)
	d.Polygon.AddPoint(x, y)
}

func (d *GGDrawing) AddPointByDirection(distance float64, direction float64) error {
	distance = value.ConvertToOne(d.Measures.Length, distance)
	direction = value.ConvertToOne(d.Measures.Angle, direction)
	return d.Polygon.AddPointByDirection(distance, direction)
}

func (d *GGDrawing) AddPointByAngle(distance float64, angle float64) error {
	distance = value.ConvertToOne(d.Measures.Length, distance)
	angle = value.ConvertToOne(d.Measures.Angle, angle)
	return d.Polygon.AddPointByAngle(distance, angle)
}

func (d *GGDrawing) SetPoint(i int, point *figure.Point) error {
	if point.Calculator == nil {
		point.X, point.Y = value.ConvertToOne(d.Measures.Length, point.X), value.ConvertToOne(d.Measures.Length, point.Y)
	} else {
		point.Calculator.ConvertToOne(d.Measures)
	}
	return d.Polygon.SetPoint(i, point)
}

func (d *GGDrawing) Area() float64 {
	return value.ConvertFromOneRound(d.Measures.Area, d.Polygon.Area(), numbersPrecision)
}

func (d *GGDrawing) Perimeter() float64 {
	return value.ConvertFromOneRound(d.Measures.Perimeter, d.Polygon.Perimeter(), numbersPrecision)
}

func (d *GGDrawing) Width() float64 {
	return value.ConvertFromOneRound(d.Measures.Length, d.Polygon.Width(), numbersPrecision)
}

func (d *GGDrawing) Height() float64 {
	return value.ConvertFromOneRound(d.Measures.Length, d.Polygon.Height(), numbersPrecision)
}

func (d *GGDrawing) GetPoints() []*figure.Point {
	return d.GetPointsWithParams(d.Measures.Length, numbersPrecision)
}

func (d *GGDrawing) GetPointsWithParams(m value.Measure, precision int) []*figure.Point {
	points := make([]*figure.Point, len(d.Points))
	for i, p := range d.Points {
		points[i] = &figure.Point{
			X: value.ConvertFromOneRound(m, p.X, precision),
			Y: value.ConvertFromOneRound(m, p.Y, precision),
		}
	}
	return points
}

func (d *GGDrawing) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		sData, ok := value.(string)
		if !ok {
			return ErrWrongValueScanType
		}
		data = []byte(sData)
	}
	return json.Unmarshal(data, d)
}

func (d *GGDrawing) Value() (driver.Value, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func calcDrawScale(polW, polH float64) float64 {
	scale := (drawingWidth - marginHorizontal) / polW
	hScale := (drawingHeight - marginVertical) / polH
	if scale > hScale {
		scale = hScale
	}
	return scale
}

func calcImageSize(drawScale, polW, polH float64, drawDesc bool) (w, h int) {
	w, h = int(value.Round(polW*drawScale, 0)), int(value.Round(polH*drawScale, 0))
	w, h = w+marginHorizontal, h+marginVertical
	if drawDesc {
		w += marginHorizontal + descriptionWidth
	}
	return
}

func contextToPNGBytes(ctx *gg.Context) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := ctx.EncodePNG(buf); err != nil {
		return nil, err
	}
	out, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func getXYOnDrawing(p *figure.Point, offsetX, offsetY, scale float64) (x, y float64) {
	x = offsetX + marginLeft + p.X*scale
	y = offsetY + marginDown + p.Y*scale
	return
}

func setBackground(ggCtx *gg.Context) {
	ggCtx.SetColor(colornames.White)
	ggCtx.Clear()
}

func setFontSize(ggCtx *gg.Context, size float64) error {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: size,
	})
	ggCtx.SetFontFace(face)
	return nil
}
