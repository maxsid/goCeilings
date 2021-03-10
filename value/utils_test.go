package value

import "testing"

func TestRound(t *testing.T) {
	type args struct {
		v     float64
		round int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Simple",
			args: args{
				v:     1234.56789,
				round: 2,
			},
			want: 1234.57,
		},
		{
			name: "Zero round",
			args: args{
				v:     1234.56789,
				round: 0,
			},
			want: 1235,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Round(tt.args.v, tt.args.round); got != tt.want {
				t.Errorf("GetRoundedPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	type args struct {
		a Measure
		b Measure
		v float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "M to CM",
			args: args{
				a: Metre,
				b: Centimetre,
				v: 12,
			},
			want: 1200,
		},
		{
			name: "MM to CM",
			args: args{
				a: Millimetre,
				b: Centimetre,
				v: 12,
			},
			want: 1.2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.args.a, tt.args.b, tt.args.v); got != tt.want {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertRound(t *testing.T) {
	type args struct {
		a     Measure
		b     Measure
		v     float64
		round int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "M to KM",
			args: args{
				a:     Metre,
				b:     Kilometre,
				v:     124,
				round: 2,
			},
			want: 0.12,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertRound(tt.args.a, tt.args.b, tt.args.v, tt.args.round); got != tt.want {
				t.Errorf("ConvertRound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDigitsAfterDot(t *testing.T) {
	type args struct {
		v float64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Without digits",
			args: args{24},
			want: 0,
		},
		{
			name: "2 Digits",
			args: args{3.14},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DigitsAfterDot(tt.args.v); got != tt.want {
				t.Errorf("DigitsAfterDot() = %v, want %v", got, tt.want)
			}
		})
	}
}
