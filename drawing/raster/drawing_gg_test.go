package raster

import (
	"bytes"
	"github.com/fogleman/gg"
	. "github.com/maxsid/goCeilings/figure"
	"golang.org/x/image/colornames"
	"image/png"
	"reflect"
	"testing"
)

func drawExamples(examples [][]*Point, t *testing.T) {
	for _, ps := range examples {
		for _, drawDesc := range []bool{true, false} {
			draw := NewEmptyGGDrawing()
			draw.Points = ps
			imgBytes, err := draw.Draw(drawDesc)
			if err != nil {
				t.Error(err)
			}
			if _, err := png.Decode(bytes.NewBuffer(imgBytes)); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestGGDrawing_Draw(t *testing.T) {
	examples := [][]*Point{
		{
			{X: 0, Y: 0},
			{X: 0, Y: 1.55},
			{X: 0.725, Y: 1.55},
			{X: 0.725, Y: 1.675},
			{X: 0.125, Y: 1.6751},
			{X: 0.1253, Y: 5.9751},
			{X: 3.4252, Y: 5.9999},
			{X: 3.45, Y: 0},
		},
		{
			{X: 0, Y: 0},
			{X: 0, Y: 1.25},
			{X: 0.27, Y: 1.25},
			{X: 0.2701, Y: 1.71},
			{X: 2.2201, Y: 1.6998},
			{X: 2.25, Y: 0},
		},
	}
	drawExamples(examples, t)
}

func Test_setBackground(t *testing.T) {
	ctx := gg.NewContext(16, 16)
	setBackground(ctx)
	c := ctx.Image().At(8, 8)
	if c != colornames.White {
		t.Error("Color must be white!")
	}
}

func Test_calcImageSize(t *testing.T) {
	type args struct {
		drawScale float64
		polW      float64
		polH      float64
		drawDesc  bool
	}
	tests := []struct {
		name  string
		args  args
		wantW int
		wantH int
	}{
		{
			name:  "Without desc",
			args:  args{drawScale: 1.5, polW: 100, polH: 100, drawDesc: false},
			wantH: marginVertical + 150,
			wantW: marginHorizontal + 150,
		},
		{
			name:  "With desc",
			args:  args{drawScale: 0.5, polW: 100, polH: 100, drawDesc: true},
			wantH: marginVertical + 50,
			wantW: marginHorizontal*2 + descriptionWidth + 50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := calcImageSize(tt.args.drawScale, tt.args.polW, tt.args.polH, tt.args.drawDesc)
			if gotW != tt.wantW {
				t.Errorf("calcImageSize() gotW = %v, want %v", gotW, tt.wantW)
			}
			if gotH != tt.wantH {
				t.Errorf("calcImageSize() gotH = %v, want %v", gotH, tt.wantH)
			}
		})
	}
}

func Test_calcDrawScale(t *testing.T) {
	type args struct {
		polW float64
		polH float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Height scale",
			args: args{polH: 1000, polW: 500},
			want: float64(drawingHeight-marginVertical) / 1000.0,
		},
		{
			name: "Width scale",
			args: args{polH: 500, polW: 1000},
			want: float64(drawingWidth-marginHorizontal) / 1000.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcDrawScale(tt.args.polW, tt.args.polH); got != tt.want {
				t.Errorf("calcDrawScale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getXYOnDrawing(t *testing.T) {
	type args struct {
		p       *Point
		offsetX float64
		offsetY float64
		scale   float64
	}
	tests := []struct {
		name  string
		args  args
		wantX float64
		wantY float64
	}{
		{
			name:  "OK",
			args:  args{scale: 0.5, p: NewPoint(10, 20), offsetY: 10, offsetX: 15},
			wantX: marginLeft + 15 + 5,
			wantY: marginDown + 10 + 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := getXYOnDrawing(tt.args.p, tt.args.offsetX, tt.args.offsetY, tt.args.scale)
			if gotX != tt.wantX {
				t.Errorf("getXYOnDrawing() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("getXYOnDrawing() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}

func TestGGDrawing_Draw_Sizes(t *testing.T) {
	type fields struct {
		points []*Point
	}
	type args struct {
		drawDesc bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Too few points",
			fields:  fields{points: []*Point{{}, {}}},
			args:    args{drawDesc: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &GGDrawing{Polygon: *NewPolygon(tt.fields.points...)}
			got, err := d.Draw(tt.args.drawDesc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Draw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Draw() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGGDrawing_updateOffset(t *testing.T) {
	type fields struct {
		points []*Point
	}
	type args struct {
		scale float64
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantOffsetX float64
		wantOffsetY float64
	}{
		{
			name: "Positive",
			fields: fields{points: []*Point{
				{X: -30, Y: -40},
				{X: 0, Y: 0},
				{X: 30, Y: 40},
			}},
			args:        args{scale: 0.5},
			wantOffsetX: 15,
			wantOffsetY: 20,
		},
		{
			name: "Zero",
			fields: fields{points: []*Point{
				{X: 0, Y: 0},
				{X: 30, Y: 40},
			}},
			args:        args{scale: 1.5},
			wantOffsetX: 0,
			wantOffsetY: 0,
		},
		{
			name: "Negative",
			fields: fields{points: []*Point{
				{X: 10, Y: 15},
				{X: 30, Y: 40},
			}},
			args:        args{scale: 1},
			wantOffsetX: -10,
			wantOffsetY: -15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &GGDrawing{Polygon: *NewPolygon(tt.fields.points...)}
			d.updateOffset(tt.args.scale)
			if d.offsetX != tt.wantOffsetX {
				t.Errorf("OffsetX: got %v, want %v", d.offsetX, tt.wantOffsetX)
			}
			if d.offsetY != tt.wantOffsetY {
				t.Errorf("OffsetY: got %v, want %v", d.offsetY, tt.wantOffsetY)
			}
		})
	}
}
