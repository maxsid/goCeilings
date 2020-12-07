package raster

import (
	"fmt"
	"github.com/fogleman/gg"
	. "github.com/maxsid/goCeilings/figure"
	"os"
	"testing"
)

func drawExamples(examples [][]*Point, t *testing.T) {
	for i, ps := range examples {
		for _, drawDesc := range []bool{true, false} {
			draw := NewEmptyGGDrawing()
			draw.Points = ps
			img, err := draw.Draw(drawDesc)
			if err != nil {
				t.Error(err)
			}
			home, _ := os.UserHomeDir()
			filename := fmt.Sprintf("%s/test%d_%v.png", home, i, drawDesc)
			if err = gg.SavePNG(filename, img); err != nil {
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
