package drawing

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrWrongArgumentType = errors.New("wrong argument type")

type Description [][2]string

func NewDescription() *Description {
	return new(Description)
}

func NewUnionDescription(descriptions ...*Description) *Description {
	out := make(Description, 0)
	for _, d := range descriptions {
		out = append(out, *d...)
	}
	return &out
}

func (d *Description) PushBack(key, value string) {
	*d = append(*d, [2]string{key, value})
}

func (d Description) ToStringSlice() []string {
	out := make([]string, 0)
	for _, v := range d {
		out = append(out, fmt.Sprintf("%s: %s", v[0], v[1]))
	}
	return out
}

func (d *Description) Scan(value interface{}) error {
	var data []byte
	switch value.(type) {
	case []byte:
		data = value.([]byte)
	case string:
		data = []byte(value.(string))
	default:
		return ErrWrongArgumentType
	}
	return json.Unmarshal(data, d)
}

func (d Description) Value() (driver.Value, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}
