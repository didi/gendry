package scanner

import (
	"testing"

	"time"

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
				map[string]interface{}{
					"foo": int64(1),
					"bar": int64(5),
				},
				map[string]interface{}{
					"foo": int64(3),
					"bar": int64(7),
				},
			},
		},
		{
			rows: sqlmock.NewRows([]string{"foo", "bar"}).AddRow(int64(1), 10.8).AddRow(int64(3), 20.7),
			out: []map[string]interface{}{
				map[string]interface{}{
					"foo": int64(1),
					"bar": 10.8,
				},
				map[string]interface{}{
					"foo": int64(3),
					"bar": 20.7,
				},
			},
		},
		{
			rows: sqlmock.NewRows([]string{"foo", "bar"}).AddRow("hello world", 10.8).AddRow("writing test is boring but can make your code more robust", 20.7),
			out: []map[string]interface{}{
				map[string]interface{}{
					"foo": "hello world",
					"bar": 10.8,
				},
				map[string]interface{}{
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
