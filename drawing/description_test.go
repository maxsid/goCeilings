package drawing

import (
	"testing"

	"github.com/go-test/deep"
)

func TestNewUnionDescription(t *testing.T) {
	type args struct {
		descriptions []*Description
	}
	tests := []struct {
		name string
		args args
		want *Description
	}{
		{
			name: "Empty",
			args: args{descriptions: []*Description{}},
			want: &Description{},
		},
		{
			name: "Two",
			args: args{descriptions: []*Description{
				{{"a", "b"}, {"c", "d"}},
				{{"e", "f"}, {"g", "h"}},
			}},
			want: &Description{{"a", "b"}, {"c", "d"}, {"e", "f"}, {"g", "h"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUnionDescription(tt.args.descriptions...)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("NewUnionDescription() -> %v", diff)
			}
		})
	}
}

func TestDescription_PushBack(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		d    Description
		want Description
		args args
	}{
		{
			name: "OK",
			args: args{key: "c", value: "d"},
			d:    Description{{"a", "b"}},
			want: Description{{"a", "b"}, {"c", "d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.PushBack(tt.args.key, tt.args.value)
			if diff := deep.Equal(tt.d, tt.want); diff != nil {
				t.Errorf("PushBack() -> %v", diff)
			}
		})
	}
}

func TestDescription_ToStringSlice(t *testing.T) {
	tests := []struct {
		name string
		d    Description
		want []string
	}{
		{
			name: "OK",
			d:    Description{{"a", "b"}, {"c", "d"}},
			want: []string{"a: b", "c: d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.d.ToStringSlice()
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("ToStringSlice() -> %v", diff)
			}
		})
	}
}
