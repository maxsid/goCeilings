package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maxsid/goCeilings/value"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	defaultPage           = 1
	defaultPageLimit      = 30
	pageVariableName      = "p"
	pageLimitVariableName = "lim"
)

func readListStatData(vars url.Values, amount uint) (*ListStatData, error) {
	listStat := ListStatData{Amount: amount, Page: defaultPage, PageLimit: defaultPageLimit}
	if page, err := readUintURLVar(vars, pageVariableName); err == nil {
		listStat.Page = page
	} else if !errors.Is(err, ErrCouldNotExistURLVariable) {
		return nil, err
	}
	if limit, err := readUintURLVar(vars, pageLimitVariableName); err == nil {
		listStat.PageLimit = limit
	} else if !errors.Is(err, ErrCouldNotExistURLVariable) {
		return nil, err
	}
	if ps := amount / listStat.PageLimit; amount%listStat.PageLimit != 0 {
		listStat.Pages = ps + 1
	} else {
		listStat.Pages = ps
	}
	if listStat.Page > listStat.Pages {
		listStat.Page = listStat.Pages
	}
	return &listStat, nil
}

func readUintURLVar(vars url.Values, varName string) (uint, error) {
	if v := vars.Get(varName); v != "" {
		if v, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(v), nil
		} else {
			return 0, err
		}
	}
	return 0, ErrCouldNotExistURLVariable
}

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

func marshalAndWrite(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, "%s", data)
}

func unmarshalRequestBody(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	defer req.Body.Close()
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(err)
		return false
	}
	if len(requestBody) == 0 {
		http.Error(w, "Bad Request: Got empty body", http.StatusBadRequest)
		return false
	}
	if err := json.Unmarshal(requestBody, v); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Print(err)
		return false
	}
	return true
}

func readMeasureAndPrecisionFromURL(w http.ResponseWriter, req *http.Request, measure value.Measure, precision int) (value.Measure, int, bool) {
	if p := req.URL.Query().Get("p"); p != "" {
		var err error
		if precision, err = strconv.Atoi(p); err != nil || precision < 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return 0, 0, false
		}
	}
	if m := req.URL.Query().Get("m"); m != "" {
		if measure = value.LengthMeasureByName(m); measure == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return 0, 0, false
		}
	}
	return measure, precision, true
}
