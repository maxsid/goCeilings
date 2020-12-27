package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/maxsid/goCeilings/figure"
	"github.com/maxsid/goCeilings/server/common"
	"github.com/maxsid/goCeilings/value"
)

const (
	defaultPage      = 1
	defaultPageLimit = 30
)

type ctxKey int
type pathVarKey string
type urlParamKey string

// getCtxValue gets value from context. Function returns error if value doesn't exist or it's nil.
func getCtxValue(ctx context.Context, key ctxKey) (interface{}, error) {
	ctxValue := ctx.Value(key)
	if ctxValue == nil {
		return nil, fmt.Errorf("%w: got a nil value", ErrCouldNotReadCtxValue)
	}
	return ctxValue, nil
}

// getDrawingByRequestOrWriteError gets Drawing from database by its ID in request path.
// Second value of the returning tuple contains successfulness of the operation.
func getDrawingByRequestOrWriteError(w http.ResponseWriter, req *http.Request) (*common.Drawing, bool) {
	storage, drawingID := (common.UserStorage)(nil), uint(0)
	if storage = getUserStorageOrWriteError(w, req); storage == nil {
		return nil, false
	}
	if err := parsePathValue(mux.Vars(req), pathVarDrawingID, &drawingID); writeError(w, err) {
		return nil, false
	}

	drawing, err := storage.GetDrawing(drawingID)
	if writeError(w, err) {
		return nil, false
	}
	return drawing, true
}

// getDrawingByRequestOrWriteError reads index of the point from request path.
// Second value of the returning tuple contains successfulness of the operation.
func getPointIndexByRequestOrWriteError(w http.ResponseWriter, req *http.Request, drawing *common.Drawing) (int, bool) {
	pointIndex := 0
	if err := parsePathValue(mux.Vars(req), pathVarPointNumber, &pointIndex); writeError(w, err) {
		return 0, false
	}
	if pointIndex > drawing.Len() || pointIndex < 1 {
		_ = writeError(w, ErrPointNotFound)
		return 0, false
	}
	return pointIndex - 1, true
}

// getPointsFromRequestPoint converts []*pointCalculating requests into []*figure.Point.
func getPointsFromRequestPoint(points ...*pointCalculating) []*figure.Point {
	resultPoints := make([]*figure.Point, len(points))
	for i, p := range points {
		switch {
		case p.Direction != nil && p.Distance != 0:
			resultPoints[i] = figure.NewCalculatedPoint(&figure.DirectionCalculator{
				Direction: *p.Direction,
				Distance:  p.Distance,
			})
		case p.Angle != nil && p.Distance != 0:
			resultPoints[i] = figure.NewCalculatedPoint(&figure.AngleCalculator{
				Angle:    *p.Angle,
				Distance: p.Distance,
			})
		default:
			resultPoints[i] = figure.NewPoint(p.X, p.Y)
		}
	}
	return resultPoints
}

// getSettable returns reflect.Value object of a settable parameter.
func getSettable(v interface{}) (*reflect.Value, error) {
	valueOfV := reflect.Indirect(reflect.ValueOf(v))
	if valueOfV == (reflect.Value{}) {
		return nil, fmt.Errorf("%w: got nil pointer", ErrWrongValueKind)
	}
	if !valueOfV.CanSet() {
		return nil, ErrValueIsNotSettable
	}
	return &valueOfV, nil
}

// getUserStorageOrWriteError gets UserStorage from request context.
// With errors writes into http.ResponseWriter via writeError function and returns nil.
func getUserStorageOrWriteError(w http.ResponseWriter, req *http.Request) common.UserStorage {
	st, err := getCtxValue(req.Context(), ctxKeyUserStorage)
	if writeError(w, err) {
		return nil
	}
	storage, ok := st.(common.UserStorage)
	if !ok {
		writeError(w, fmt.Errorf("%w: got a wrong value type", ErrCouldNotReadCtxValue))
		return nil
	}
	return storage
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
		readMeasure := value.LengthMeasureByName(measureName)
		if readMeasure == 0 {
			return fmt.Errorf("%w of measure (%s)", ErrCouldNotReadURLParameter, urlParamMeasure)
		}
		*measure = readMeasure
	}
	return nil
}

// readListStatData reads a number of the page and the page limit,
// calculates amount of the pages from the amount of elements, specified in parameter.
func readListStatData(vars url.Values, amount uint) (*listStatData, error) {
	listStat := listStatData{Amount: amount, Page: defaultPage, PageLimit: defaultPageLimit}
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
