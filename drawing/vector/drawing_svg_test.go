package vector

import (
	. "github.com/maxsid/goCeilings/figure"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestDrawing_Draw(t *testing.T) {
	var example = []*Point{
		{0, 0},
		{0, 1.55},
		{0.725, 1.55},
		{0.725, 1.675},
		{0.125, 1.6751},
		{0.1253, 5.9751},
		{3.4252, 5.9999},
		{3.45, 0},
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
