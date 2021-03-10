package generator

import (
	"math/rand"
	"testing"

	"github.com/go-test/deep"
)

func Test_getSymbolsSlice(t *testing.T) {
	type args struct {
		start byte
		end   byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "A-Z",
			args: args{'A', 'Z'},
			want: []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		},
		{
			name: "0-9",
			args: args{'0', '9'},
			want: []byte("0123456789"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSymbolsSlice(tt.args.start, tt.args.end)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("getSymbolsSlice() -> %v", diff)
			}
		})
	}
}

func Test_getRandomString(t *testing.T) {
	rand.Seed(0)
	type args struct {
		symbols []byte
		count   int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "OK",
			args: args{
				symbols: []byte("0123456789"),
				count:   10,
			},
			want: []byte("4436567788"),
		},
		{
			name: "Zero count",
			args: args{
				symbols: []byte("0123456789"),
				count:   0,
			},
			want: []byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRandomString(tt.args.symbols, tt.args.count)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("getRandomString() -> %v", diff)
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	rand.Seed(0)
	type args struct {
		lc int
		uc int
		dc int
		sc int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{
				lc: 10,
				uc: 10,
				dc: 10,
				sc: 10,
			},
			want: "<O8zik)ENB!Syz*%2*a0K$1ECc2L0u6A{<b(61h5",
		},
		{
			name: "Zero counts",
			args: args{
				lc: 0,
				uc: 0,
				dc: 0,
				sc: 0,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GeneratePassword(tt.args.lc, tt.args.uc, tt.args.dc, tt.args.sc); got != tt.want {
				t.Errorf("GeneratePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
