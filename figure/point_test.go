package figure

import (
	"fmt"
	"github.com/go-test/deep"
	. "github.com/maxsid/goCeilings/value"
	"testing"
)

func TestDirectionCalculatorCircle(t *testing.T) {
	type args struct {
		cp        *Point
		distance  float64
		direction float64
	}
	type testData struct {
		name    string
		args    args
		wantErr bool
	}
	tests := make([]testData, 0)
	cp, distance := &Point{X: 0, Y: 0}, 150.0
	for i := 0.0; i <= 360; i += 15 {
		td := testData{
			name: fmt.Sprintf("%.2f degree", i),
			args: args{cp, distance, i}}
		tests = append(tests, td)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCalculatedPoint(&DirectionCalculator{
				Direction: Convert(Degree, Radian, tt.args.direction),
				Distance:  tt.args.distance,
			})
			if err := got.CalculateCoordinates(tt.args.cp); err != nil {
				t.Error(err)
			}
			pd := ConvertRound(Radian, Degree, pointDirection(cp, got), 2)
			if d := Round(tt.args.direction, 2); pd != d {
				t.Errorf("DirectionCalculator got %v, want %v", pd, d)
			}
		})
	}
}

func TestDirectionCalculator(t *testing.T) {
	type args struct {
		cp        *Point
		distance  float64
		direction float64
	}
	tests := []struct {
		name    string
		args    args
		want    *Point
		wantErr bool
	}{
		{
			name: "Simple 1",
			args: args{
				cp:        &Point{X: 0, Y: 0},
				distance:  125,
				direction: 0,
			},
			want: &Point{X: 125, Y: 0},
		},
		{
			name: "Simple 2",
			args: args{
				cp:        &Point{X: 15, Y: 30},
				distance:  125,
				direction: 90,
			},
			want: &Point{X: 15, Y: 155},
		},
		{
			name: "Simple 3",
			args: args{
				cp:        &Point{X: -10, Y: -30},
				distance:  125,
				direction: 180,
			},
			want: &Point{X: -135, Y: -30},
		},
		{
			name: "Simple 4",
			args: args{
				cp:        &Point{X: 0, Y: 0},
				distance:  140,
				direction: 350.011,
			},
			want: &Point{X: 137.878, Y: -24.284},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radDir := Convert(Degree, Radian, tt.args.direction)
			got := NewCalculatedPoint(&DirectionCalculator{
				Direction: radDir,
				Distance:  tt.args.distance,
			})
			if err := got.CalculateCoordinates(tt.args.cp); (err != nil) != tt.wantErr {
				t.Errorf("DirectionCalculator error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !comparePoints(tt.want, got, 0.1) {
				t.Errorf("DirectionCalculator got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAngleCalculator(t *testing.T) {
	type args struct {
		a, b      *Point
		distance  float64
		direction float64
	}
	tests := []struct {
		name    string
		args    args
		want    *Point
		wantErr bool
	}{
		{
			name: "Simple 1",
			args: args{
				a:         &Point{X: 0, Y: 0},
				b:         &Point{X: 0, Y: 125},
				distance:  27,
				direction: 90,
			},
			want: &Point{X: 27, Y: 125},
		},
		{
			name: "Simple 2",
			args: args{
				a:         &Point{X: 137.878, Y: -24.284},
				b:         &Point{X: 0, Y: 0},
				distance:  140,
				direction: 25.965,
			},
			want: &Point{X: 134.59, Y: 38.53},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radAngle := Convert(Degree, Radian, tt.args.direction)
			got := NewCalculatedPoint(&AngleCalculator{
				Angle:    radAngle,
				Distance: tt.args.distance,
			})
			if err := got.CalculateCoordinates(tt.args.a, tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("AngleCalculator error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !comparePoints(tt.want, got, 0.1) {
				t.Errorf("AngleCalculator = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAngleCalculator_ConvertToOne(t *testing.T) {
	type fields struct {
		Angle    float64
		Distance float64
	}
	type args struct {
		measures *FigureMeasures
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name:   "OK",
			fields: fields{Angle: 275, Distance: 436.4},
			args: args{measures: &FigureMeasures{
				Angle:  Degree,
				Length: Centimetre,
			}},
			want: fields{
				Angle:    ConvertToOne(Degree, 275),
				Distance: ConvertToOne(Centimetre, 436.4),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &AngleCalculator{
				Angle:    tt.fields.Angle,
				Distance: tt.fields.Distance,
			}
			ac.ConvertToOne(tt.args.measures)
			if !compareFloats(ac.Angle, tt.want.Angle, 0) && !compareFloats(ac.Distance, tt.want.Distance, 0) {
				t.Errorf("Got %+v, want %+v", ac, tt.want)
			}
		})
	}
}

func TestAngleCalculator_ConvertFromOne(t *testing.T) {
	type fields struct {
		Angle    float64
		Distance float64
	}
	type args struct {
		measures *FigureMeasures
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name:   "OK",
			fields: fields{Angle: ConvertToOne(Degree, 275), Distance: ConvertToOne(Centimetre, 436.4)},
			args: args{measures: &FigureMeasures{
				Angle:  Degree,
				Length: Centimetre,
			}},
			want: fields{
				Angle:    275,
				Distance: 436.4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &AngleCalculator{
				Angle:    tt.fields.Angle,
				Distance: tt.fields.Distance,
			}
			ac.ConvertFromOne(tt.args.measures)
			if !compareFloats(ac.Angle, tt.want.Angle, 0) && !compareFloats(ac.Distance, tt.want.Distance, 0) {
				t.Errorf("Got %+v, want %+v", ac, tt.want)
			}
		})
	}
}

func TestDirectionCalculator_ConvertToOne(t *testing.T) {
	type fields struct {
		Direction float64
		Distance  float64
	}
	type args struct {
		measures *FigureMeasures
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name:   "OK",
			fields: fields{Distance: 421.5, Direction: 140},
			args: args{measures: &FigureMeasures{
				Angle:  Degree,
				Length: Centimetre,
			}},
			want: fields{Distance: ConvertToOne(Centimetre, 421.5), Direction: ConvertToOne(Degree, 140)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DirectionCalculator{
				Direction: tt.fields.Direction,
				Distance:  tt.fields.Distance,
			}
			d.ConvertToOne(tt.args.measures)
			if !compareFloats(d.Direction, tt.want.Direction, 0) && !compareFloats(d.Distance, tt.want.Distance, 0) {
				t.Errorf("Got %+v, want %+v", d, tt.want)
			}
		})
	}
}

func TestDirectionCalculator_ConvertFromOne(t *testing.T) {
	type fields struct {
		Direction float64
		Distance  float64
	}
	type args struct {
		measures *FigureMeasures
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name:   "OK",
			fields: fields{Distance: ConvertToOne(Centimetre, 421.5), Direction: ConvertToOne(Degree, 140)},
			args: args{measures: &FigureMeasures{
				Angle:  Degree,
				Length: Centimetre,
			}},
			want: fields{Distance: 421.5, Direction: 140},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DirectionCalculator{
				Direction: tt.fields.Direction,
				Distance:  tt.fields.Distance,
			}
			d.ConvertFromOne(tt.args.measures)
			if !compareFloats(d.Direction, tt.want.Direction, 0) && !compareFloats(d.Distance, tt.want.Distance, 0) {
				t.Errorf("Got %+v, want %+v", d, tt.want)
			}
		})
	}
}

func TestPoint_UnmarshalJSON(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Point
		wantErr bool
	}{
		{
			name: "Nil calculator, only coordinates",
			args: args{bytes: []byte(`{"x":0,"y":1.25}`)},
			want: &Point{X: 0, Y: 1.25},
		},
		{
			name: "Direction calculator",
			args: args{bytes: []byte(`{"x":0.27,"y":1.25,"calculator":{"direction":0,"distance":0.27}}`)},
			want: &Point{X: 0.27, Y: 1.25, Calculator: &DirectionCalculator{0, 0.27}},
		},
		{
			name: "Angle calculator",
			args: args{bytes: []byte(`{"x":0.27,"y":1.71,"calculator":{"angle":4.7123889,"distance":0.46}}`)},
			want: &Point{X: 0.27, Y: 1.71, Calculator: &AngleCalculator{4.7123889, 0.46}},
		},
		{
			name:    "Wrong calculator",
			args:    args{bytes: []byte(`{"x":0.27,"y":1.71,"calculator":123}`)},
			wantErr: true,
		},
		{
			name:    "Bad data",
			args:    args{bytes: []byte(`-`)},
			wantErr: true,
		},
		{
			name:    "Bad data: Empty struct",
			args:    args{bytes: []byte(`{}`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Point{}
			if err := p.UnmarshalJSON(tt.args.bytes); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(p, tt.want); diff != nil {
				t.Errorf("UnmarshalJSON() -> %v)", diff)
			}
		})
	}
}
