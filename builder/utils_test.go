package builder

import (
	"context"
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestResultResolver(t *testing.T) {
	var testData = []struct {
		origin   interface{}
		intout   int64
		floatout float64
	}{
		{
			origin:   []byte{'+', '0', '5', '3', '2', '.', '0', '7'},
			intout:   532,
			floatout: 532.07,
		},
		{
			origin:   []byte{'+', '5', '3', '2', '.', '3', '7'},
			intout:   532,
			floatout: 532.37,
		},
		{
			origin:   []byte{'-', '5', '3', '2', '.', '3', '7'},
			intout:   -532,
			floatout: -532.37,
		},
		{
			origin:   []uint8{'-', '5', '3', '2', '.', '3', '7'},
			intout:   -532,
			floatout: -532.37,
		},
		{
			origin:   []uint8{'5', '3', '2', '.', '3', '7'},
			intout:   532,
			floatout: 532.37,
		},
		{
			origin:   10,
			intout:   10,
			floatout: 10.0,
		},
		{
			origin:   4.5,
			intout:   4,
			floatout: 4.5,
		},
		{
			origin:   []uint8{'5', '3', '2'},
			intout:   532,
			floatout: 532.0,
		},
	}
	ass := assert.New(t)
	for idx, tc := range testData {
		rr := resultResolve{tc.origin}
		ass.Equal(tc.intout, rr.Int64(), "case#%d fail", idx)
		ass.True(math.Abs(tc.floatout-rr.Float64()) < 1e-5, "case#%d fail", idx)
	}
}

func TestAggregateQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if nil != err {
		t.Error(err)
	}
	var testData = []struct {
		rows     *sqlmock.Rows
		reg      string
		ag       AggregateSymbleBuilder
		intout   int64
		floatout float64
		where    map[string]interface{}
	}{
		{
			sqlmock.NewRows([]string{"count(*)"}).AddRow(12),
			"SELECT count\\(\\*\\) FROM tb1",
			AggregateCount("*"),
			12,
			12,
			nil,
		},
		{
			sqlmock.NewRows([]string{"sum(age)"}).AddRow(math.MaxInt64),
			"SELECT sum\\(age\\) FROM tb1",
			AggregateSum("age"),
			math.MaxInt64,
			float64(math.MaxInt64),
			nil,
		},
		{
			sqlmock.NewRows([]string{"avg(age)"}).AddRow(100.957),
			"SELECT avg\\(age\\) FROM tb1",
			AggregateAvg("age"),
			100,
			100.957,
			nil,
		},
		{
			sqlmock.NewRows([]string{"sum(age)"}).AddRow(100.957),
			"sum\\(age\\)",
			AggregateSum("age"),
			100,
			100.957,
			nil,
		},
		{
			sqlmock.NewRows([]string{"max(age)"}).AddRow(math.Pi),
			"max\\(age\\)",
			AggregateMax("age"),
			3,
			math.Pi,
			nil,
		},
		{
			sqlmock.NewRows([]string{"min(age)"}).AddRow(math.Pi),
			"min\\(age\\) .+foo",
			AggregateMin("age"),
			3,
			math.Pi,
			map[string]interface{}{"foo": "hello", "bar": 123},
		},
	}
	ass := assert.New(t)
	ctx := context.Background()
	for _, tc := range testData {
		mock.ExpectQuery(tc.reg).WillReturnRows(tc.rows)
		result, err := AggregateQuery(ctx, db, "tb1", tc.where, tc.ag)
		ass.NoError(err)
		ass.NoError(mock.ExpectationsWereMet())
		ass.Equal(tc.intout, result.Int64())
		ass.True(math.Abs(result.Float64()-tc.floatout) < 1e6)
	}
}

func TestOmitEmpty(t *testing.T) {
	var (
		m  map[string]string
		sl []string
		i  interface{}
	)

	type stru struct {
		x string
		y int
	}

	var testData = []struct {
		where      map[string]interface{}
		omitKey    []string
		finalWhere map[string]interface{}
	}{
		// Bool
		{
			map[string]interface{}{"b1": true, "b2": false},
			[]string{"b1", "b2", "b2"},
			map[string]interface{}{"b1": true},
		},
		// Array, String
		{
			map[string]interface{}{"a1": [0]string{}, "a2": [...]string{"2"}},
			[]string{"a1", "a2"},
			map[string]interface{}{"a2": [...]string{"2"}},
		},
		// Float32, Float64
		{
			map[string]interface{}{"f1": float32(0), "f2": float32(1.1)},
			[]string{"f1", "f2"},
			map[string]interface{}{"f2": float32(1.1)},
		},
		{
			map[string]interface{}{"f1": float64(0), "f2": float64(1.1)},
			[]string{"f1", "f2"},
			map[string]interface{}{"f2": float64(1.1)},
		},
		// Int, Int8, Int16, Int32, Int64
		{
			map[string]interface{}{"i8": int8(0), "i8_1": int8(8)},
			[]string{"i8", "i8_1"},
			map[string]interface{}{"i8_1": int8(8)},
		},
		{
			map[string]interface{}{"i16": int16(0), "i16_1": int16(16)},
			[]string{"i16", "i16_1"},
			map[string]interface{}{"i16_1": int16(16)},
		},
		{
			map[string]interface{}{"i32": int32(0), "i32_1": int32(32)},
			[]string{"i32", "i32_1"},
			map[string]interface{}{"i32_1": int32(32)},
		},
		{
			map[string]interface{}{"i64": int64(0), "i64_1": int64(64)},
			[]string{"i64", "i64_1"},
			map[string]interface{}{"i64_1": int64(64)},
		},
		// Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
		{
			map[string]interface{}{"ui": uint(0), "ui_1": uint(1)},
			[]string{"ui", "ui_1"},
			map[string]interface{}{"ui_1": uint(1)},
		},
		{
			map[string]interface{}{"ui8": uint8(0), "ui8_1": uint8(8)},
			[]string{"ui8", "ui8_1"},
			map[string]interface{}{"ui8_1": uint8(8)},
		},
		{
			map[string]interface{}{"ui16": uint16(0), "ui16_1": uint16(16)},
			[]string{"ui16", "ui16_1"},
			map[string]interface{}{"ui16_1": uint16(16)},
		},
		{
			map[string]interface{}{"ui32": uint32(0), "ui32_1": uint32(32)},
			[]string{"ui32", "ui32_1"},
			map[string]interface{}{"ui32_1": uint32(32)},
		},
		{
			map[string]interface{}{"ui64": uint64(0), "ui64_1": uint64(64)},
			[]string{"ui64", "ui64_1"},
			map[string]interface{}{"ui64_1": uint64(64)},
		},
		{
			map[string]interface{}{"uip": uintptr(0), "uip_1": uintptr(1)},
			[]string{"uip", "uip_1"},
			map[string]interface{}{"uip_1": uintptr(1)},
		},
		// Map, Slice, Interface
		{
			map[string]interface{}{"m1": m, "m2": map[string]string{"foo": "hi"}, "m3": map[string]string{}},
			[]string{"m1", "m2", "m3"},
			map[string]interface{}{"m2": map[string]string{"foo": "hi"}},
		},
		{
			map[string]interface{}{"sl1": sl, "sl2": []string{"sl"}, "sl3": []int{}},
			[]string{"sl1", "sl2", "sl3"},
			map[string]interface{}{"sl2": []string{"sl"}},
		},
		{
			map[string]interface{}{"i": i},
			[]string{"i"},
			map[string]interface{}{},
		},
		// struct
		{
			map[string]interface{}{"stru1": stru{x: "s", y: 0}, "stru2": stru{x: "", y: 0}},
			[]string{"stru1", "stru2"},
			map[string]interface{}{"stru1": stru{x: "s", y: 0}},
		},
		// implement zero
		{
			map[string]interface{}{"time": time.Time{}},
			[]string{"time"},
			map[string]interface{}{},
		},
	}
	ass := assert.New(t)
	for _, tc := range testData {
		r := OmitEmpty(tc.where, tc.omitKey)
		ass.True(reflect.DeepEqual(tc.finalWhere, r))
	}
}

func TestCustom(t *testing.T) {
	// Custom in where
	type inStruct struct {
		table  string
		where  map[string]interface{}
		fields []string
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":       "bar",
					"_custom_1": Custom("(x=? OR y=?)", 20, "1"),
					"age in":    []interface{}{1, 3, 5, 7, 9},
					"vx":        []interface{}{1, 3, 5},
					"faith <>":  "Muslim",
					"_or": []map[string]interface{}{
						{
							"aa": 11,
							"bb": "xswl",
						},
						{
							"cc":    "234",
							"dd in": []interface{}{7, 8},
							"_or": []map[string]interface{}{
								{
									"neeest_ee <>": "dw42",
									"neeest_ff in": []interface{}{34, 59},
								},
								{
									"neeest_gg":        1259,
									"neeest_hh not in": []interface{}{358, 1245},
								},
							},
						},
					},
					"_orderby":  "age DESC,score ASC",
					"_groupby":  "department",
					"_limit":    []uint{0, 100},
					"_custom_2": Custom("(xx=? AND yy=?)", 20, "2"),
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE ((x=? OR y=?) AND (xx=? AND yy=?) AND ((aa=? AND bb=?) OR (((neeest_ff IN (?,?) AND neeest_ee!=?) OR (neeest_gg=? AND neeest_hh NOT IN (?,?))) AND cc=? AND dd IN (?,?))) AND foo=? AND age IN (?,?,?,?,?) AND vx IN (?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC,score ASC LIMIT ?,?",
				vals: []interface{}{20, "1", 20, "2", 11, "xswl", 34, 59, "dw42", 1259, 358, 1245, "234", 7, 8, "bar", 1, 3, 5, 7, 9, 1, 3, 5, "Muslim", 0, 100},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"name like": "%123",
					"_custom_1": Custom("(x=? OR y=?)", 20, "1"),
				},
				fields: nil,
			},
			out: outStruct{
				cond: `SELECT * FROM tb WHERE ((x=? OR y=?) AND name LIKE ?)`,
				vals: []interface{}{20, "1", "%123"},
				err:  nil,
			},
		},

		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":       "bar",
					"_custom_1": Custom("(x=? OR y=?)", 20, "1"),
					"_orderby":  "  ",
				},
				fields: nil,
			},
			out: outStruct{
				cond: "SELECT * FROM tb WHERE ((x=? OR y=?) AND foo=?)",
				vals: []interface{}{20, "1", "bar"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_custom_0":  JsonContains("my_json->'$.data.list'", []int{1, 0}),
					"_custom_12": Custom("x=?", 20),
					"_custom_1":  Custom("(age=? OR name=?)", 20, "test"),
				},
				fields: nil,
			},
			out: outStruct{
				cond: "SELECT * FROM tb WHERE (JSON_CONTAINS(my_json->'$.data.list',JSON_ARRAY(?,?)) AND (age=? OR name=?) AND x=?)",
				vals: []interface{}{1, 0, 20, "test", 20},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := BuildSelect(tc.in.table, tc.in.where, tc.in.fields)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}

	// Custom in update
	{
		update := map[string]interface{}{
			"name":     "name000",
			"_custom_": Custom("a=a*?,aa=999", 10),
		}
		where := map[string]interface{}{
			"id": 5,
		}
		cond, vals, err := BuildUpdate("xx", where, update)
		ass.NoError(err)
		ass.Equal(
			"UPDATE xx SET a=a*?,aa=999,name=? WHERE (id=?)",
			cond)
		ass.Equal(
			[]interface{}{10, "name000", 5},
			vals)
	}

	// Custom both in update and where
	{
		update := map[string]interface{}{
			"name":     "name000",
			"_custom_": Custom("a=a*?,aa=999", 10),
		}
		where := map[string]interface{}{
			"_custom_": Custom("json_contains(my_json,cast(? as json))", 10),
			"name !=":  "",
		}
		cond, vals, err := BuildUpdate("xx", where, update)
		ass.NoError(err)
		ass.Equal(
			"UPDATE xx SET a=a*?,aa=999,name=? WHERE (json_contains(my_json,cast(? as json)) AND name!=?)",
			cond)
		ass.Equal(
			[]interface{}{10, "name000", 10, ""},
			vals)
	}
}

func TestGenJsonObj(t *testing.T) {
	json := map[string]interface{}{
		"a": "--a",
		"b": 1.5,
		"c": []map[string]interface{}{
			{"a": "ca0", "b": 2, "c": nil},
			{"a": "ca1", "b": 1, "c": []map[string]interface{}{
				{"a": ";", "b": true},
			},
			},
		},
		"d": map[string]interface{}{
			"a": "da",
		},
	}

	validMapOrder := make(map[string]int, 11)
	for i := 0; i < 11; i++ {
		validMapOrder["k"+strconv.Itoa(i)] = i
	}

	const testCount = 100
	testData := []struct {
		in     interface{}
		outSql string
		outVal []interface{}
	}{
		{18, "?", []interface{}{18}},
		{false, "false", []interface{}(nil)},
		{nil, "null", []interface{}(nil)},
		{[]int{1, 2, 3}, "JSON_ARRAY(?,?,?)", []interface{}{1, 2, 3}},
		{validMapOrder,
			"JSON_OBJECT(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
			[]interface{}{"k0", 0, "k1", 1, "k10", 10, "k2", 2, "k3", 3, "k4", 4, "k5", 5, "k6", 6, "k7", 7, "k8", 8, "k9", 9},
		},
		{json,
			"JSON_OBJECT(?,?,?,?,?,JSON_ARRAY(JSON_OBJECT(?,?,?,?,?,null),JSON_OBJECT(?,?,?,?,?,JSON_ARRAY(JSON_OBJECT(?,?,?,true)))),?,JSON_OBJECT(?,?))",
			[]interface{}{"a", "--a", "b", 1.5, "c", "a", "ca0", "b", 2, "c", "a", "ca1", "b", 1, "c", "a", ";", "b", "d", "a", "da"},
		},
	}
	ass := assert.New(t)
	for i := 0; i < testCount; i++ {
		for _, v := range testData {
			sql, val := genJsonObj(v.in)
			ass.Equal(v.outSql, sql)
			ass.Equal(v.outVal, val)
		}
	}
}

func TestJsonArrayAppend(t *testing.T) {
	type inStruct struct {
		table  string
		update map[string]interface{}
		where  map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var testData = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				update: map[string]interface{}{
					"_custom_0": JsonArrayAppend("my_json", "$", 0, "$", 1),
					"name":      "newName",
				},
				where: map[string]interface{}{
					"id": 3,
				},
			},
			out: outStruct{
				cond: "UPDATE tb SET my_json=JSON_ARRAY_APPEND(my_json,'$',?,'$',?),name=? WHERE (id=?)",
				vals: []interface{}{0, 1, "newName", 3},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				update: map[string]interface{}{
					"_custom_0": JsonArrayAppend("my_json", "$", 0, "$[0]", []int{7, 8, 9}),
				},
				where: map[string]interface{}{
					"id": []int{4, 5, 6},
				},
			},
			out: outStruct{
				cond: "UPDATE tb SET my_json=JSON_ARRAY_APPEND(my_json,'$',?,'$[0]',JSON_ARRAY(?,?,?)) WHERE (id IN (?,?,?))",
				vals: []interface{}{0, 7, 8, 9, 4, 5, 6},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)

	for _, v := range testData {
		sql, val, err := BuildUpdate(v.in.table, v.in.where, v.in.update)
		ass.Equal(v.out.cond, sql)
		ass.Equal(v.out.vals, val)
		ass.NoError(err)
	}

}

func TestJsonSet(t *testing.T) {
	type inStruct struct {
		table  string
		update map[string]interface{}
		where  map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var testData = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "xx",
				update: map[string]interface{}{
					"_custom_0": JsonSet("my_json", "$.a", 0, "$.friend_id", []int{7, 8, 9}),
					"name":      "newName",
				},
				where: map[string]interface{}{
					"id": 8,
				},
			},
			out: outStruct{
				cond: "UPDATE xx SET my_json=JSON_SET(my_json,'$.a',?,'$.friend_id',JSON_ARRAY(?,?,?)),name=? WHERE (id=?)",
				vals: []interface{}{0, 7, 8, 9, "newName", 8},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "xx",
				update: map[string]interface{}{
					"_custom_0": JsonSet("my_json", "$[0]", []int{7, 8, 9}),
				},
				where: map[string]interface{}{
					"id": []int{4, 5, 6},
				},
			},
			out: outStruct{
				cond: "UPDATE xx SET my_json=JSON_SET(my_json,'$[0]',JSON_ARRAY(?,?,?)) WHERE (id IN (?,?,?))",
				vals: []interface{}{7, 8, 9, 4, 5, 6},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)

	for _, v := range testData {
		sql, val, err := BuildUpdate(v.in.table, v.in.where, v.in.update)
		ass.Equal(v.out.cond, sql)
		ass.Equal(v.out.vals, val)
		ass.NoError(err)
	}

}

func TestJsonContains(t *testing.T) {
	testData := []struct {
		fullJsonPath string
		jsonLike     interface{}

		outSql string
		outVal []interface{}
	}{
		{
			fullJsonPath: "my_json->'$'",
			jsonLike:     18,
			outSql:       "(? MEMBER OF(my_json->'$'))",
			outVal:       []interface{}{18},
		},
		{
			fullJsonPath: "my_json->'$'",
			jsonLike:     []interface{}{18, "xxx"},
			outSql:       "JSON_CONTAINS(my_json->'$',JSON_ARRAY(?,?))",
			outVal:       []interface{}{18, "xxx"},
		},
		{
			fullJsonPath: "my_json",
			jsonLike:     true,
			outSql:       "(true MEMBER OF(my_json))",
			outVal:       nil,
		},
		{
			fullJsonPath: "my_json->'$'",
			jsonLike:     nil,
			outSql:       "JSON_CONTAINS(my_json->'$','null')",
			outVal:       nil,
		},
		{
			fullJsonPath: "my_json->'$.friend'",
			jsonLike:     map[string]interface{}{"name": "A", "age": 88},
			outSql:       "JSON_CONTAINS(my_json->'$.friend',JSON_OBJECT(?,?,?,?))",
			outVal:       []interface{}{"age", 88, "name", "A"},
		},
	}
	ass := assert.New(t)

	for _, v := range testData {
		sql, val := JsonContains(v.fullJsonPath, v.jsonLike).Build()
		ass.Equal(v.outSql, sql[0])
		ass.Equal(v.outVal, val)
	}
}

func TestJsonRemove(t *testing.T) {
	type inStruct struct {
		table  string
		update map[string]interface{}
		where  map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var testData = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "xx",
				update: map[string]interface{}{
					"_custom_0": JsonRemove("my_json", "$.unused", "$.list[last]"),
					"name":      "newName",
				},
				where: map[string]interface{}{
					"_custom_0": JsonContains("my_json->'$.list[last]'", nil),
				},
			},
			out: outStruct{
				cond: "UPDATE xx SET my_json=JSON_REMOVE(my_json,'$.unused','$.list[last]'),name=? WHERE (JSON_CONTAINS(my_json->'$.list[last]','null'))",
				vals: []interface{}{"newName"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "xx",
				update: map[string]interface{}{
					"_custom_0": JsonRemove("my_json"),
				},
				where: map[string]interface{}{
					"id": []int{4, 5, 6},
				},
			},
			out: outStruct{
				cond: "UPDATE xx SET my_json=my_json WHERE (id IN (?,?,?))",
				vals: []interface{}{4, 5, 6},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)

	for _, v := range testData {
		sql, val, err := BuildUpdate(v.in.table, v.in.where, v.in.update)
		ass.Equal(v.out.cond, sql)
		ass.Equal(v.out.vals, val)
		ass.NoError(err)
	}

}

func TestJsonArrayInsert(t *testing.T) {
	type inStruct struct {
		table  string
		update map[string]interface{}
		where  map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var testData = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "xx",
				update: map[string]interface{}{
					"a":         18,
					"_custom_0": JsonArrayInsert("my_json", "$[0]", "first", "$[1]", 2),
					"name":      "newName",
				},
				where: map[string]interface{}{
					"_or": []map[string]interface{}{
						{"_custom_0": JsonContains("my_json->'$[0]'", 1)},
						{"_custom_0": JsonContains("my_json->'$[0]'", 7)},
					},
				},
			},
			out: outStruct{
				cond: "UPDATE xx SET my_json=JSON_ARRAY_INSERT(my_json,'$[0]',?,'$[1]',?),a=?,name=? WHERE ((((? MEMBER OF(my_json->'$[0]'))) OR ((? MEMBER OF(my_json->'$[0]')))))",
				vals: []interface{}{"first", 2, 18, "newName", 1, 7},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "xx",
				update: map[string]interface{}{
					"_custom_0": JsonArrayInsert("my_json", "$[0]", 0, "$[0]", []int{7, 8, 9}),
				},
				where: map[string]interface{}{
					"id": []int{4, 5, 6},
				},
			},
			out: outStruct{
				cond: "UPDATE xx SET my_json=JSON_ARRAY_INSERT(my_json,'$[0]',?,'$[0]',JSON_ARRAY(?,?,?)) WHERE (id IN (?,?,?))",
				vals: []interface{}{0, 7, 8, 9, 4, 5, 6},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)

	for _, v := range testData {
		cond, val, err := BuildUpdate(v.in.table, v.in.where, v.in.update)
		ass.Equal(v.out.cond, cond)
		ass.Equal(v.out.vals, val)
		ass.NoError(err)
	}
}

func TestOrAndCustomJsonContains(t *testing.T) {
	where := map[string]interface{}{
		"_or1":       []map[string]interface{}{{"a": 1}, {"b": 1}},
		"_or2":       []map[string]interface{}{{"aa": 1}, {"bb": 1}},
		"_or3":       []map[string]interface{}{{"aaa": 1}, {"bbb": 1}},
		"_custom_0":  JsonContains("my_json", []int(nil)),
		"_custom_01": JsonContains("my_json", map[string]interface{}{}),
		"_custom_1":  JsonContains("my_json->'$'", 8),
		"_custom_3":  JsonContains("my_json->'$[last]'", map[string]interface{}{"a": 1, "b": 2}),
		"_custom_4":  JsonContains("my_json", []int{1, 3}),
		"_custom_5":  JsonContains("my_json->'$[0]'", nil),
	}
	s := "SELECT * FROM xx WHERE (JSON_CONTAINS(my_json,JSON_ARRAY()) AND JSON_CONTAINS(my_json,JSON_OBJECT()) AND (? MEMBER OF(my_json->'$')) AND JSON_CONTAINS(my_json->'$[last]',JSON_OBJECT(?,?,?,?)) AND JSON_CONTAINS(my_json,JSON_ARRAY(?,?)) AND JSON_CONTAINS(my_json->'$[0]','null') AND ((a=?) OR (b=?)) AND ((aa=?) OR (bb=?)) AND ((aaa=?) OR (bbb=?)))"
	v := []interface{}{8, "a", 1, "b", 2, 1, 3, 1, 1, 1, 1, 1, 1}
	ass := assert.New(t)

	s1, v1, err := BuildSelect("xx", where, nil)
	ass.Equal(s, s1)
	ass.Equal(v, v1)
	ass.NoError(err)

}
