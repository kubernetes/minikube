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
			return e, fmt.Errorf("Unable to find field by name: %s", field)
		}
		// Start the loop again, on the next level.
	}
	return e, nil
}

// setElement sets the supplied element to the value in the supplied string. The string will be coerced to the correct type.
func setElement(e reflect.Value, v string) error {
	switch e.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("Error converting input %s to an integer: %s", v, err)
		}
		e.SetInt(int64(i))
	case reflect.String:
		e.SetString(v)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("Error converting input %s to a float: %s", v, err)
		}
		e.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("Error converting input %s to a bool: %s", v, err)
		}
		e.SetBool(b)
	default:
		switch t := e.Interface().(type) {
		case net.IP:
			ip := net.ParseIP(v)
			if ip == nil {
				return fmt.Errorf("Error converting input %s to an IP.", v)
			}
			e.Set(reflect.ValueOf(ip))
		case net.IPNet:
			_, cidr, err := net.ParseCIDR(v)
			if err != nil {
				return fmt.Errorf("Error converting input %s to a CIDR: %s", v, err)
			}
			e.Set(reflect.ValueOf(*cidr))
		case utilnet.PortRange:
			pr, err := utilnet.ParsePortRange(v)
			if err != nil {
				return fmt.Errorf("Error converting input %s to PortRange: %s", v, err)
			}
			e.Set(reflect.ValueOf(*pr))
		case []string:
			vals := strings.Split(v, ",")
			e.Set(reflect.ValueOf(vals))
		default:
			return fmt.Errorf("Unable to set type %T.", t)
		}
	}
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
