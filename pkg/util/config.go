/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
)

// findNestedElement uses reflection to find the element corresponding to the dot-separated string parameter.
func findNestedElement(s string, c interface{}) (reflect.Value, error) {
	fields := strings.Split(s, ".")

	// Take the ValueOf to get a pointer, so we can actually mutate the element.
	e := reflect.Indirect(reflect.ValueOf(c).Elem())

	for _, field := range fields {
		e = reflect.Indirect(e.FieldByName(field))

		// FieldByName returns the zero value if the field does not exist.
		if e == (reflect.Value{}) {
			return e, fmt.Errorf("unable to find field by name: %s", field)
		}
		// Start the loop again, on the next level.
	}
	return e, nil
}

// setElement sets the supplied element to the value in the supplied string. The string will be coerced to the correct type.
func setElement(e reflect.Value, v string) error {
	switch e.Interface().(type) {
	case int, int32, int64:
		return convertInt(e, v)
	case string:
		return convertString(e, v)
	case float32, float64:
		return convertFloat(e, v)
	case bool:
		return convertBool(e, v)
	case net.IP:
		return convertIP(e, v)
	case net.IPNet:
		return convertCIDR(e, v)
	case utilnet.PortRange:
		return convertPortRange(e, v)
	case time.Duration:
		return convertDuration(e, v)
	case []string:
		vals := strings.Split(v, ",")
		e.Set(reflect.ValueOf(vals))
	case map[string]string:
		return convertMap(e, v)
	default:
		// Last ditch attempt to convert anything based on its underlying kind.
		// This covers any types that are aliased to a native type
		return convertKind(e, v)
	}

	return nil
}

func convertMap(e reflect.Value, v string) error {
	if e.IsNil() {
		e.Set(reflect.MakeMap(e.Type()))
	}
	vals := strings.Split(v, ",")
	for _, subitem := range vals {
		subvals := strings.FieldsFunc(subitem, func(c rune) bool {
			return c == '<' || c == '=' || c == '>'
		})
		if len(subvals) != 2 {
			return fmt.Errorf("unparsable %s", v)
		}
		e.SetMapIndex(reflect.ValueOf(subvals[0]), reflect.ValueOf(subvals[1]))
	}
	return nil
}

func convertKind(e reflect.Value, v string) error {
	switch e.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		return convertInt(e, v)
	case reflect.String:
		return convertString(e, v)
	case reflect.Float32, reflect.Float64:
		return convertFloat(e, v)
	case reflect.Bool:
		return convertBool(e, v)
	default:
		return fmt.Errorf("unable to set type %T", e.Kind())
	}
}

func convertInt(e reflect.Value, v string) error {
	i, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("Error converting input %s to an integer: %v", v, err)
	}
	e.SetInt(int64(i))
	return nil
}

func convertString(e reflect.Value, v string) error {
	e.SetString(v)
	return nil
}

func convertFloat(e reflect.Value, v string) error {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("Error converting input %s to a float: %v", v, err)
	}
	e.SetFloat(f)
	return nil
}

func convertBool(e reflect.Value, v string) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fmt.Errorf("Error converting input %s to a bool: %v", v, err)
	}
	e.SetBool(b)
	return nil
}

func convertIP(e reflect.Value, v string) error {
	ip := net.ParseIP(v)
	if ip == nil {
		return fmt.Errorf("Error converting input %s to an IP", v)
	}
	e.Set(reflect.ValueOf(ip))
	return nil
}

func convertCIDR(e reflect.Value, v string) error {
	_, cidr, err := net.ParseCIDR(v)
	if err != nil {
		return fmt.Errorf("Error converting input %s to a CIDR: %v", v, err)
	}
	e.Set(reflect.ValueOf(*cidr))
	return nil
}

func convertPortRange(e reflect.Value, v string) error {
	pr, err := utilnet.ParsePortRange(v)
	if err != nil {
		return fmt.Errorf("Error converting input %s to PortRange: %v", v, err)
	}
	e.Set(reflect.ValueOf(*pr))
	return nil
}

func convertDuration(e reflect.Value, v string) error {
	dur, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("Error converting input %s to Duration: %v", v, err)
	}
	e.Set(reflect.ValueOf(dur))
	return nil
}

// FindAndSet sets the nested value.
func FindAndSet(path string, c interface{}, value string) error {
	elem, err := findNestedElement(path, c)
	if err != nil {
		return err
	}
	return setElement(elem, value)
}
