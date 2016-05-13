package bqschema

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"google.golang.org/api/bigquery/v2"
)

var (
	ArrayOfArray = errors.New("Array of Arrays not allowed")
	UnknownType  = errors.New("Unknown type")
	NotStruct    = errors.New("Can not convert non structs")
)

func ToSchema(src interface{}) (*bigquery.TableSchema, error) {
	value := reflect.ValueOf(src)
	t := value.Type()

	schema := &bigquery.TableSchema{}

	if t.Kind() == reflect.Struct {
		schema.Fields = make([]*bigquery.TableFieldSchema, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			if sf.PkgPath != "" { // unexported
				continue
			}
			v := pointerGuard(value.Field(i))

			var name string
			jsonTag := sf.Tag.Get("json")
			switch jsonTag {
			case "-":
				continue
			case "":
				name = sf.Name
			default:
				name = strings.Split(jsonTag, ",")[0]
			}

			tfs := &bigquery.TableFieldSchema{
				Mode: "required",
				Name: name,
				Type: "",
			}
			schema.Fields = append(schema.Fields, tfs)

			kind := v.Kind()
			t, isSimple := simpleType(kind)

			if isSimple {
				tfs.Type = t
			} else {
				switch kind {
				case reflect.Struct:
					tfs.Mode = "nullable"
					if t, fields, err := structConversion(v.Interface()); err == nil {
						tfs.Type = t
						if t == "string" {
							tfs.Mode = "required"
						}
						tfs.Fields = fields
					} else {
						return schema, err
					}
				case reflect.Array, reflect.Slice:
					tfs.Mode = "repeated"
					subKind := pointerGuard(v.Type().Elem()).Kind()
					t, isSimple := simpleType(subKind)
					if isSimple {
						schema.Fields[i].Type = t
					} else if subKind == reflect.Struct {
						subStruct := reflect.Zero(pointerGuard(v.Type().Elem()).Type()).Interface()
						if t, fields, err := structConversion(subStruct); err == nil {
							schema.Fields[i].Type = t
							schema.Fields[i].Fields = fields

						} else {
							return schema, err
						}
					} else {
						return schema, ArrayOfArray
					}
				default:
					return schema, UnknownType
				}
			}
		}
	} else {
		return schema, NotStruct
	}

	return schema, nil
}

func MustToSchema(src interface{}) *bigquery.TableSchema {
	schema, err := ToSchema(src)
	if err != nil {
		panic(err)
	}
	return schema
}

func simpleType(kind reflect.Kind) (string, bool) {
	isSimple := true
	t := ""
	switch kind {
	case reflect.Bool:
		t = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t = "integer"
	case reflect.Float32, reflect.Float64:
		t = "float"
	case reflect.String:
		t = "string"
	default:
		isSimple = false
	}
	return t, isSimple
}

func structConversion(src interface{}) (string, []*bigquery.TableFieldSchema, error) {
	v := reflect.ValueOf(src)
	if v.Type().Name() == "Key" && strings.Contains(v.Type().PkgPath(), "appengine") {
		return "string", nil, nil
	} else if v.Type().ConvertibleTo(reflect.TypeOf(time.Time{})) {
		return "timestamp", nil, nil
	} else {
		schema, err := ToSchema(src)
		return "record", schema.Fields, err
	}
}

func pointerGuard(i interface{}) reflect.Value {
	var v reflect.Value
	var ok bool
	v, ok = i.(reflect.Value)
	if !ok {
		if t, ok := i.(reflect.Type); ok {
			v = reflect.Indirect(reflect.New(t))
		}
	}

	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(reflect.New(v.Type().Elem()))
	}
	return v
}
