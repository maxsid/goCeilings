package gorm

import "testing"

func Test_getHexHash(t *testing.T) {
	type args struct {
		v    string
		salt string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Simple 1",
			args: args{v: "misterpropper", salt: "agakakskazhesh"},
			want: "a5bf581d04824b1a0faf858748a9d99b312e87a34df49cefbde4d44ae7a365ad",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHexHash(tt.args.v, tt.args.salt); got != tt.want {
				t.Errorf("getHexHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
