package logutil

import "fmt"

type Fields map[string]interface{}

func (f Fields) String() string {
	var s string
	for k, v := range f {
		if sv, ok := v.(string); ok {
			v = fmt.Sprintf("%q", sv)
		}
		s += fmt.Sprintf(" %s=%v", k, v)
	}
	return s
}
