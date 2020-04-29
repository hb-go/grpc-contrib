package log

import (
	"encoding/json"
)

type String interface {
	String() string
}

type jsonString struct {
	v interface{}
}

func (l *jsonString) String() string {
	b, _ := json.Marshal(l.v)
	return string(b)
}
func StringJSON(v interface{}) String {
	return &jsonString{v: v}
}
