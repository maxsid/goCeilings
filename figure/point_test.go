package figure

import (
	"fmt"
	. "github.com/maxsid/goCeilings/value"
	"testing"
)

func TestNewPointByDirectionAndPointDirection(t *testing.T) {
	type args struct {
		cp        *Point
		distance  float64
		direction float64
	}
	type testData struct {
		name string
		args args
	}
	tests := make([]testData, 0)
	cp, distance := &Point{0, 0}, 150.0
	for i := 0.0; i <= 360; i += 15 {
		td := testData{
			name: fmt.Sprintf("%.2f degree", i),
			args: args{cp, distance, i}}
		tests = append(tests, td)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPointByDirection(tt.args.cp, tt.args.distance, Convert(Degree, Radian, tt.args.direction))
			pd := ConvertRound(Radian, Degree, PointDirection(cp, got), 2)
			if d := Round(tt.args.direction, 2); pd != d {
				t.Errorf("NewPointByDirection()->PointDirection() = %v, want %v", pd, d)
			}
		})
	}
}

func TestNewPointByDirection(t *testing.T) {
	type args struct {
		cp        *Point
		distance  float64
		direction float64
	}
	tests := []struct {
		name string
		args args
		want *Point
	}{
		{
			name: "Simple 1",
			args: args{
				cp:        &Point{0, 0},
				distance:  125,
				direction: 0,
			},
			want: &Point{125, 0},
		},
		{
			name: "Simple 2",
			args: args{
				cp:        &Point{15, 30},
				distance:  125,
				direction: 90,
			},
			want: &Point{15, 155},
		},
		{
			name: "Simple 3",
			args: args{
				cp:        &Point{-10, -30},
				distance:  125,
				direction: 180,
			},
			want: &Point{-135, -30},
		},
		{
			name: "Simple 4",
			args: args{
				cp:        &Point{0, 0},
				distance:  140,
				direction: 350.011,
			},
			want: &Point{137.878, -24.284},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radDir := Convert(Degree, Radian, tt.args.direction)
			if got := NewPointByDirection(tt.args.cp, tt.args.distance, radDir); !comparePoints(tt.want, got, 0.1) {
				t.Errorf("NewPointByDirection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPointByDirectionFromLine(t *testing.T) {
	type args struct {
		l         *Segment
		distance  float64
		direction float64
	}
	tests := []struct {
		name string
		args args
		want *Point
	}{
		{
			name: "Simple 1",
			args: args{
				l: &Segment{
					&Point{0, 0},
					&Point{0, 125},
				},
				distance:  27,
				direction: 90,
			},
			want: &Point{27, 125},
		},
		{
			name: "Simple 2",
			args: args{
				l: &Segment{
					&Point{137.878, -24.284},
					&Point{0, 0},
				},
				distance:  140,
				direction: 25.965,
			},
			want: &Point{134.59, 38.53},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radDir := Convert(Degree, Radian, tt.args.direction)
			if got := NewPointByDirectionFromSegment(tt.args.l, tt.args.distance, radDir); !comparePoints(tt.want, got, 0.1) {
				t.Errorf("NewPointByDirectionFromSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}
