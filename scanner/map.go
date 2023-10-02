package scanner

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

var (
	// ErrNoneStructTarget as its name says
	ErrNoneStructTarget = errors.New("[scanner] target must be a struct type")

	enableMapNameCache bool

	nameCache = &keyNameCache{
		nameMap: make(map[reflect.Type]map[string][]string),
	}
)

// EnableMapNameCache improve performance by caching resolved key names
func EnableMapNameCache(isEnable bool) {
	enableMapNameCache = isEnable
}

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

	var getKeyName func(int) string
	if enableMapNameCache {
		names := nameCache.GetKeyNames(t, useTag)
		getKeyName = func(i int) string {
			return names[i]
		}
	} else {
		getKeyName = func(i int) string {
			return getKey(t.Field(i), useTag)
		}
	}

	for i := 0; i < t.NumField(); i++ {
		keyName := getKeyName(i)
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

type keyNameCache struct {
	mutex sync.Mutex

	// nameMap: type of struct -> map(tag -> resolved key names of the struct, array indexed by field index)
	nameMap map[reflect.Type]map[string][]string
}

func (cache *keyNameCache) GetKeyNames(t reflect.Type, useTag string) []string {
	names := cache.nameMap[t][useTag]
	if names != nil {
		return names
	}
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	// double check
	names = cache.nameMap[t][useTag]
	if names != nil {
		return names
	}
	// resolve names then set to cache
	names = cache.resolveKeyNamesOf(t, useTag)
	if m, ok := cache.nameMap[t]; ok {
		m[useTag] = names
	} else {
		m = make(map[string][]string)
		cache.nameMap[t] = m
		m[useTag] = names
	}
	return names
}

func (cache *keyNameCache) resolveKeyNamesOf(t reflect.Type, useTag string) []string {
	result := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		result[i] = getKey(field, useTag)
	}
	return result
}
