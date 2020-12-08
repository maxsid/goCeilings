package figure

import (
	"github.com/go-test/deep"
	. "github.com/maxsid/goCeilings/value"
	"testing"
)

var example1 = []*Point{
	NewPoint(0, 0),
	NewPoint(0, 125),
	NewPoint(27, 125),
	NewPoint(27.01, 171),
	NewPoint(222.01, 169.98),
	NewPoint(225, 0)}
var example2 = []*Point{
	NewPoint(0, 0),
	NewPoint(0, 155),
	NewPoint(72.5, 155),
	NewPoint(72.5, 167.5),
	NewPoint(12.5, 167.51),
	NewPoint(12.53, 597.51),
	NewPoint(342.52, 599.99),
	NewPoint(345, 0),
}

func TestPolygon_Len(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Example 1",
			fields: fields{example1},
			want:   6,
		},
		{
			name:   "Example 2",
			fields: fields{example2},
			want:   8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if got := pol.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolygon_LastPoint(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Point
		wantErr bool
	}{
		{
			name:   "Example 1",
			fields: fields{Points: example1},
			want:   example1[5],
		},
		{
			name:   "Example 2",
			fields: fields{Points: example2},
			want:   example2[7],
		},
		{
			name:    "Err",
			fields:  fields{Points: []*Point{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			got, err := pol.LastPoint()
			if (err != nil) != tt.wantErr {
				t.Errorf("LastPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr || err != nil {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("LastPoint() -> %v", diff)
			}
		})
	}
}

func TestPolygon_Area(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name:   "Example 1",
			fields: fields{example1},
			want:   3.69,
		},
		{
			name:   "Example 2",
			fields: fields{example2},
			want:   19.95,
		},
		{
			name:   "Too small points",
			fields: fields{[]*Point{NewPoint(0, 0), NewPoint(1, 1)}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if got := ConvertRound(Centimetre2, Metre2, pol.Area(), 2); got != tt.want {
				t.Errorf("Area() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolygon_Perimeter(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name:   "Example 1",
			fields: fields{example1},
			want:   7.88,
		},
		{
			name:   "Example 2",
			fields: fields{example2},
			want:   20.05,
		},
		{
			name:   "Too small points",
			fields: fields{[]*Point{NewPoint(0, 0)}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if got := ConvertRound(Centimetre, Metre, pol.Perimeter(), 2); got != tt.want {
				t.Errorf("Perimeter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolygon_Width(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name:   "Example 1",
			fields: fields{example1},
			want:   225,
		},
		{
			name:   "Example 2",
			fields: fields{example2},
			want:   345,
		},
		{
			name:   "Too small points",
			fields: fields{[]*Point{NewPoint(0, 0)}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if got := pol.Width(); got != tt.want {
				t.Errorf("Width() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolygon_Height(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name:   "Example 1",
			fields: fields{example1},
			want:   171,
		},
		{
			name:   "Example 2",
			fields: fields{example2},
			want:   599.99,
		},
		{
			name:   "Too small points",
			fields: fields{[]*Point{NewPoint(0, 0)}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if got := pol.Height(); got != tt.want {
				t.Errorf("Height() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolygon_Rotation(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		o *Point
		a float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantS  []*Point
	}{
		{
			name:   "Simple",
			fields: fields{example1},
			args: args{
				o: example1[0],
				a: Convert(Degree, Radian, 45),
			},
			wantS: []*Point{
				{X: 0, Y: 0},
				{X: -88.388, Y: 88.388},
				{X: -69.289, Y: 107.487},
				{X: -101.816, Y: 140.014},
				{X: 36.791, Y: 277.179},
				{X: 159.099, Y: 159.099},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wPol := NewPolygon(tt.wantS...)
			pol := NewPolygon(tt.fields.Points...)
			pol.Rotation(tt.args.o, tt.args.a)
			if wp, p := pol.Perimeter(), wPol.Perimeter(); !compareFloats(wp, p, 0.2) {
				t.Errorf("Incorrect Perimeter. Got %f, want %f", p, wp)
			}
			if ws, s := pol.Area(), wPol.Area(); !compareFloats(ws, s, 0.3) {
				t.Errorf("Incorrect Area. Got %f, want %f", s, ws)
			}
			if err := comparePointsSlices(wPol.Points, pol.Points, 0.1); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPolygon_LastSide(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Segment
		wantErr bool
	}{
		{
			name:   "Example 1",
			fields: fields{Points: example1},
			want:   &Segment{A: example1[len(example1)-2], B: example1[len(example1)-1]},
		},
		{
			name:   "Example 2",
			fields: fields{Points: example2},
			want:   &Segment{A: example2[len(example2)-2], B: example2[len(example2)-1]},
		},
		{
			name:    "Err 0 points",
			fields:  fields{Points: []*Point{}},
			wantErr: true,
		},
		{
			name:    "Err 1 points",
			fields:  fields{Points: []*Point{example1[0]}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			got, err := pol.LastSide()
			if (err != nil) != tt.wantErr {
				t.Errorf("LastSide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("LastSide() -> %v", diff)
			}
		})
	}
}

func TestPolygon_AddPointByDirection(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		d float64
		a float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Point
		wantErr bool
	}{
		{
			name:    "No points error",
			fields:  fields{make([]*Point, 0)},
			args:    args{d: 10, a: 90},
			wantErr: true,
		},
		{
			name:   "One point",
			fields: fields{[]*Point{{X: 0, Y: 0}}},
			args:   args{125, 90},
			want:   []*Point{{X: 0, Y: 0}, {X: 0, Y: 125}},
		},
		{
			name:   "Two point",
			fields: fields{[]*Point{{X: 0, Y: 0}, {X: 0, Y: 125}}},
			args:   args{27, 0},
			want:   []*Point{{X: 0, Y: 0}, {X: 0, Y: 125}, {X: 27, Y: 125}},
		},
		{
			name:   "Three points",
			fields: fields{[]*Point{{X: 0, Y: 0}, {X: 0, Y: 125}, {X: 27, Y: 125}}},
			args:   args{46, 90},
			want:   []*Point{{X: 0, Y: 0}, {X: 0, Y: 125}, {X: 27, Y: 125}, {X: 27, Y: 171}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			rads := ConvertToOne(Degree, tt.args.a)
			if err := pol.AddPointByDirection(tt.args.d, rads); (err != nil) != tt.wantErr {
				t.Errorf("AddPointByDirection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, want := notPointerPoints(pol.Points), notPointerPoints(tt.want)
			if diff := deep.Equal(got, want); diff != nil {
				t.Errorf("AddPointByDirection() -> %v", diff)
			}
		})
	}
}

func TestPolygon_AddPointByAngle(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		distance float64
		angle    float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Point
		wantErr bool
	}{
		{
			name:    "Without points",
			fields:  fields{[]*Point{}},
			args:    args{125, 90},
			wantErr: true,
		},
		{
			name:    "1 point error",
			fields:  fields{[]*Point{{X: 0, Y: 0}}},
			args:    args{125, 90},
			wantErr: true,
		},
		{
			name:   "Two point",
			fields: fields{[]*Point{{X: 0, Y: 0}, {X: 0, Y: 125}}},
			args:   args{27, 90},
			want:   []*Point{{X: 0, Y: 0}, {X: 0, Y: 125}, {X: 27, Y: 125}},
		},
		{
			name:   "Three points",
			fields: fields{[]*Point{{X: 0, Y: 0}, {X: 0, Y: 125}, {X: 27, Y: 125}}},
			args:   args{46, 270},
			want:   []*Point{{X: 0, Y: 0}, {X: 0, Y: 125}, {X: 27, Y: 125}, {X: 27, Y: 171}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			rads := ConvertToOne(Degree, tt.args.angle)
			if err := pol.AddPointByAngle(tt.args.distance, rads); (err != nil) != tt.wantErr {
				t.Errorf("AddPointByAngle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			got, want := notPointerPoints(pol.Points), notPointerPoints(tt.want)
			if diff := deep.Equal(got, want); diff != nil {
				t.Errorf("AddPointByDirection() -> %v", diff)
			}
		})
	}
}

func notPointerPoints(points []*Point) []Point {
	out := make([]Point, len(points))
	for i, p := range points {
		out[i] = *NewPoint(Round(p.X, 2), Round(p.Y, 2))
	}
	return out
}

func TestPolygon_calculatePoints(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Point
		wantErr bool
	}{
		{
			name: "Combination",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{Calculator: &DirectionCalculator{Direction: ConvertToOne(Degree, 90), Distance: 125}},
				{Calculator: &AngleCalculator{Angle: ConvertToOne(Degree, 90), Distance: 27}},
				{Calculator: &AngleCalculator{Angle: ConvertToOne(Degree, 270), Distance: 46}},
				{Calculator: &AngleCalculator{Angle: ConvertToOne(Degree, 90), Distance: 195}},
				{Calculator: &AngleCalculator{Angle: ConvertToOne(Degree, 90), Distance: 170}},
			}},
			want: []*Point{
				{X: 0, Y: 0},
				{X: 0, Y: 125},
				{X: 27, Y: 125},
				{X: 27, Y: 171},
				{X: 222, Y: 171},
				{X: 222, Y: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if err := pol.CalculatePoints(); (err != nil) != tt.wantErr {
				t.Errorf("CalculatePoints() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if err := comparePointsSlices(pol.Points, tt.want, 0.01); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestPolygon_calculatePoint(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		index int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []*Point // checks only coordinates
		wantErr   bool
		wantPanic bool
	}{
		{
			name: "Incorrect index",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{X: 100, Y: 200},
			}},
			args:      args{index: -1},
			wantPanic: true,
		},
		{
			name: "Doesn't have calculator",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{X: 100, Y: 200},
			}},
			args:    args{index: 0},
			wantErr: true,
		},
		{
			name: "Square",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{Calculator: &DirectionCalculator{ConvertToOne(Degree, 90), 120}},
				{Calculator: &DirectionCalculator{0, 40}},
				{Calculator: &DirectionCalculator{ConvertToOne(Degree, 270), 120}},
			}},
			args: args{index: 1},
			want: []*Point{
				{X: 0, Y: 0},
				{X: 0, Y: 120},
				{X: 0, Y: 0},
				{X: 0, Y: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil && !tt.wantPanic {
					t.Errorf("Caught panic: %v", err)
				} else if err == nil && tt.wantPanic {
					t.Error("wantPanic = true, but panic is not caught")
				}
			}()
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if err := pol.calculatePoint(tt.args.index); (err != nil) != tt.wantErr {
				t.Errorf("calculatePoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				pol.RoundAllPoints(0)
				if err := comparePointsSlices(tt.want, pol.Points, 0.01); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestPolygon_RoundAllPoints(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		round int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*Point
	}{
		{
			name: "Round",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{X: 12.523, Y: 43.12},
				{X: 124.5356, Y: 43.1212},
				{X: 1212.000000000012, Y: 43},
			}},
			args: args{round: 2},
			want: []*Point{
				{X: 0, Y: 0},
				{X: 12.52, Y: 43.12},
				{X: 124.54, Y: 43.12},
				{X: 1212, Y: 43},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			pol.RoundAllPoints(tt.args.round)
			if err := comparePointsSlices(pol.Points, tt.want, 0); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPolygon_RoundCalculatedPoints(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		round int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*Point
	}{
		{
			name: "Round",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{X: 12.523, Y: 43.12},
				{X: 124.5356, Y: 43.1212, Calculator: &DirectionCalculator{}},
				{X: 124.5356, Y: 43.1212},
				{X: 1212.000000000012, Y: 43, Calculator: &DirectionCalculator{}},
				{X: 1212.000000000012, Y: 43},
			}},
			args: args{round: 2},
			want: []*Point{
				{X: 0, Y: 0},
				{X: 12.523, Y: 43.12},
				{X: 124.54, Y: 43.12},
				{X: 124.5356, Y: 43.1212},
				{X: 1212, Y: 43},
				{X: 1212.000000000012, Y: 43},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			pol.RoundCalculatedPoints(tt.args.round)
			if err := comparePointsSlices(pol.Points, tt.want, 0); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPolygon_SetPoint(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	type args struct {
		i     int
		point *Point
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Point
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{Points: []*Point{
				{X: 0, Y: 0},
				{X: 45, Y: 0},
				{X: 45, Y: 45, Calculator: &AngleCalculator{Angle: ConvertToOne(Degree, 90), Distance: 45}},
				{X: 0, Y: 45, Calculator: &AngleCalculator{Angle: ConvertToOne(Degree, 90), Distance: 45}}},
			},
			args: args{
				i:     1,
				point: &Point{Calculator: &DirectionCalculator{Direction: ConvertToOne(Degree, 90), Distance: 45}},
			},
			want: []*Point{
				{X: 0, Y: 0},
				{X: 0, Y: 45},
				{X: 45, Y: 45},
				{X: 45, Y: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if err := pol.SetPoint(tt.args.i, tt.args.point); (err != nil) != tt.wantErr {
				t.Errorf("SetPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			pol.RoundAllPoints(2)
			if err := comparePointsSlices(pol.Points, tt.want, 0.01); err != nil {
				t.Errorf("SetPoint() got wrong coordinates. Got = %v, want %v", pol.Points, tt.want)
			}
		})
	}
}

func TestPolygon_DiagonalsLen(t *testing.T) {
	type fields struct {
		Points []*Point
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "4 points",
			fields: fields{Points: []*Point{{}, {}, {}, {}}},
			want:   2,
		},
		{
			name:   "3 points",
			fields: fields{Points: []*Point{{}, {}, {}}},
			want:   0,
		},
		{
			name:   "1 point",
			fields: fields{Points: []*Point{{}}},
			want:   0,
		},
		{
			name:   "0 point",
			fields: fields{Points: []*Point{}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := &Polygon{
				Points: tt.fields.Points,
			}
			if got := pol.DiagonalsLen(); got != tt.want {
				t.Errorf("DiagonalsLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
