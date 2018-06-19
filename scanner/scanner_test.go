package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"time"

	"math"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestBindOne(t *testing.T) {
	type Person struct {
		Name string `ddb:"name"`
		Age  int    `ddb:"ag"`
	}
	var p Person
	name := "deen"
	age := 23
	var mp = map[string]interface{}{
		"name": name,
		"ag":   age,
	}
	err := bind(mp, &p)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(name, p.Name)
	ass.Equal(age, p.Age)
}

func TestBindOne_byte_string(t *testing.T) {
	type Person struct {
		Name string `ddb:"name"`
		Age  int    `ddb:"ag"`
	}
	var p Person
	name := []byte{'d', 'e', 'e', 'n'}
	age := 23
	var mp = map[string]interface{}{
		"name": name,
		"ag":   age,
	}
	err := bind(mp, &p)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(string(name), p.Name)
	ass.Equal(age, p.Age)
}

func TestBindOne_byte_uint8(t *testing.T) {
	type Person struct {
		Name []uint8 `ddb:"name"`
		Age  int     `ddb:"ag"`
	}
	var p Person
	name := []byte{'d', 'e', 'e', 'n'}
	age := 23
	var mp = map[string]interface{}{
		"name": name,
		"ag":   age,
	}
	err := bind(mp, &p)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(name, p.Name)
	ass.Equal(age, p.Age)
}

func TestBindOne_byte_uint8_pointer(t *testing.T) {
	type Person struct {
		Name []uint8 `ddb:"name"`
		Age  int     `ddb:"ag"`
	}
	p := new(Person)
	name := []byte{'d', 'e', 'e', 'n'}
	age := 23
	var mp = map[string]interface{}{
		"name": name,
		"ag":   age,
	}
	err := bind(mp, p)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(name, p.Name)
	ass.Equal(age, p.Age)
}

func TestBindOne_uint8_byte(t *testing.T) {
	type Person struct {
		Name []byte `ddb:"name"`
		Age  int    `ddb:"ag"`
	}
	var p Person
	name := []uint8{'d', 'e', 'e', 'n'}
	age := 23
	var mp = map[string]interface{}{
		"name": name,
		"ag":   age,
	}
	err := bind(mp, &p)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(name, p.Name)
	ass.Equal(age, p.Age)
}

func TestBindOne_float(t *testing.T) {
	type Person struct {
		Salary float64 `ddb:"sl"`
	}
	var p Person
	salary := 100.123
	var mp = map[string]interface{}{
		"sl": salary,
	}
	err := bind(mp, &p)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(salary, p.Salary)
}

func TestBindSlice(t *testing.T) {
	type Stu struct {
		Age int `ddb:"age"`
	}
	var students []Stu
	testCases := []int{1, 2, 3, 4, 5, 6, 9, 0, 7, 8}
	var data []map[string]interface{}
	for _, v := range testCases {
		data = append(data, map[string]interface{}{"age": v})
	}
	err := bindSlice(data, &students)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(len(testCases), len(students))
	for idx, p := range students {
		ass.Equal(testCases[idx], p.Age)
	}
}
func Test_Scan_PointerArr(t *testing.T) {
	type Stuu struct {
		Name   string  `ddb:"name"`
		Salary float32 `ddb:"sala"`
	}
	var stus []*Stuu
	var data []map[string]interface{}
	data = append(data,
		map[string]interface{}{
			"name": "name_1",
			"sala": float32(20.5),
		},
		map[string]interface{}{
			"name": "name_2",
			"sala": float32(30.82),
		},
		map[string]interface{}{
			"name": "name_3",
			"sala": float32(0.0),
		},
	)
	err := bindSlice(data, &stus)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(len(data), len(stus))
	for i := 0; i < len(stus); i++ {
		ass.Equal(data[i]["name"], stus[i].Name, "bind pointer name")
		ass.Equal(data[i]["sala"], stus[i].Salary, "bind pointer sala")
	}
}

func Test_Bind_Float32_2_Float64(t *testing.T) {
	type A struct {
		Num float64 `ddb:"num"`
	}
	var a A
	err := bind(map[string]interface{}{
		"num": float32(10.5),
	}, &a)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(float64(10.5), a.Num)
}

func Test_Bind_Float64_2_Float32(t *testing.T) {
	type A struct {
		Num float32 `ddb:"num"`
	}
	var a A
	err := bind(map[string]interface{}{
		"num": float64(10.5),
	}, &a)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(float32(10.5), a.Num)
}

func Test_Bind_int64_2_uint64(t *testing.T) {
	type A struct {
		Num uint64 `ddb:"num"`
		Age uint8  `ddb:"age"`
	}
	var a A
	err := bind(map[string]interface{}{
		"num": int64(10),
		"age": int64(20),
	}, &a)
	ass := assert.New(t)
	ass.NoError(err, `shouldn't be error when bind int64 to uint64`)
	ass.Equal(uint64(10), a.Num)
	ass.Equal(uint8(20), a.Age)
}

func Test_Ignore_Unexported_Field(t *testing.T) {
	type Person struct {
		Name string `ddb:"name"`
		age  int    `ddb:"age"`
	}
	var Tom Person
	var data = map[string]interface{}{
		"name": []byte("Tommmm"),
		"age":  int64(100),
	}
	err := bind(data, &Tom)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal(0, Tom.age)
	ass.Equal("Tommmm", Tom.Name)
}

func Test_Bind_Time_2_String(t *testing.T) {
	type Whatever struct {
		When string `ddb:"create_time"`
	}
	now := time.Now()
	var data = map[string]interface{}{
		"create_time": now,
	}
	var tObj Whatever
	ass := assert.New(t)
	err := bind(data, &tObj)
	ass.NoError(err, "time.Time should transform to string and bind to string type")
	ass.Equal(now.Format("2006-01-02 15:04:05"), tObj.When)
	type WillErr struct {
		When int `ddb:"create_time"`
	}
	var some WillErr
	err = bind(data, &some)
	ass.Error(err, "time.Time could only bind to time.Time&string type %v", some)
}

func Test_ScanMap(t *testing.T) {
	var testData = []struct {
		rows *sqlmock.Rows
		out  []map[string]interface{}
	}{
		{
			rows: sqlmock.NewRows([]string{"foo", "bar"}).AddRow(int64(1), int64(5)).AddRow(int64(3), int64(7)),
			out: []map[string]interface{}{
				{
					"foo": int64(1),
					"bar": int64(5),
				},
				{
					"foo": int64(3),
					"bar": int64(7),
				},
			},
		},
		{
			rows: sqlmock.NewRows([]string{"foo", "bar"}).AddRow(int64(1), 10.8).AddRow(int64(3), 20.7),
			out: []map[string]interface{}{
				{
					"foo": int64(1),
					"bar": 10.8,
				},
				{
					"foo": int64(3),
					"bar": 20.7,
				},
			},
		},
		{
			rows: sqlmock.NewRows([]string{"foo", "bar"}).AddRow("hello world", 10.8).AddRow("writing test is boring but can make your code more robust", 20.7),
			out: []map[string]interface{}{
				{
					"foo": "hello world",
					"bar": 10.8,
				},
				{
					"foo": "writing test is boring but can make your code more robust",
					"bar": 20.7,
				},
			},
		},
	}
	ass := assert.New(t)
	db, mock, err := sqlmock.New()
	ass.NoError(err)
	for _, tc := range testData {
		mock.ExpectQuery("select \\* from tb").WillReturnRows(tc.rows)
		rows, err := db.Query("select * from tb")
		ass.NoError(err)
		ass.NotNil(rows)
		ass.NoError(mock.ExpectationsWereMet())
		mpArr, err := ScanMap(rows)
		ass.NoError(err)
		ass.Equal(tc.out, mpArr)
	}
}

func Test_Slice_2_Int(t *testing.T) {
	type user struct {
		Age int `ddb:"age"`
	}
	var testData = []struct {
		in  []byte
		out int
		err error
	}{
		{
			in:  []byte{'1', '2', '3'},
			out: 123,
			err: nil,
		},
		{
			in:  []byte{'0', '2', '3'},
			out: 23,
			err: nil,
		},
		{
			in:  []byte{'0'},
			out: 0,
			err: nil,
		},
		{
			in:  []byte("9223372036854775807"),
			out: 9223372036854775807,
			err: nil,
		},
		{
			in:  []byte("9223372036854775808"),
			out: 9223372036854775807,
			err: errors.New("test"),
		},
	}
	var u user
	ass := assert.New(t)
	for _, tc := range testData {
		mp := map[string]interface{}{
			"age": tc.in,
		}
		err := bind(mp, &u)
		if tc.err == nil {
			ass.NoError(err)
		} else {
			ass.Error(err)
		}
		ass.Equal(tc.out, u.Age)
	}
}

func Test_Slice_2_UInt(t *testing.T) {
	type user struct {
		Age uint `ddb:"age"`
	}
	var testData = []struct {
		in  []byte
		out uint
		err error
	}{
		{
			in:  []byte{'1', '2', '3'},
			out: 123,
			err: nil,
		},
		{
			in:  []byte{'0', '2', '3'},
			out: 23,
			err: nil,
		},
		{
			in:  []byte{'0'},
			out: 0,
			err: nil,
		},
		{
			in:  []byte("9223372036854775807"),
			out: 9223372036854775807,
			err: nil,
		},
		{
			in:  []byte("9223372036854775808"),
			out: 9223372036854775808,
			err: nil,
		},
		{
			in:  []byte("18446744073709551615"),
			out: 18446744073709551615,
			err: nil,
		},
		{
			in:  []byte("18446744073709551616"),
			out: 18446744073709551615,
			err: errors.New("error"),
		},
		{
			in:  []byte("-1"),
			out: 0xffffffffffffffff,
			err: errors.New("error"),
		},
	}
	var u user
	ass := assert.New(t)
	for _, tc := range testData {
		mp := map[string]interface{}{
			"age": tc.in,
		}
		err := bind(mp, &u)
		if tc.err == nil {
			ass.NoError(err)
		} else {
			ass.Error(err)
		}
		ass.Equal(tc.out, u.Age)
	}
}

func Test_Slice_2_Float(t *testing.T) {
	type user struct {
		Score float64 `ddb:"score"`
	}
	var testData = []struct {
		in  []byte
		out float64
		err error
	}{
		{
			in:  []byte("123"),
			out: 123,
			err: nil,
		},
		{
			in:  []byte("023"),
			out: 23,
			err: nil,
		},
		{
			in:  []byte("0.1234"),
			out: 0.1234,
			err: nil,
		},
		{
			in:  []byte{'0'},
			out: 0,
			err: nil,
		},
		{
			in:  []byte("-5.76902"),
			out: -5.76902,
			err: nil,
		},
		{
			in:  []byte("-5.7ff902"),
			out: 0,
			err: errors.New("will error"),
		},
	}
	var u user
	ass := assert.New(t)
	for _, tc := range testData {
		mp := map[string]interface{}{
			"score": tc.in,
		}
		err := bind(mp, &u)
		if tc.err == nil {
			ass.NoError(err)
		} else {
			ass.Error(err)
		}
		ass.True(math.Abs(tc.out-u.Score) < 1e5)
	}
}

func Test_uint8_2_any(t *testing.T) {
	type user struct {
		Name  string  `ddb:"name"`
		Age   int     `ddb:"_age"`
		Score float64 `ddb:"sc"`
	}
	var testData = []struct {
		in  map[string]interface{}
		out user
		err error
	}{
		{
			in: map[string]interface{}{
				"name": []uint8("xxx"),
				"_age": []uint8("52"),
				"sc":   []uint8("3.7"),
			},
			out: user{
				Name:  "xxx",
				Age:   52,
				Score: 3.7,
			},
			err: nil,
		},
		{
			in: map[string]interface{}{
				"name": []byte("xxx"),
				"_age": []byte("52"),
				"sc":   []byte("3.7"),
			},
			out: user{
				Name:  "xxx",
				Age:   52,
				Score: 3.7,
			},
			err: nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range testData {
		var u user
		err := bind(tc.in, &u)
		if tc.err == nil {
			ass.NoError(err)
		} else {
			ass.Error(err)
		}
		ass.Equal(tc.out, u)
	}
}

func TestTagSetOnlyOnce(t *testing.T) {
	userDefinedTagName = "a"
	SetTagName("foo")
	assert.Equal(t, "a", userDefinedTagName)
	userDefinedTagName = ""
	SetTagName("foo")
	assert.Equal(t, "foo", userDefinedTagName)

}

type fakeRows struct {
	columns []string
	dataset [][]interface{}
	idx     int
}

var errCloseForTest = errors.New("just for test")

func (r *fakeRows) Close() error {
	return errCloseForTest
}

func (r *fakeRows) Columns() ([]string, error) {
	return r.columns, nil
}

func (r *fakeRows) Next() bool {
	return r.idx < len(r.dataset)
}

func (r *fakeRows) Scan(dt ...interface{}) (err error) {
	lendt := len(dt)
	lenfact := len(r.dataset[r.idx])
	if lendt != lenfact {
		return fmt.Errorf("sql: expected %d destination arguments in Scan, not %d", lenfact, lendt)
	}
	defer func() { r.idx++ }()
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	for i := 0; i < lendt; i++ {
		data := r.dataset[r.idx][i]
		reflect.ValueOf(dt[i]).Elem().Set(reflect.ValueOf(data))
	}
	return nil
}

func TestScanNotSettable(t *testing.T) {
	ass := assert.New(t)
	err := Scan(&fakeRows{}, nil)
	ass.Equal(ErrTargetNotSettable, err)
	var rows Rows
	err = Scan(rows, nil)
	ass.Equal(ErrTargetNotSettable, err)
}

func TestScanMapClose(t *testing.T) {
	var rows Rows
	_, err := ScanMapClose(rows)
	ass := assert.New(t)
	ass.Equal(ErrNilRows, err)
	scannn := &fakeRows{
		columns: []string{"foo", "bar"},
		dataset: [][]interface{}{
			{1, 2},
			{3, 4},
		},
	}
	result, err := ScanMapClose(scannn)
	ass.Equal(2, len(result))
	ass.Equal(errCloseForTest.Error(), err.Error())
	v, ok := result[0]["foo"]
	if !ass.True(ok) {
		return
	}
	ass.Equal(1, v)
	v, ok = result[1]["bar"]
	if !ass.True(ok) {
		return
	}
	ass.Equal(4, v)
}

func TestScanMock(t *testing.T) {
	ass := assert.New(t)
	scannn := &fakeRows{
		columns: []string{"name", "age"},
		dataset: [][]interface{}{
			{"deen", 23},
			{"caibirdme", 24},
		},
	}
	type curdBoy struct {
		Name string `ddb:"name"`
		Age  int    `ddb:"age"`
	}
	var boys []curdBoy
	userDefinedTagName = "ddb"
	err := Scan(scannn, &boys)
	ass.NoError(err)
	ass.Equal("deen", boys[0].Name)
	ass.Equal("caibirdme", boys[1].Name)
	ass.Equal(23, boys[0].Age)
	ass.Equal(24, boys[1].Age)
}

func TestScanEmpty(t *testing.T) {
	ass := assert.New(t)
	scannn := &fakeRows{}
	type curdBoy struct {
		Name string `ddb:"name"`
		Age  int    `ddb:"age"`
	}
	var boys []curdBoy
	userDefinedTagName = "ddb"
	err := Scan(scannn, &boys)
	ass.NoError(err)
	ass.Equal(0, len(boys))
	var boy curdBoy
	err = Scan(scannn, &boy)
	ass.Equal(ErrEmptyResult, err)
}

type human struct {
	Age   int        `ddb:"ag"`
	Extra *extraInfo `ddb:"ext"`
}

type extraInfo struct {
	Hobbies     []string `json:"hobbies"`
	LuckyNumber int      `json:"ln"`
}

func (ext *extraInfo) UnmarshalByte(data []byte) error {
	return json.Unmarshal(data, ext)
}

func TestUnmarshalByte(t *testing.T) {
	var testCase = []struct {
		mapv   map[string]interface{}
		expect human
		err    error
	}{
		{
			mapv: map[string]interface{}{
				"ag":  20,
				"ext": []byte(`{"ln":18, "hobbies": ["soccer", "swimming", "jogging"]}`),
			},
			expect: human{
				Age: 20,
				Extra: &extraInfo{
					LuckyNumber: 18,
					Hobbies:     []string{"soccer", "swimming", "jogging"},
				},
			},
			err: nil,
		},
		{
			mapv: map[string]interface{}{
				"ag":  20,
				"ext": []byte(`{"ln":18, illegalJSON, "hobbies": ["soccer", "swimming", "jogging"]}`),
			},
			expect: human{
				Age: 20,
			},
			err: errors.New("[scanner]: extraInfo.UnmarshalByte fail to unmarshal the bytes, err: invalid character 'i' looking for beginning of object key string"),
		},
		{
			mapv: map[string]interface{}{
				"ag":  20,
				"ext": []byte(`{"ln":18, "hobbies": ["soccer", "swimming", "jogging"]}`),
			},
			expect: human{
				Age: 20,
				Extra: &extraInfo{
					LuckyNumber: 18,
					Hobbies:     []string{"soccer", "swimming", "jogging"},
				},
			},
			err: nil,
		},
		{
			mapv: map[string]interface{}{
				"ag":  20,
				"ext": []byte("null"),
			},
			expect: human{
				Age:   20,
				Extra: &extraInfo{},
			},
			err: nil,
		},
	}
	ass := assert.New(t)
	for idx, tc := range testCase {
		var student human
		if idx >= 2 {
			student.Extra = &extraInfo{}
		}
		err := bind(tc.mapv, &student)
		ass.Equal(tc.err, err, "idx:%d", idx)
		ass.Equal(tc.expect, student, "idx:%d", idx)
	}
}

func TestScanClose(t *testing.T) {
	rows := &fakeRows{
		columns: []string{"foo", "bar"},
		dataset: [][]interface{}{
			[]interface{}{1, 2},
		},
	}
	var testObj = struct {
		Foo int `ddb:"foo"`
		Bar int `ddb:"bar"`
	}{}
	ass := assert.New(t)
	err := ScanClose(rows, &testObj)
	e, ok := err.(CloseErr)
	ass.True(ok)
	ass.Equal(errCloseForTest.Error(), e.Error())
	ass.Equal(1, testObj.Foo)
	ass.Equal(2, testObj.Bar)
}
