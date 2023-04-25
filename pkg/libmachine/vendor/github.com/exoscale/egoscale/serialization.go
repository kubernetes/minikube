package egoscale

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func csQuotePlus(s string) string {
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "%5B", "[", -1)
	s = strings.Replace(s, "%5D", "]", -1)
	return s
}

func csEncode(s string) string {
	return csQuotePlus(url.QueryEscape(s))
}

func rawValue(b json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for _, v := range m {
		return v, nil
	}
	return nil, nil
}

func rawValues(b json.RawMessage) (json.RawMessage, error) {
	var i []json.RawMessage

	if err := json.Unmarshal(b, &i); err != nil {
		return nil, nil
	}

	return i[0], nil
}

// prepareValues uses a command to build a POST request
//
// command is not a Command so it's easier to Test
func prepareValues(prefix string, params *url.Values, command interface{}) error {
	value := reflect.ValueOf(command)
	typeof := reflect.TypeOf(command)

	// Going up the pointer chain to find the underlying struct
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
		value = value.Elem()
	}

	for i := 0; i < typeof.NumField(); i++ {
		field := typeof.Field(i)
		val := value.Field(i)
		tag := field.Tag
		if json, ok := tag.Lookup("json"); ok {
			n, required := extractJSONTag(field.Name, json)
			name := prefix + n

			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v := val.Int()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, strconv.FormatInt(v, 10))
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v := val.Uint()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, strconv.FormatUint(v, 10))
				}
			case reflect.Float32, reflect.Float64:
				v := val.Float()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, strconv.FormatFloat(v, 'f', -1, 64))
				}
			case reflect.String:
				v := val.String()
				if v == "" {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got \"\"", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					(*params).Set(name, v)
				}
			case reflect.Bool:
				v := val.Bool()
				if v == false {
					if required {
						params.Set(name, "false")
					}
				} else {
					(*params).Set(name, "true")
				}
			case reflect.Ptr:
				if val.IsNil() {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got tempty ptr", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					switch field.Type.Elem().Kind() {
					case reflect.Bool:
						params.Set(name, strconv.FormatBool(val.Elem().Bool()))
					default:
						log.Printf("[SKIP] %s.%s (%v) not supported", typeof.Name(), field.Name, field.Type.Elem().Kind())
					}
				}
			case reflect.Slice:
				switch field.Type.Elem().Kind() {
				case reflect.Uint8:
					switch field.Type {
					case reflect.TypeOf(net.IPv4zero):
						ip := (net.IP)(val.Bytes())
						if ip == nil || ip.Equal(net.IPv4zero) {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got zero IPv4 address", typeof.Name(), field.Name, val.Kind())
							}
						} else {
							(*params).Set(name, ip.String())
						}
					default:
						if val.Len() == 0 {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), field.Name, val.Kind())
							}
						} else {
							v := val.Bytes()
							(*params).Set(name, base64.StdEncoding.EncodeToString(v))
						}
					}
				case reflect.String:
					{
						if val.Len() == 0 {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), field.Name, val.Kind())
							}
						} else {
							elems := make([]string, 0, val.Len())
							for i := 0; i < val.Len(); i++ {
								// XXX what if the value contains a comma? Double encode?
								s := val.Index(i).String()
								elems = append(elems, s)
							}
							(*params).Set(name, strings.Join(elems, ","))
						}
					}
				default:
					if val.Len() == 0 {
						if required {
							return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), field.Name, val.Kind())
						}
					} else {
						err := prepareList(name, params, val.Interface())
						if err != nil {
							return err
						}
					}
				}
			case reflect.Map:
				if val.Len() == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got empty map", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					err := prepareMap(name, params, val.Interface())
					if err != nil {
						return err
					}
				}
			default:
				if required {
					return fmt.Errorf("Unsupported type %s.%s (%v)", typeof.Name(), field.Name, val.Kind())
				}
			}
		} else {
			log.Printf("[SKIP] %s.%s no json label found", typeof.Name(), field.Name)
		}
	}

	return nil
}

func prepareList(prefix string, params *url.Values, slice interface{}) error {
	value := reflect.ValueOf(slice)

	for i := 0; i < value.Len(); i++ {
		prepareValues(fmt.Sprintf("%s[%d].", prefix, i), params, value.Index(i).Interface())
	}

	return nil
}

func prepareMap(prefix string, params *url.Values, m interface{}) error {
	value := reflect.ValueOf(m)

	for i, key := range value.MapKeys() {
		var keyName string
		var keyValue string

		switch key.Kind() {
		case reflect.String:
			keyName = key.String()
		default:
			return fmt.Errorf("Only map[string]string are supported (XXX)")
		}

		val := value.MapIndex(key)
		switch val.Kind() {
		case reflect.String:
			keyValue = val.String()
		default:
			return fmt.Errorf("Only map[string]string are supported (XXX)")
		}
		params.Set(fmt.Sprintf("%s[%d].%s", prefix, i, keyName), keyValue)
	}
	return nil
}

// extractJSONTag returns the variable name or defaultName as well as if the field is required (!omitempty)
func extractJSONTag(defaultName, jsonTag string) (string, bool) {
	tags := strings.Split(jsonTag, ",")
	name := tags[0]
	required := true
	for _, tag := range tags {
		if tag == "omitempty" {
			required = false
		}
	}

	if name == "" || name == "omitempty" {
		name = defaultName
	}
	return name, required
}
