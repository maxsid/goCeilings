package figure

import (
	. "github.com/maxsid/goCeilings/value"
	"reflect"
	"testing"
)

var example1 = []*Point{
	{0, 0},
	{0, 125},
	{27, 125},
	{27.01, 171},
	{222.01, 169.98},
	{225, 0}}
var example2 = []*Point{
	{0, 0},
	{0, 155},
	{72.5, 155},
	{72.5, 167.5},
	{12.5, 167.51},
	{12.53, 597.51},
	{342.52, 599.99},
	{345, 0},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastPoint() got = %v, want %v", got, tt.want)
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
			fields: fields{[]*Point{{0, 0}, {1, 1}}},
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
			fields: fields{[]*Point{{0, 0}}},
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
			fields: fields{[]*Point{{0, 0}}},
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
			fields: fields{[]*Point{{0, 0}}},
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
				{0, 0},
				{-88.388, 88.388},
				{-69.289, 107.487},
				{-101.816, 140.014},
				{36.791, 277.179},
				{159.099, 159.099},
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
			for i := 0; i < pol.Len(); i++ {
				if !comparePoints(wPol.Points[i], pol.Points[i], 0.1) {
					t.Errorf("Incorrect point %d. Got %v, want %v", i, pol.Points[i], wPol.Points[i])
				}
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastSide() got = %v, want %v", got, tt.want)
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
			fields: fields{[]*Point{{0, 0}}},
			args:   args{125, 90},
			want:   []*Point{{0, 0}, {0, 125}},
		},
		{
			name:   "Two point",
			fields: fields{[]*Point{{0, 0}, {0, 125}}},
			args:   args{27, 0},
			want:   []*Point{{0, 0}, {0, 125}, {27, 125}},
		},
		{
			name:   "Three points",
			fields: fields{[]*Point{{0, 0}, {0, 125}, {27, 125}}},
			args:   args{46, 90},
			want:   []*Point{{0, 0}, {0, 125}, {27, 125}, {27, 171}},
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
			if !tt.wantErr {
				got, want := notPointerPoints(pol.Points), notPointerPoints(tt.want)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("AddPointByDirection() wrong result = %v, want %v", got, want)
				}
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
			fields:  fields{[]*Point{{0, 0}}},
			args:    args{125, 90},
			wantErr: true,
		},
		{
			name:   "Two point",
			fields: fields{[]*Point{{0, 0}, {0, 125}}},
			args:   args{27, 90},
			want:   []*Point{{0, 0}, {0, 125}, {27, 125}},
		},
		{
			name:   "Three points",
			fields: fields{[]*Point{{0, 0}, {0, 125}, {27, 125}}},
			args:   args{46, 270},
			want:   []*Point{{0, 0}, {0, 125}, {27, 125}, {27, 171}},
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
			if !tt.wantErr {
				got, want := notPointerPoints(pol.Points), notPointerPoints(tt.want)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("AddPointByDirection() wrong result = %v, want %v", got, want)
				}
			}
		})
	}
}

func notPointerPoints(points []*Point) []Point {
	out := make([]Point, len(points))
	for i, p := range points {
		out[i] = *p.GetRoundedPoint(2)
	}
	return out
}
