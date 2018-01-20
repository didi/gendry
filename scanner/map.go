package scanner

import (
	"errors"
	"reflect"
	"strings"
)

var (
	// ErrNoneStructTarget as its name says
	ErrNoneStructTarget = errors.New("[scanner] target must be a struct type")
)

// Map converts a struct to a map
// type for each field of the struct must be built-in type
func Map(target interface{}, useTag string) (map[string]interface{}, error) {
	if nil == target {
		return nil, nil
	}
	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrNoneStructTarget
	}
	t := v.Type()
	result := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		keyName := getKey(t.Field(i), useTag)
		if "" == keyName {
			continue
		}
		result[keyName] = v.Field(i).Interface()
	}
	return result, nil
}

func isExportedField(name string) bool {
	return strings.Title(name) == name
}

func getKey(field reflect.StructField, useTag string) string {
	if !isExportedField(field.Name) {
		return ""
	}
	if field.Type.Kind() == reflect.Ptr {
		return ""
	}
	if "" == useTag {
		return field.Name
	}
	tag, ok := field.Tag.Lookup(useTag)
	if !ok {
		return ""
	}
	return resolveTagName(tag)
}

func resolveTagName(tag string) string {
	idx := strings.IndexByte(tag, ',')
	if -1 == idx {
		return tag
	}
	return tag[:idx]
}
