package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maxsid/goCeilings/value"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

const (
	defaultPage      = 1
	defaultPageLimit = 30
)

type ctxKey int
type pathVarKey string
type urlParamKey string

// readListStatData reads a number of the page and the page limit,
// calculates amount of the pages from the amount of elements, specified in parameter.
func readListStatData(vars url.Values, amount uint) (*ListStatData, error) {
	listStat := ListStatData{Amount: amount, Page: defaultPage, PageLimit: defaultPageLimit}
	if err := parseURLParamValue(vars, urlParamPage, &listStat.Page); err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err := parseURLParamValue(vars, urlParamPageLimit, &listStat.PageLimit); err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	listStat.Pages = uint(math.Ceil(float64(amount) / float64(listStat.PageLimit)))
	if listStat.Page > listStat.Pages {
		listStat.Page = listStat.Pages
	}
	return &listStat, nil
}

// addRequestPointsToDrawing calculates and adds points to the Drawing from PointCalculating requests.
func addRequestPointsToDrawing(drawing *Drawing, points ...*PointCalculating) error {
	for _, p := range points {
		if p.Direction != nil && p.Distance != 0 {
			if err := drawing.AddPointByDirection(p.Distance, *p.Direction); err != nil {
				return err
			}
		} else if p.Angle != nil && p.Distance != 0 {
			if err := drawing.AddPointByAngle(p.Distance, *p.Angle); err != nil {
				return err
			}
		} else {
			drawing.AddPointByMeasure(p.X, p.Y)
		}
	}
	return nil
}

// marshalAndWrite does marshal of v variable and writes result into http.ResponseWriter.
func marshalAndWrite(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if writeError(w, err) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, "%s", data)
}

// unmarshalReaderContent does unmarshal of v argument into v variable.
func unmarshalReaderContent(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if len(bodyBytes) == 0 {
		return ErrEmptyRequestBody
	}
	if err := json.Unmarshal(bodyBytes, v); err != nil {
		return err
	}
	return nil
}

// readLengthMeasureAndPrecision parses and writes measure and precision from specified GET URL parameters.
func readLengthMeasureAndPrecision(vars url.Values, measure *value.Measure, precision *int) error {
	wasPrecision := *precision
	if err := parseURLParamValue(vars, urlParamPrecision, precision); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if *precision < 0 {
		*precision = wasPrecision
		return fmt.Errorf("%w of precision (%s) - the parameter less then zero", ErrCouldNotReadURLParameter, urlParamPrecision)
	}
	measureName := ""
	if err := parseURLParamValue(vars, urlParamMeasure, &measureName); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if measureName != "" {
		if readMeasure := value.LengthMeasureByName(measureName); readMeasure == 0 {
			return fmt.Errorf("%w of measure (%s)", ErrCouldNotReadURLParameter, urlParamMeasure)
		} else {
			*measure = readMeasure
		}
	}
	return nil
}

// parseURLParamValue checks existing of URL parameter, then parses and write it to v parameter.
// Parsing performs into v type.
func parseURLParamValue(vars url.Values, key urlParamKey, v interface{}) error {
	paramValue := vars.Get(string(key))
	if paramValue == "" {
		return fmt.Errorf("%w %s url parameter", ErrNotFound, key)
	}
	if err := parseString(v, paramValue); err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			return fmt.Errorf("%w - syntax error", ErrCouldNotReadURLParameter)
		}
		return err
	}
	return nil
}

// parsePathValue checks existing of mux path variable, then parses and write it to v parameter.
// Parsing performs into v type.
func parsePathValue(vars map[string]string, key pathVarKey, v interface{}) error {
	pathValue, ok := vars[string(key)]
	if !ok {
		return fmt.Errorf("%w %s path value", ErrNotFound, key)
	}
	if err := parseString(v, pathValue); err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			return fmt.Errorf("%w - syntax error", ErrCouldNotReadPathVar)
		}
		return err
	}
	return nil
}

// readCtxValue gets value with key from context ctx and write it in v argument, which must be a settable.
func readCtxValue(ctx context.Context, key ctxKey, v interface{}) error {
	valueOfV, err := getSettable(v)
	if err != nil {
		return err
	}
	ctxValue := ctx.Value(key)
	if ctxValue == nil {
		return fmt.Errorf("%w: got a nil value", ErrCouldNotReadCtxValue)
	}
	valueOfCtxV := reflect.ValueOf(ctxValue)
	if valueOfCtxV.Kind() == reflect.Ptr {
		valueOfCtxV = valueOfCtxV.Elem()
	}
	if got, want := valueOfCtxV.Type(), valueOfV.Type(); got != want {
		return fmt.Errorf("%w: got different values types. Got %s, want %s", ErrCouldNotReadCtxValue, got, want)
	}
	valueOfV.Set(valueOfCtxV)
	return nil
}

// getSettable returns reflect.Value object of a settable parameter.
func getSettable(v interface{}) (*reflect.Value, error) {
	valueOfV := reflect.ValueOf(v)
	if !valueOfV.CanSet() {
		if valueOfV.Kind() == reflect.Ptr {
			valueOfV = valueOfV.Elem()
		} else {
			return nil, ErrValueIsNotSettable
		}
	}
	return &valueOfV, nil
}

// parseString parses string value and write it to v parameter.
// Parsing performs into v type.
func parseString(v interface{}, parsingString string) (err error) {
	valueOfV, err := getSettable(v)
	if err != nil {
		return err
	}
	var setValue interface{}
	switch valueOfV.Kind() {
	case reflect.String:
		setValue = parsingString
	case reflect.Int:
		setValue, err = strconv.Atoi(parsingString)
	case reflect.Int64:
		setValue, err = strconv.ParseInt(parsingString, 10, 64)
	case reflect.Int32:
		setValue, err = strconv.ParseInt(parsingString, 10, 32)
	case reflect.Int16:
		setValue, err = strconv.ParseInt(parsingString, 10, 16)
	case reflect.Int8:
		setValue, err = strconv.ParseInt(parsingString, 10, 8)
	case reflect.Uint:
		sv64 := uint64(0)
		sv64, err = strconv.ParseUint(parsingString, 10, 64)
		setValue = uint(sv64)
	case reflect.Uint64:
		setValue, err = strconv.ParseUint(parsingString, 10, 64)
	case reflect.Uint32:
		setValue, err = strconv.ParseUint(parsingString, 10, 32)
	case reflect.Uint16:
		setValue, err = strconv.ParseUint(parsingString, 10, 16)
	case reflect.Uint8:
		setValue, err = strconv.ParseUint(parsingString, 10, 8)
	case reflect.Bool:
		setValue, err = strconv.ParseBool(parsingString)
	case reflect.Float64:
		setValue, err = strconv.ParseFloat(parsingString, 64)
	case reflect.Float32:
		setValue, err = strconv.ParseFloat(parsingString, 32)
	case reflect.Complex64:
		setValue, err = strconv.ParseComplex(parsingString, 64)
	case reflect.Complex128:
		setValue, err = strconv.ParseComplex(parsingString, 128)
	default:
		return ErrWrongValueKind
	}
	if err != nil {
		return
	}
	valueOfV.Set(reflect.ValueOf(setValue))
	return
}
