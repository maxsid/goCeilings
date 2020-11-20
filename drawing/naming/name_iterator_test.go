package naming

import "testing"

func TestNameIterator_Next(t *testing.T) {
	type fields struct {
		start   int32
		end     int32
		counter uint
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "First symbol",
			fields: fields{
				start:   'a',
				end:     'z',
				counter: 0,
			},
			want: "a",
		},
		{
			name: "Last symbol",
			fields: fields{
				start:   'a',
				end:     'z',
				counter: 25,
			},
			want: "z",
		},
		{
			name: "With nums first symbol",
			fields: fields{
				start:   'a',
				end:     'z',
				counter: 26,
			},
			want: "a1",
		},
		{
			name: "With nums last symbol",
			fields: fields{
				start:   'a',
				end:     'z',
				counter: 51,
			},
			want: "z1",
		},
		{
			name: "With nums big Counter",
			fields: fields{
				start:   'A',
				end:     'Z',
				counter: 456455,
			},
			want: "Z17555",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ni := &NameIterator{
				start:   tt.fields.start,
				end:     tt.fields.end,
				Counter: tt.fields.counter,
			}
			if got := ni.Next(); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
