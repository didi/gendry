package builder

import (
	"context"
	"database/sql"
	"reflect"
	"strconv"
)

// AggregateQuery is a helper function to execute the aggregate query and return the result
func AggregateQuery(ctx context.Context, db *sql.DB, table string, where map[string]interface{}, aggregate AggregateSymbleBuilder) (ResultResolver, error) {
	cond, vals, err := BuildSelect(table, where, []string{aggregate.Symble()})
	if nil != err {
		return resultResolve{0}, err
	}
	rows, err := db.QueryContext(ctx, cond, vals...)
	if nil != err {
		return resultResolve{0}, err
	}
	var result interface{}
	for rows.Next() {
		err = rows.Scan(&result)
	}
	rows.Close()
	return resultResolve{result}, err
}

// ResultResolver is a helper for retrieving data
// caller should know the type and call the responding method
type ResultResolver interface {
	Int64() int64
	Float64() float64
}

type resultResolve struct {
	data interface{}
}

func (r resultResolve) Int64() int64 {
	switch t := r.data.(type) {
	case int64:
		return t
	case int32:
		return int64(t)
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case float32:
		return int64(t)
	case []uint8:
		i64, err := strconv.ParseInt(string(t), 10, 64)
		if nil != err {
			return int64(r.Float64())
		}
		return i64
	default:
		return 0
	}
}

// from go-mysql-driver/mysql the value returned could be int64 float64 float32

func (r resultResolve) Float64() float64 {
	switch t := r.data.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case []uint8:
		f64, _ := strconv.ParseFloat(string(t), 64)
		return f64
	default:
		return float64(r.Int64())
	}
}

// AggregateSymbleBuilder need to be implemented so that executor can
// get what should be put into `select Symble() from xxx where yyy`
type AggregateSymbleBuilder interface {
	Symble() string
}

type agBuilder string

func (a agBuilder) Symble() string {
	return string(a)
}

// AggregateCount count(col)
func AggregateCount(col string) AggregateSymbleBuilder {
	return agBuilder("count(" + col + ")")
}

// AggregateSum sum(col)
func AggregateSum(col string) AggregateSymbleBuilder {
	return agBuilder("sum(" + col + ")")
}

// AggregateAvg avg(col)
func AggregateAvg(col string) AggregateSymbleBuilder {
	return agBuilder("avg(" + col + ")")
}

// AggregateMax max(col)
func AggregateMax(col string) AggregateSymbleBuilder {
	return agBuilder("max(" + col + ")")
}

// AggregateMin min(col)
func AggregateMin(col string) AggregateSymbleBuilder {
	return agBuilder("min(" + col + ")")
}

// OmitEmpty is a helper function to clear where map zero value
func OmitEmpty(where map[string]interface{}, omitKey []string) map[string]interface{} {
	for _, key := range omitKey {
		v, ok := where[key]
		if !ok {
			continue
		}

		if isZero(reflect.ValueOf(v)) {
			delete(where, key)
		}
	}
	return where
}

// isZero reports whether a value is a zero value
// Including support: Bool, Array, String, Float32, Float64, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
// Map, Slice, Interface, Struct
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	case reflect.Interface:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}

	if v.Kind() != reflect.Struct {
		return false
	}

	// Traverse the Struct and only return true
	// if all of its fields return IsZero == true
	n := v.NumField()
	for i := 0; i < n; i++ {
		vf := v.Field(i)
		if !isZero(vf) {
			return false
		}
	}
	return true
}
