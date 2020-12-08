package figure

import (
	"fmt"
	"github.com/go-test/deep"
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
				a: &Point{X: 0, Y: 0},
				b: &Point{X: 0, Y: 155},
			},
			want: 90,
		},
		{
			name: "Simple 2",
			args: args{
				a: &Point{X: 0, Y: 155},
				b: &Point{X: 0, Y: 0},
			},
			want: 270,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pointDirection(tt.args.a, tt.args.b)
			got = ConvertRound(Radian, Degree, got, 2)
			if got != tt.want {
				t.Errorf("pointDirection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sliceOfForwardElements(t *testing.T) {
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	type args struct {
		i     int
		slice interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "Not a slice error",
			args:    args{5, &slice},
			wantErr: true,
		},
		{
			name: "Middle",
			args: args{5, slice},
			want: []int{6, 7, 8, 9, 0, 1, 2, 3, 4},
		},
		{
			name: "Front",
			args: args{0, slice},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "Back",
			args: args{9, slice},
			want: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sliceOfForwardElements(tt.args.i, tt.args.slice)
			if (err != nil) != tt.wantErr {
				t.Errorf("sliceOfForwardElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("sliceOfForwardElements() -> %v", diff)
			}
		})
	}
}

// comparePointsSlices compares all points in two slides with comparePoints function.
// Accuracy is a maximum value of difference between coordinates.
// If the difference more than accuracy, function returns false.
// Used for tests only.
func comparePointsSlices(a, b []*Point, accuracy float64) error {
	if len(a) != len(b) {
		return fmt.Errorf("sizes of slices is not equal: len(a) = %d, len(b) = %d", len(a), len(b))
	}
	for i := 0; i < len(a); i++ {
		if !comparePoints(a[i], b[i], accuracy) {
			return fmt.Errorf("points on position %d is not in accuracy: a = %v, b = %v", i, a[i], b[i])
		}
	}
	return nil
}

// comparePoints compares two points with compareFloats function.
// Accuracy is a maximum value of difference between coordinates.
// If the difference more than accuracy, function returns false.
// Used for tests only.
func comparePoints(pw, p *Point, accuracy float64) bool {
	return compareFloats(pw.X, p.X, accuracy) && compareFloats(pw.Y, p.Y, accuracy)
}

// compareFloats compares two floats.
// Accuracy is a maximum value of difference between numbers.
// If the difference more than accuracy, function returns false.
// Used for tests only.
func compareFloats(a, b, accuracy float64) bool {
	d := math.Abs(a - b)
	if d > accuracy {
		return false
	}
	return true
}
