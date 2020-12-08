package api

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-test/deep"
	"github.com/maxsid/goCeilings/value"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func Test_readListStatData(t *testing.T) {
	type args struct {
		vars   url.Values
		amount uint
	}
	tests := []struct {
		name    string
		args    args
		want    *ListStatData
		wantErr bool
	}{
		{
			name: "Empty vars",
			args: args{vars: url.Values{}, amount: 10},
			want: &ListStatData{
				PageLimit: defaultPageLimit,
				Page:      defaultPage,
				Amount:    10,
				Pages:     1,
			},
		},
		{
			name: "Page is more than pages",
			args: args{
				amount: 133,
				vars: url.Values{
					string(urlParamPageLimit): []string{"60"},
					string(urlParamPage):      []string{"123"},
				},
			},
			want: &ListStatData{
				PageLimit: 60,
				Page:      3,
				Amount:    133,
				Pages:     3,
			},
		},
		{
			name: "Page limit",
			args: args{
				amount: 10,
				vars:   url.Values{string(urlParamPageLimit): []string{"55"}},
			},
			want: &ListStatData{
				PageLimit: 55,
				Page:      defaultPage,
				Amount:    10,
				Pages:     1,
			},
		},
		{
			name: "Pages number",
			args: args{
				amount: 133,
				vars: url.Values{
					string(urlParamPageLimit): []string{"60"},
					string(urlParamPage):      []string{"2"},
				},
			},
			want: &ListStatData{
				PageLimit: 60,
				Page:      2,
				Amount:    133,
				Pages:     3,
			},
		},
		{
			name: "Wrong page",
			args: args{
				amount: 10,
				vars:   url.Values{string(urlParamPage): []string{"3d"}},
			},
			wantErr: true,
		},
		{
			name: "Wrong page limit",
			args: args{
				amount: 10,
				vars:   url.Values{string(urlParamPageLimit): []string{"30.5"}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readListStatData(tt.args.vars, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("readListStatData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("readListStatData() -> %v", diff)
			}
		})
	}
}

type MockResponseWriter struct {
	*bytes.Buffer
	header     http.Header
	StatusCode int
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		Buffer:     bytes.NewBuffer(nil),
		header:     http.Header{},
		StatusCode: 200,
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

func Test_marshalAndWrite(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		v interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantBody   string
		wantStatus int
		wantPanic  bool
	}{
		{
			name: "OK",
			args: args{
				w: NewMockResponseWriter(),
				v: []string{"hello", "my", "test"},
			},
			wantBody:   `["hello","my","test"]`,
			wantStatus: 200,
		},
		{
			name: "Object can't be Marshaled",
			args: args{
				w: NewMockResponseWriter(),
				v: 10 + 12i,
			},
			wantPanic:  true,
			wantStatus: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil && !tt.wantPanic {
					t.Errorf("Caught unexpected panic")
				}
				buf := tt.args.w.(*MockResponseWriter)
				if out := buf.String(); tt.wantBody != "" && out != tt.wantBody {
					t.Errorf("Got unexpected body. Got %s, want %s", out, tt.wantBody)
				}
				if tt.wantStatus != buf.StatusCode {
					t.Errorf("Got unexpected status code. Got %d, want %d", buf.StatusCode, tt.wantStatus)
				}
			}()
			marshalAndWrite(tt.args.w, tt.args.v)
		})
	}
}

type MockReadCloser struct {
	text    []byte
	nextErr bool
}

func NewMockReadCloser(text string, makeErr bool) *MockReadCloser {
	return &MockReadCloser{
		text:    []byte(text),
		nextErr: makeErr,
	}
}

func (r *MockReadCloser) Read(p []byte) (n int, err error) {
	if r.nextErr {
		return 0, errors.New("bad read")
	}
	for ; n < len(p) && n < len(r.text); n++ {
		p[n] = r.text[n]
	}
	r.text = r.text[n:]
	if len(r.text) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (r *MockReadCloser) Close() error {
	return nil
}

func Test_unmarshalRequestBody(t *testing.T) {
	type args struct {
		body io.ReadCloser
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantV   interface{}
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				body: NewMockReadCloser(`["hello","my","test"]`, false),
				v:    &[]string{},
			},
			wantV: &[]string{"hello", "my", "test"},
		},
		{
			name: "Empty body",
			args: args{
				body: NewMockReadCloser("", false),
				v:    &[]string{},
			},
			wantErr: true,
		},
		{
			name: "Type mismatch",
			args: args{
				body: NewMockReadCloser("10+32i", false),
				v:    &[]string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unmarshalReaderContent(tt.args.body, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalReaderContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(tt.wantV, tt.args.v); diff != nil {
				t.Errorf("unmarshalReaderContent() -> %v", diff)
			}
		})
	}
}

func Test_readLengthMeasureAndPrecision(t *testing.T) {
	const (
		defaultMeasure   = value.Measure(math.MaxFloat64)
		defaultPrecision = math.MaxInt64
	)
	var (
		measure   value.Measure
		precision int
	)

	type args struct {
		vars      url.Values
		measure   *value.Measure
		precision *int
	}
	tests := []struct {
		name          string
		args          args
		wantMeasure   value.Measure
		wantPrecision int
		wantErr       bool
	}{
		{
			name: "Empty values",
			args: args{
				vars:      url.Values{},
				measure:   &measure,
				precision: &precision,
			},
			wantPrecision: defaultPrecision,
			wantMeasure:   defaultMeasure,
		},
		{
			name: "Only precision",
			args: args{
				vars: url.Values{
					string(urlParamPrecision): []string{"9"},
				},
				measure:   &measure,
				precision: &precision,
			},
			wantPrecision: 9,
			wantMeasure:   defaultMeasure,
		},
		{
			name: "Only measure",
			args: args{
				vars: url.Values{
					string(urlParamMeasure): []string{"ft"},
				},
				measure:   &measure,
				precision: &precision,
			},
			wantPrecision: defaultPrecision,
			wantMeasure:   value.Foot,
		},
		{
			name: "Both parameters",
			args: args{
				vars: url.Values{
					string(urlParamMeasure):   []string{"ft"},
					string(urlParamPrecision): []string{"9"},
				},
				measure:   &measure,
				precision: &precision,
			},
			wantPrecision: 9,
			wantMeasure:   value.Foot,
		},
		{
			name: "Precision less than zero",
			args: args{
				vars: url.Values{
					string(urlParamMeasure):   []string{"ft"},
					string(urlParamPrecision): []string{"-9"},
				},
				measure:   &measure,
				precision: &precision,
			},
			wantErr: true,
		},
		{
			name: "Precision syntax error",
			args: args{
				vars: url.Values{
					string(urlParamMeasure):   []string{"ft"},
					string(urlParamPrecision): []string{"-da9"},
				},
				measure:   &measure,
				precision: &precision,
			},
			wantErr: true,
		},
		{
			name: "Unknown measure",
			args: args{
				vars: url.Values{
					string(urlParamMeasure): []string{"ftd"},
				},
				measure:   &measure,
				precision: &precision,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measure, precision = defaultMeasure, defaultPrecision
			if err := readLengthMeasureAndPrecision(tt.args.vars, tt.args.measure, tt.args.precision); (err != nil) != tt.wantErr {
				t.Errorf("readLengthMeasureAndPrecision() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if measure != tt.wantMeasure {
					t.Errorf("readLengthMeasureAndPrecision() measure got = %v, want %v", measure, tt.wantMeasure)
				}
				if precision != tt.wantPrecision {
					t.Errorf("readLengthMeasureAndPrecision() precision got = %v, want %v", precision, tt.wantPrecision)
				}
			}
		})
	}
}

func Test_parseURLParamValue(t *testing.T) {
	var (
		s string
		u uint
	)
	type args struct {
		vars url.Values
		key  urlParamKey
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantS   interface{}
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				vars: url.Values{"ping": []string{"pong"}},
				key:  "ping",
				v:    &s,
			},
			wantS: "pong",
		},
		{
			name: "Not found",
			args: args{
				vars: url.Values{},
				key:  "ping",
				v:    &s,
			},
			wantErr: true,
		},
		{
			name: "Syntax error",
			args: args{
				vars: url.Values{"ping": []string{"pong"}},
				key:  "ping",
				v:    &u,
			},
			wantErr: true,
		},
		{
			name: "Another error",
			args: args{
				vars: url.Values{},
				key:  "ping",
				v:    s,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, s = 0, ""
			if err := parseURLParamValue(tt.args.vars, tt.args.key, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("parseURLParamValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.wantS != "" && tt.wantS != s {
				t.Errorf("parseURLParamValue() got = %v, want %v", s, tt.wantS)
			}
		})
	}
}

func Test_parsePathValue(t *testing.T) {
	var (
		u uint
	)

	type args struct {
		vars map[string]string
		key  pathVarKey
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantU   uint
		wantErr bool
	}{
		{
			name:  "OK",
			args:  args{vars: map[string]string{"one": "1"}, key: "one", v: &u},
			wantU: 1,
		},
		{
			name:    "Not found",
			args:    args{vars: map[string]string{"one": "1"}, key: "two", v: &u},
			wantErr: true,
		},
		{
			name:    "Syntax error",
			args:    args{vars: map[string]string{"one": "ds"}, key: "one", v: &u},
			wantErr: true,
		},
		{
			name:    "Another error",
			args:    args{vars: map[string]string{"one": "ds"}, key: "one", v: u},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		u = 0
		t.Run(tt.name, func(t *testing.T) {
			if err := parsePathValue(tt.args.vars, tt.args.key, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("parsePathValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.wantU != 0 && tt.wantU != u {
				t.Errorf("parsePathValue() got = %v, want %v", u, tt.wantU)
			}
		})
	}
}

func Test_readCtxValue(t *testing.T) {
	var (
		i       int
		example = math.MaxInt64
		ctx     context.Context
	)

	type args struct {
		ctxValues [][2]interface{}
		key       ctxKey
		v         interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantI   int
		wantErr bool
	}{
		{
			name:  "OK",
			args:  args{ctxValues: [][2]interface{}{{"no", "no"}, {ctxKey(123), 3}}, key: ctxKey(123), v: &i},
			wantI: 3,
		},
		{
			name:    "Not settable",
			args:    args{ctxValues: [][2]interface{}{{"no", "no"}, {ctxKey(123), 3}}, key: ctxKey(123), v: i},
			wantErr: true,
		},
		{
			name:    "Not found",
			args:    args{ctxValues: [][2]interface{}{{"no", "no"}, {ctxKey(123), 3}}, key: ctxKey(1234), v: &i},
			wantErr: true,
		},
		{
			name:  "Get value from pointer",
			args:  args{ctxValues: [][2]interface{}{{"no", "no"}, {ctxKey(123), &example}}, key: ctxKey(123), v: &i},
			wantI: example,
		},
		{
			name:    "Different types",
			args:    args{ctxValues: [][2]interface{}{{"no", "no"}, {ctxKey(123), uint(4)}}, key: ctxKey(123), v: &i},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, i = context.Background(), 0
			for _, vs := range tt.args.ctxValues {
				ctx = context.WithValue(ctx, vs[0], vs[1])
			}
			if err := readCtxValue(ctx, tt.args.key, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("readCtxValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.wantI != 0 && tt.wantI != i {
				t.Errorf("readCtxValue() got = %v, want %v", i, tt.wantI)
			}
		})
	}
}

func Test_getSettable(t *testing.T) {
	var i int
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    reflect.Value
		wantErr bool
	}{
		{
			name: "OK",
			args: args{v: &i},
			want: reflect.ValueOf(&i).Elem(),
		},
		{
			name:    "Not settable",
			args:    args{v: i},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i = math.MaxInt64
			got, err := getSettable(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSettable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := deep.Equal(got, &tt.want); diff != nil {
				t.Errorf("getSettable() -> %v", diff)
			}
		})
	}
}
