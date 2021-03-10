package vector

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/maxsid/goCeilings/figure"
)

func TestDrawing_Draw(t *testing.T) {
	var example = []*Point{
		{X: 0, Y: 0},
		{X: 0, Y: 1.55},
		{X: 0.725, Y: 1.55},
		{X: 0.725, Y: 1.675},
		{X: 0.125, Y: 1.6751},
		{X: 0.1253, Y: 5.9751},
		{X: 3.4252, Y: 5.9999},
		{X: 3.45, Y: 0},
	}
	draw := NewDrawing()
	draw.Points = example
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	f, err := ioutil.TempFile(homeDir, "*.svg")
	if err != nil {
		panic(err)
	}
	draw.Draw(f)
	_ = f.Close()
	log.Printf("See %s file", f.Name())
}
