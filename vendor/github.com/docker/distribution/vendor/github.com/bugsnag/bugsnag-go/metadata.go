package bugsnag

import (
	"fmt"
	"reflect"
	"strings"
)

// MetaData is added to the Bugsnag dashboard in tabs. Each tab is
// a map of strings -> values. You can pass MetaData to Notify, Recover
// and AutoNotify as rawData.
type MetaData map[string]map[string]interface{}

// Update the meta-data with more information. Tabs are merged together such
// that unique keys from both sides are preserved, and duplicate keys end up
// with the provided values.
func (meta MetaData) Update(other MetaData) {
	for name, tab := range other {

		if meta[name] == nil {
			meta[name] = make(map[string]interface{})
		}

		for key, value := range tab {
			meta[name][key] = value
		}
	}
}

// Add creates a tab of Bugsnag meta-data.
// If the tab doesn't yet exist it will be created.
// If the key already exists, it will be overwritten.
func (meta MetaData) Add(tab string, key string, value interface{}) {
	if meta[tab] == nil {
		meta[tab] = make(map[string]interface{})
	}

	meta[tab][key] = value
}

// AddStruct creates a tab of Bugsnag meta-data.
// The struct will be converted to an Object using the
// reflect library so any private fields will not be exported.
// As a safety measure, if you pass a non-struct the value will be
// sent to Bugsnag under the "Extra data" tab.
func (meta MetaData) AddStruct(tab string, obj interface{}) {
	val := sanitizer{}.Sanitize(obj)
	content, ok := val.(map[string]interface{})
	if ok {
		meta[tab] = content
	} else {
		// Wasn't a struct
		meta.Add("Extra data", tab, obj)
	}

}

// Remove any values from meta-data that have keys matching the filters,
// and any that are recursive data-structures
func (meta MetaData) sanitize(filters []string) interface{} {
	return sanitizer{
		Filters: filters,
		Seen:    make([]interface{}, 0),
	}.Sanitize(meta)

}

// The sanitizer is used to remove filtered params and recursion from meta-data.
type sanitizer struct {
	Filters []string
	Seen    []interface{}
}

func (s sanitizer) Sanitize(data interface{}) interface{} {
	for _, s := range s.Seen {
		// TODO: we don't need deep equal here, just type-ignoring equality
		if reflect.DeepEqual(data, s) {
			return "[RECURSION]"
		}
	}

	// Sanitizers are passed by value, so we can modify s and it only affects
	// s.Seen for nested calls.
	s.Seen = append(s.Seen, data)

	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return data

	case reflect.String:
		return data

	case reflect.Interface, reflect.Ptr:
		return s.Sanitize(v.Elem().Interface())

	case reflect.Array, reflect.Slice:
		ret := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			ret[i] = s.Sanitize(v.Index(i).Interface())
		}
		return ret

	case reflect.Map:
		return s.sanitizeMap(v)

	case reflect.Struct:
		return s.sanitizeStruct(v, t)

		// Things JSON can't serialize:
		// case t.Chan, t.Func, reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
	default:
		return "[" + t.String() + "]"

	}

}

func (s sanitizer) sanitizeMap(v reflect.Value) interface{} {
	ret := make(map[string]interface{})

	for _, key := range v.MapKeys() {
		val := s.Sanitize(v.MapIndex(key).Interface())
		newKey := fmt.Sprintf("%v", key.Interface())

		if s.shouldRedact(newKey) {
			val = "[REDACTED]"
		}

		ret[newKey] = val
	}

	return ret
}

func (s sanitizer) sanitizeStruct(v reflect.Value, t reflect.Type) interface{} {
	ret := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {

		val := v.Field(i)
		// Don't export private fields
		if !val.CanInterface() {
			continue
		}

		name := t.Field(i).Name
		var opts tagOptions

		// Parse JSON tags. Supports name and "omitempty"
		if jsonTag := t.Field(i).Tag.Get("json"); len(jsonTag) != 0 {
			name, opts = parseTag(jsonTag)
		}

		if s.shouldRedact(name) {
			ret[name] = "[REDACTED]"
		} else {
			sanitized := s.Sanitize(val.Interface())
			if str, ok := sanitized.(string); ok {
				if !(opts.Contains("omitempty") && len(str) == 0) {
					ret[name] = str
				}
			} else {
				ret[name] = sanitized
			}

		}
	}

	return ret
}

func (s sanitizer) shouldRedact(key string) bool {
	for _, filter := range s.Filters {
		if strings.Contains(strings.ToLower(filter), strings.ToLower(key)) {
			return true
		}
	}
	return false
}
