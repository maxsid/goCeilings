package figure

import (
	"testing"

	. "github.com/maxsid/goCeilings/value"
)

func TestLine_Distance(t *testing.T) {
	type fields struct {
		A Point
		B Point
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
		round  int
	}{
		{
			name: "Simple test 1",
			fields: fields{
				A: Point{X: 9, Y: 7},
				B: Point{X: 3, Y: 2},
			},
			want:  7.8102,
			round: 4,
		},
		{
			name: "Simple test 2",
			fields: fields{
				A: Point{X: 3, Y: 2},
				B: Point{X: 9, Y: 7},
			},
			want:  7.8102,
			round: 4,
		},
		{
			name: "Zero test",
			fields: fields{
				A: Point{},
				B: Point{},
			},
			want:  0,
			round: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Segment{
				A: &tt.fields.A,
				B: &tt.fields.B,
			}
			if got := Round(l.Distance(), tt.round); got != tt.want {
				t.Errorf("Distance() = %v, want %v", got, tt.want)
			}
		})
	}
}
