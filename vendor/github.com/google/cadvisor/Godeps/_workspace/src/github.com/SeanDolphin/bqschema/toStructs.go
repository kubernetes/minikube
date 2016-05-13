package bqschema

import (
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/api/bigquery/v2"
)

func ToStructs(result *bigquery.QueryResponse, dst interface{}) error {
	var err error
	value := reflect.Indirect(reflect.ValueOf(dst))

	itemType := value.Type().Elem()
	rowCount := len(result.Rows)

	nameMap := map[string]string{}

	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		nameMap[strings.ToLower(field.Name)] = field.Name
	}

	items := reflect.MakeSlice(value.Type(), rowCount, rowCount)
	for i := 0; i < rowCount; i++ {
		item := reflect.Indirect(reflect.New(itemType))
		row := result.Rows[i]
		for j, cell := range row.F {
			schemaField := result.Schema.Fields[j]

			if name, ok := nameMap[strings.ToLower(schemaField.Name)]; ok {
				field := item.FieldByName(name)
				if field.IsValid() {
					switch field.Kind() {
					case reflect.Float64, reflect.Float32:
						f, err := strconv.ParseFloat(cell.V.(string), 64)
						if err == nil {
							field.SetFloat(f)
						} else {
							return err
						}
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						i, err := strconv.ParseInt(cell.V.(string), 10, 64)
						if err == nil {
							field.SetInt(i)
						} else {
							return err
						}
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						i, err := strconv.ParseUint(cell.V.(string), 10, 64)
						if err == nil {
							field.SetUint(i)
						} else {
							return err
						}

					case reflect.Bool:
						b, err := strconv.ParseBool(cell.V.(string))
						if err == nil {
							field.SetBool(b)
						} else {
							return err
						}
					case reflect.String:
						field.Set(reflect.ValueOf(cell.V))
					}

				}
			}
		}
		items.Index(i).Set(item)
	}
	value.Set(items)
	return err
}
