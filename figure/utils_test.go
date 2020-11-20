package figure

import (
	. "github.com/maxsid/goCeilings/value"
	"math"
	"testing"
)

func TestPointDirection(t *testing.T) {
	type args struct {
		a *Point
		b *Point
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Simple 1",
			args: args{
				a: &Point{0, 0},
				b: &Point{0, 155},
			},
			want: 90,
		},
		{
			name: "Simple 2",
			args: args{
				a: &Point{0, 155},
				b: &Point{0, 0},
			},
			want: 270,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PointDirection(tt.args.a, tt.args.b)
			got = ConvertRound(Radian, Degree, got, 2)
			if got != tt.want {
				t.Errorf("PointDirection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func comparePoints(pw, p *Point, accuracy float64) bool {
	x, y := math.Abs(pw.X-p.X), math.Abs(pw.Y-p.Y)
	if x > accuracy || y > accuracy {
		return false
	}
	return true
}

func compareFloats(a, b, accuracy float64) bool {
	d := math.Abs(a - b)
	if d > accuracy {
		return false
	}
	return true
}
