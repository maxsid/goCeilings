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
			{0, 0},
			{0, 1.55},
			{0.725, 1.55},
			{0.725, 1.675},
			{0.125, 1.6751},
			{0.1253, 5.9751},
			{3.4252, 5.9999},
			{3.45, 0},
		},
		{
			{0, 0},
			{0, 1.25},
			{0.27, 1.25},
			{0.2701, 1.71},
			{2.2201, 1.6998},
			{2.25, 0},
		},
	}
	drawExamples(examples, t)
}
