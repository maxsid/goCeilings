package drawing

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrWrongArgumentType = errors.New("wrong argument type")

type Description map[string]string

func (d Description) Add(key, value string) {
	d[key] = value
}

func (d Description) AddDescription(ad Description) {
	for k, v := range ad {
		d[k] = v
	}
}

func (d Description) ToStringSlice() []string {
	out := make([]string, 0)
	for k, v := range d {
		out = append(out, fmt.Sprintf("%s: %s", k, v))
	}
	return out
}

func (d Description) Scan(value interface{}) error {
	var data []byte
	switch value.(type) {
	case []byte:
		data = value.([]byte)
	case string:
		data = []byte(value.(string))
	default:
		return ErrWrongArgumentType
	}
	return json.Unmarshal(data, &d)
}

func (d Description) Value() (driver.Value, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}
