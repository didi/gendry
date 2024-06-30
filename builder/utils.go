package builder

import (
	"context"
	"database/sql"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

type IsZeroer interface {
	IsZero() bool
}

var IsZeroType = reflect.TypeOf((*IsZeroer)(nil)).Elem()

// isZero reports whether a value is a zero value
// Including support: Bool, Array, String, Float32, Float64, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
// Map, Slice, Interface, Struct
func isZero(v reflect.Value) bool {
	if v.IsValid() && v.Type().Implements(IsZeroType) {
		return v.Interface().(IsZeroer).IsZero()
	}
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

type rawSql struct {
	sqlCond string
	values  []interface{}
}

func (r rawSql) Build() ([]string, []interface{}) {
	return []string{r.sqlCond}, r.values
}

func Custom(query string, args ...interface{}) Comparable {
	return rawSql{sqlCond: query, values: args}
}

// JsonContains aim to check target json contains all items in given obj;if check certain value just use direct
// where := map[string]interface{}{"your_json_field.'$.path_to_key' =": val}
//
// notice: fullJsonPath should hard code, never from user input;
// jsonLike only support json element like array,map,string,number etc., struct input will result panic!!!
//
// usage where := map[string]interface{}{"_custom_xxx": builder.JsonContains("my_json->'$.my_data.list'", 7)}
//
// usage where := map[string]interface{}{"_custom_xxx": builder.JsonContains("my_json->'$'", []int{1,2})}
//
// usage where := map[string]interface{}{"_custom_xxx": builder.JsonContains("my_json->'$.user_info'", map[string]any{"name": "", "age": 18})}
func JsonContains(fullJsonPath string, jsonLike interface{}) Comparable {
	// MEMBER OF cant not deal null in json array
	if jsonLike == nil {
		return rawSql{
			sqlCond: "JSON_CONTAINS(" + fullJsonPath + ",'null')",
			values:  nil,
		}
	}

	s, v := genJsonObj(jsonLike)
	// jsonLike is number, string, bool
	_, ok := jsonLike.(string) // this check avoid eg jsonLike "JSONa"
	if ok || !strings.HasPrefix(s, "JSON") {
		return rawSql{
			sqlCond: "(" + s + " MEMBER OF(" + fullJsonPath + "))",
			values:  v,
		}
	}
	// jsonLike is array or map
	return rawSql{
		sqlCond: "JSON_CONTAINS(" + fullJsonPath + "," + s + ")",
		values:  v,
	}
}

// JsonSet aim to simply set/update json field operation;
//
// notice: jsonPath should hard code, never from user input;
//
// usage update := map[string]interface{}{"_custom_xxx": builder.JsonSet(field, "$.code", 1, "$.user_info", map[string]any{"name": "", "age": 18})}
func JsonSet(field string, pathAndValuePair ...interface{}) Comparable {
	return jsonUpdateCall("JSON_SET", field, pathAndValuePair...)
}

// JsonArrayAppend gen JsonObj and call MySQL JSON_ARRAY_APPEND function;
// usage update := map[string]interface{}{"_custom_xxx": builder.JsonArrayAppend(field, "$", 1, "$[last]", []string{"2","3"}}
func JsonArrayAppend(field string, pathAndValuePair ...interface{}) Comparable {
	return jsonUpdateCall("JSON_ARRAY_APPEND", field, pathAndValuePair...)
}

// JsonArrayInsert gen JsonObj and call MySQL JSON_ARRAY_INSERT function; insert at index
// usage update := map[string]interface{}{"_custom_xxx": builder.JsonArrayInsert(field, "$[0]", 1, "$[0]", []string{"2","3"}}
func JsonArrayInsert(field string, pathAndValuePair ...interface{}) Comparable {
	return jsonUpdateCall("JSON_ARRAY_INSERT", field, pathAndValuePair...)
}

// JsonRemove call MySQL JSON_REMOVE function; remove element from Array or Map
// path removed in order, prev remove affect the later operation, maybe the array shrink
//
// remove last array element; update := map[string]interface{}{"_custom_xxx":builder.JsonRemove(field,'$.list[last]')}
// remove element; update := map[string]interface{}{"_custom_xxx":builder.JsonRemove(field,'$.key0')}
func JsonRemove(field string, path ...string) Comparable {
	if len(path) == 0 {
		// do nothing, update xxx set a=a;
		return rawSql{
			sqlCond: field + "=" + field,
			values:  nil,
		}
	}
	return rawSql{
		sqlCond: field + "=JSON_REMOVE(" + field + ",'" + strings.Join(path, "','") + "')",
		values:  nil,
	}
}

// jsonUpdateCall build args then call fn
func jsonUpdateCall(fn string, field string, pathAndValuePair ...interface{}) Comparable {
	if len(pathAndValuePair) == 0 || len(pathAndValuePair)%2 != 0 {
		return rawSql{sqlCond: field, values: nil}
	}
	val := make([]interface{}, 0, len(pathAndValuePair)/2)
	var buf strings.Builder
	buf.WriteString(field)
	buf.WriteByte('=')
	buf.WriteString(fn + "(")
	buf.WriteString(field)
	for i := 0; i < len(pathAndValuePair); i += 2 {
		buf.WriteString(",'")
		buf.WriteString(pathAndValuePair[i].(string))
		buf.WriteString("',")

		jsonSql, jsonVals := genJsonObj(pathAndValuePair[i+1])
		buf.WriteString(jsonSql)
		val = append(val, jsonVals...)
	}
	buf.WriteByte(')')

	return rawSql{
		sqlCond: buf.String(),
		values:  val,
	}
}

// genJsonObj build MySQL JSON object using JSON_ARRAY, JSON_OBJECT or ?; return sql string and args
func genJsonObj(obj interface{}) (string, []interface{}) {
	if obj == nil {
		return "null", nil
	}
	rValue := reflect.Indirect(reflect.ValueOf(obj))
	rType := rValue.Kind()
	var s []string
	var vals []interface{}
	switch rType {
	case reflect.Array, reflect.Slice:
		s = append(s, "JSON_ARRAY(")
		length := rValue.Len()
		for i := 0; i < length; i++ {
			subS, subVals := genJsonObj(rValue.Index(i).Interface())
			s = append(s, subS, ",")
			vals = append(vals, subVals...)
		}

		if s[len(s)-1] == "," {
			s[len(s)-1] = ")"
		} else { // empty slice
			s = append(s, ")")
		}
	case reflect.Map:
		s = append(s, "JSON_OBJECT(")
		// sort keys in map to keep generate result same.
		keys := rValue.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		length := rValue.Len()
		for i := 0; i < length; i++ {
			k := keys[i]
			v := rValue.MapIndex(k)
			subS, subVals := genJsonObj(v.Interface())
			s = append(s, "?,", subS, ",")
			vals = append(vals, k.String())
			vals = append(vals, subVals...)
		}

		if s[len(s)-1] == "," {
			s[len(s)-1] = ")"
		} else { // empty map
			s = append(s, ")")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return "?", []interface{}{rValue.Interface()}
	case reflect.Bool:
		if rValue.Bool() {
			return "true", nil
		}
		return "false", nil
	default:
		panic("genJsonObj not support type: " + rType.String())
	}
	return strings.Join(s, ""), vals
}
