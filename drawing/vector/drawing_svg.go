package vector

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	svg "github.com/ajstarks/svgo/float"
	"github.com/maxsid/goCeilings/drawing/naming"
	. "github.com/maxsid/goCeilings/figure"
	. "github.com/maxsid/goCeilings/value"
)

const (
	marginLeft, marginRight, marginTop, marginDown float64 = 40, 40, 40, 40
	marginHorizontal, marginVertical                       = marginLeft + marginRight, marginTop + marginDown
	marginNotes                                    float64 = 20
	drawingWidth, drawingHeight                    float64 = 1200, 800
	pointSize                                      float64 = 2
	marginLetterX, marginLetterY                   float64 = 4, 20
)

type SVGDrawing struct {
	Points                          []*Point
	Notes                           []string
	lengthMeasure, perimeterMeasure Measure
	areaMeasure, angleMeasure       Measure
}

func NewDrawing() *SVGDrawing {
	return &SVGDrawing{
		Points:           make([]*Point, 0),
		Notes:            make([]string, 0),
		lengthMeasure:    Centimetre,
		perimeterMeasure: Metre,
		areaMeasure:      Metre2,
		angleMeasure:     Degree,
	}
}

func (d *SVGDrawing) Draw(wr io.Writer) {
	d.addAreaToNotes()
	d.addPerimeterToNotes()
	d.addSidesToNotes()
	d.addPointsToNotes()
	xs, ys := d.getXYs()
	scale := d.calcScale()
	canvas := svg.New(wr)
	canvas.Start(drawingWidth, drawingHeight)
	canvas.Translate(marginLeft, marginTop)
	canvas.Scale(scale)
	canvas.Group()
	canvas.Polygon(xs, ys, `stroke="#ff0000"`, `fill="#00000012"`, `stroke-width="2"`)
	d.drawPoints(canvas, xs, ys)
	d.drawSidesLengths(canvas, xs, ys)
	canvas.Gend()
	canvas.Group()
	d.drawNotes(canvas)
	canvas.Gend()
	canvas.Gend()
	canvas.Gend()
	canvas.End()
}

func (d *SVGDrawing) getXYs() ([]float64, []float64) {
	pol := Polygon{Points: d.Points}
	polHeight := pol.Height()
	xs, ys := make([]float64, len(d.Points)), make([]float64, len(d.Points))
	for i, p := range d.Points {
		xs[i] = ConvertFromOneRound(d.lengthMeasure, p.X, 2)
		ys[i] = ConvertFromOneRound(d.lengthMeasure, polHeight-p.Y, 2)
	}
	return xs, ys
}

func (d *SVGDrawing) drawSidesLengths(canvas *svg.SVG, xs, ys []float64) {
	xsLen, pol := len(xs), Polygon{Points: d.Points}
	sides := pol.Sides()
	for i := 0; i < xsLen; i++ {
		x1, y1, x2, y2 := xs[i], ys[i], xs[(i+1)%xsLen], ys[(i+1)%xsLen]
		x, y := (x1+x2)/2, (y1+y2)/2
		dist := ConvertFromOneRound(d.lengthMeasure, sides[i].Distance(), 2)
		canvas.Text(x, y, fmt.Sprintf("%v", dist))
	}
}

func (d *SVGDrawing) drawPoints(canvas *svg.SVG, xs, ys []float64) {
	letterIter := naming.NewNameIterator('A', 'Z')
	for i := 0; i < len(xs); i++ {
		canvas.CenterRect(xs[i], ys[i], pointSize, pointSize, `fill="black"`)
		canvas.Text(xs[i]+marginLetterX, ys[i]+marginLetterY, letterIter.Next())
	}
}

func (d *SVGDrawing) calcScale() float64 {
	pol := Polygon{Points: d.Points}
	wScale := (drawingWidth - marginHorizontal) / ConvertFromOne(d.lengthMeasure, pol.Width())
	hScale := (drawingHeight - marginVertical) / ConvertFromOne(d.lengthMeasure, pol.Height())
	scale := wScale
	if wScale > hScale {
		scale = hScale
	}
	return scale
}

func (d *SVGDrawing) addSidesToNotes() {
	letterIter := naming.NewNameIterator('A', 'Z')
	l1, l2, pol := letterIter.Next(), letterIter.Next(), Polygon{Points: d.Points}
	note := make([]string, 0)
	for _, s := range pol.Sides() {
		dist := ConvertFromOneRound(d.lengthMeasure, s.Distance(), 2)
		note = append(note, fmt.Sprintf("%s%s=%v", l1, l2, dist))
		l1, l2 = l2, letterIter.Next()
	}
	d.Notes = append(d.Notes, "Sides: "+strings.Join(note, ", "))
}

func (d *SVGDrawing) addPointsToNotes() {
	letterIter := naming.NewNameIterator('A', 'Z')
	s, pol := make([]string, 0), Polygon{Points: d.Points}
	for _, p := range pol.Points {
		l := letterIter.Next()
		x, y := ConvertFromOneRound(d.lengthMeasure, p.X, 2), ConvertFromOneRound(d.lengthMeasure, p.Y, 2)
		s = append(s, fmt.Sprintf("%s=(%v;%v)", l, x, y))
	}
	d.Notes = append(d.Notes, "Points: "+strings.Join(s, ", "))
}

func (d *SVGDrawing) addAreaToNotes() {
	pol := Polygon{Points: d.Points}
	s := ConvertFromOneRound(d.areaMeasure, pol.Area(), 2)
	d.Notes = append(d.Notes, "Area: "+fmt.Sprintf("%v", s))
}

func (d *SVGDrawing) addPerimeterToNotes() {
	pol := Polygon{Points: d.Points}
	p := ConvertFromOneRound(d.perimeterMeasure, pol.Perimeter(), 2)
	d.Notes = append(d.Notes, "Perimeter: "+fmt.Sprintf("%v", p))
}

func (d *SVGDrawing) drawNotes(canvas *svg.SVG) {
	pol := Polygon{Points: d.Points}
	w := ConvertFromOne(d.lengthMeasure, pol.Width())
	canvas.Textlines(w+marginNotes, 0, d.Notes, 16, 20, "", "")
}

func (d *SVGDrawing) DrawBytes() (bOut []byte) {
	buf := bytes.NewBuffer(bOut)
	d.Draw(buf)
	return
}
