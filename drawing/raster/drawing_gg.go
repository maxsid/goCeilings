package raster

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/maxsid/goCeilings/drawing"
	"github.com/maxsid/goCeilings/drawing/naming"
	. "github.com/maxsid/goCeilings/figure"
	. "github.com/maxsid/goCeilings/value"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"strings"
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
	Polygon
	Description *drawing.Description `json:"description"`
	imageWidth  float64
	imageHeight float64
	Measures    *FigureMeasures `json:"measures"`
}

func NewEmptyGGDrawing() *GGDrawing {
	d := &GGDrawing{
		Polygon:     *NewPolygon(),
		Description: drawing.NewDescription(),
		Measures:    NewFigureMeasures(),
	}
	return d
}

func NewGGDrawingWithPoints(points ...*Point) (*GGDrawing, error) {
	d := NewEmptyGGDrawing()
	if err := d.AddPoints(points...); err != nil {
		return nil, err
	}
	return d, nil
}

func NewGGDrawing() *GGDrawing {
	d := &GGDrawing{
		Polygon:     *NewPolygonZero(),
		Description: drawing.NewDescription(),
		Measures:    NewFigureMeasures(),
		imageWidth:  drawingWidth + marginHorizontal,
		imageHeight: drawingHeight + marginVertical,
	}
	return d
}

func (d *GGDrawing) Draw(drawDesc bool) (image.Image, error) {
	scale := d.calcDrawScale()
	d.calcImageSize(scale, drawDesc)
	ggCtx := gg.NewContext(int(d.imageWidth), int(d.imageHeight))
	d.setBackground(ggCtx)
	d.drawLines(ggCtx, scale)
	d.drawPoints(ggCtx, scale)
	if err := d.setFontSize(ggCtx, fontSizeSideTitle); err != nil {
		return nil, err
	}
	d.drawPointsTitles(ggCtx, scale)
	if err := d.setFontSize(ggCtx, fontSizePointTitle); err != nil {
		return nil, err
	}
	d.drawLinesTitles(ggCtx, scale)
	if drawDesc {
		desc := drawing.NewDescription()
		d.addPolygonInfoNotes(desc)
		d.addSidesNote(desc)
		d.addPointsNote(desc)
		if err := d.setFontSize(ggCtx, fontSizeNotes); err != nil {
			return nil, err
		}
		d.drawDescription(ggCtx, scale, drawing.NewUnionDescription(d.Description, desc))
	}
	return ggCtx.Image(), nil
}

func (d *GGDrawing) setBackground(ggCtx *gg.Context) {
	ggCtx.SetColor(colornames.White)
	ggCtx.Clear()
}

func (d *GGDrawing) drawPoints(ggCtx *gg.Context, scale float64) {
	ggCtx.InvertY()
	defer ggCtx.InvertY()
	ggCtx.SetColor(colornames.Black)
	for _, p := range d.Points {
		x, y := d.getXY(p, scale)
		ggCtx.DrawPoint(x, y, pointSize)
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawLines(ggCtx *gg.Context, scale float64) {
	ggCtx.InvertY()
	defer ggCtx.InvertY()
	pol := Polygon{Points: d.Points}
	ggCtx.SetColor(colornames.Red)
	ggCtx.SetLineWidth(lineWidth)
	for _, s := range pol.Sides() {
		x1, y1 := d.getXY(s.A, scale)
		x2, y2 := d.getXY(s.B, scale)
		ggCtx.DrawLine(x1, y1, x2, y2)
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawLinesTitles(ggCtx *gg.Context, scale float64) {
	pol := d.Polygon
	for _, l := range pol.Sides() {
		dist := fmt.Sprint(ConvertFromOneRound(d.Measures.Length, l.Distance(), 2))
		w, h := ggCtx.MeasureString(dist)
		x1, y1 := d.getXY(l.A, scale)
		x2, y2 := d.getXY(l.B, scale)
		x, y := (x1+x2)/2-(w/2), d.imageHeight-((y1+y2)/2-(h/2))
		ggCtx.SetColor(colornames.White)
		ggCtx.DrawRectangle(x+2, y-h, w-2, h)
		ggCtx.Fill()
		ggCtx.SetColor(colornames.Black)
		ggCtx.DrawString(dist, x, y)
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawPointsTitles(ggCtx *gg.Context, scale float64) {
	ni := naming.NewNameIterator('A', 'Z')
	for _, p := range d.Points {
		x, y := d.getXY(p, scale)
		ggCtx.DrawString(ni.Next(), marginLetterX+x, d.imageHeight-(y-marginLetterY))
		ggCtx.Stroke()
	}
}

func (d *GGDrawing) drawDescription(ggCtx *gg.Context, drawingScale float64, desc *drawing.Description) {
	s := strings.Join(desc.ToStringSlice(), "\n")
	sx, _ := d.calcDrawingSize(drawingScale)
	sx += marginHorizontal + marginLeft
	ggCtx.DrawStringWrapped(s, sx, marginTop, 0, 0, descriptionWidth, 1.5, gg.AlignLeft)
	ggCtx.Stroke()
}

func (d *GGDrawing) addPointsNote(desc *drawing.Description) {
	ni := naming.NewNameIterator('A', 'Z')
	ps := make([]string, len(d.Points))
	for i, p := range d.Points {
		x := ConvertFromOneRound(d.Measures.Length, p.X, 2)
		y := ConvertFromOneRound(d.Measures.Length, p.Y, 2)
		ps[i] = fmt.Sprintf("%s=(%v;%v)", ni.Next(), x, y)
	}
	desc.PushBack("Points", strings.Join(ps, ", "))
}

func (d *GGDrawing) addSidesNote(desc *drawing.Description) {
	pol := d.Polygon
	ni := naming.NewNameIterator('A', 'Z')
	sides, l1, l2 := pol.Sides(), ni.Next(), ni.Next()
	ss := make([]string, len(sides))
	for i, s := range sides {
		dist := ConvertFromOneRound(d.Measures.Length, s.Distance(), 2)
		ss[i] = fmt.Sprintf("%s%s=%v", l1, l2, dist)
		l1, l2 = l2, ni.Next()
	}
	desc.PushBack("Sides", strings.Join(ss, ", "))
}

func (d *GGDrawing) addPolygonInfoNotes(desc *drawing.Description) {
	desc.PushBack("Area", fmt.Sprintf("%.2f", d.Area()))
	desc.PushBack("Perimeter", fmt.Sprintf("%.2f", d.Perimeter()))
	desc.PushBack("Width", fmt.Sprintf("%.2f", d.Width()))
	desc.PushBack("Height", fmt.Sprintf("%.2f", d.Height()))
	desc.PushBack("Points", fmt.Sprintf("%d", d.Len()))
}

func (d *GGDrawing) getXY(p *Point, scale float64) (x, y float64) {
	x = marginLeft + (ConvertFromOneRound(d.Measures.Length, p.X, 2) * scale)
	y = marginDown + (ConvertFromOneRound(d.Measures.Length, p.Y, 2) * scale)
	return
}

func (d *GGDrawing) getContextHeight(scale float64) int {
	_, hf := d.calcDrawingSize(scale)
	h := int(hf)
	if hf-float64(h) != 0 {
		h++
	}
	return h + marginVertical
}

func (d *GGDrawing) calcDrawScale() float64 {
	pol := d.Polygon
	wScale := (drawingWidth - marginHorizontal) / ConvertFromOne(d.Measures.Length, pol.Width())
	hScale := (drawingHeight - marginVertical) / ConvertFromOne(d.Measures.Length, pol.Height())
	scale := wScale
	if wScale > hScale {
		scale = hScale
	}
	return scale
}

func (d *GGDrawing) calcImageSize(drawScale float64, drawDesc bool) {
	w, h := d.calcDrawingSize(drawScale)
	d.imageWidth, d.imageHeight = w+marginHorizontal, h+marginVertical
	if drawDesc {
		d.imageWidth += marginHorizontal + descriptionWidth
	}
}

func (d *GGDrawing) calcDrawingSize(drawScale float64) (w, h float64) {
	pol := d.Polygon
	w = Round(ConvertFromOne(d.Measures.Length, pol.Width())*drawScale, 2)
	h = Round(ConvertFromOne(d.Measures.Length, pol.Height())*drawScale, 2)
	return
}

func (d *GGDrawing) setFontSize(ggCtx *gg.Context, size float64) error {
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

func (d *GGDrawing) AddPoints(points ...*Point) error {
	for _, p := range points {
		if p.Calculator == nil {
			p.X, p.Y = ConvertToOne(d.Measures.Length, p.X), ConvertToOne(d.Measures.Length, p.Y)
		} else {
			p.Calculator.ConvertToOne(d.Measures)
		}
	}
	return d.Polygon.AddPoints(points...)
}

func (d *GGDrawing) AddPoint(x, y float64) {
	x, y = ConvertToOne(d.Measures.Length, x), ConvertToOne(d.Measures.Length, y)
	d.Polygon.AddPoint(x, y)
}

func (d *GGDrawing) AddPointByDirection(distance float64, direction float64) error {
	distance = ConvertToOne(d.Measures.Length, distance)
	direction = ConvertToOne(d.Measures.Angle, direction)
	return d.Polygon.AddPointByDirection(distance, direction)
}

func (d *GGDrawing) AddPointByAngle(distance float64, angle float64) error {
	distance = ConvertToOne(d.Measures.Length, distance)
	angle = ConvertToOne(d.Measures.Angle, angle)
	return d.Polygon.AddPointByAngle(distance, angle)
}

func (d *GGDrawing) SetPoint(i int, point *Point) error {
	if point.Calculator == nil {
		point.X, point.Y = ConvertToOne(d.Measures.Length, point.X), ConvertToOne(d.Measures.Length, point.Y)
	} else {
		point.Calculator.ConvertToOne(d.Measures)
	}
	return d.Polygon.SetPoint(i, point)
}

func (d *GGDrawing) Area() float64 {
	return ConvertFromOneRound(d.Measures.Area, d.Polygon.Area(), numbersPrecision)
}

func (d *GGDrawing) Perimeter() float64 {
	return ConvertFromOneRound(d.Measures.Perimeter, d.Polygon.Perimeter(), numbersPrecision)
}

func (d *GGDrawing) Width() float64 {
	return ConvertFromOneRound(d.Measures.Length, d.Polygon.Width(), numbersPrecision)
}

func (d *GGDrawing) Height() float64 {
	return ConvertFromOneRound(d.Measures.Length, d.Polygon.Height(), numbersPrecision)
}

func (d *GGDrawing) GetPoints() []*Point {
	return d.GetPointsWithParams(d.Measures.Length, numbersPrecision)
}

func (d *GGDrawing) GetPointsWithParams(m Measure, precision int) []*Point {
	points := make([]*Point, len(d.Points))
	for i, p := range d.Points {
		points[i] = &Point{
			X: ConvertFromOneRound(m, p.X, precision),
			Y: ConvertFromOneRound(m, p.Y, precision),
		}
	}
	return points
}

func (d *GGDrawing) Scan(value interface{}) error {
	var data []byte
	switch value.(type) {
	case []byte:
		data = value.([]byte)
	case string:
		data = []byte(value.(string))
	default:
		return ErrWrongValueScanType
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
