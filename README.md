## Gendry
[![Build Status](https://github.com/didi/gendry/workflows/Go/badge.svg)](https://github.com/didi/gendry/actions)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/didi-gendry/Lobby)
[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://github.com/didi/gendry/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/didi/gendry?status.svg)](https://godoc.org/github.com/didi/gendry)

**gendry** is a Go library that helps you operate database. Based on `go-sql-driver/mysql`, it provides a series of simple but useful tools to prepare parameters for calling methods in standard library `database/sql`.

The name **gendry** comes from the role in the hottest drama: `The Game of Throne`, in which Gendry is not only the bastardy of the late king Robert Baratheon but also a skilled blacksmith. Like the one in drama, this library also forges something called `SQL`.

**gendry** consists of three isolated parts, and you can use each one of them partially:

* [manager](#manager)
* [builder](#builder)
* [scanner](#scanner)
* [CLI tool](#tools)

### Translation
* [Chinese Simplified (中文)](translation/zhcn/README.md)



<h3 id="manager">Manager</h3>

The manager is used for initializing the database connection pool(i.e., `sql.DB`).
You can set almost all parameters for those MySQL drivers supported. For example, initializing a database connection pool:

``` go
var db *sql.DB
var err error
db, err = manager
		.New(dbName, user, password, host)
		.Set(
			manager.SetCharset("utf8"),
			manager.SetAllowCleartextPasswords(true),
			manager.SetInterpolateParams(true),
			manager.SetTimeout(1 * time.Second),
			manager.SetReadTimeout(1 * time.Second)
		).Port(3302).Open(true)
```
In fact, all things manager does is just to generate the `dataSourceName`

the format of a `dataSourceName` is：

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
```

Manager is based on `go-mysql-driver/mysql`. If you don't know some of the manager and SetXXX series functions, see it on [mysql driver's github home page](https://github.com/go-sql-driver/mysql). For more details see [manager's doc](manager/README.md)

<h3 id="builder">Builder</h3>
Builder as its name says is for building SQL. Writing SQL manually is intuitive but somewhat difficult to maintain. And for `where in,` if you have a huge amount of elements in the `in` set, it's very hard to write.

Builder isn't an ORM. In fact, one of the most important reasons we create Gendry is we don't like ORM. So Gendry just provides some simple APIs to help you build SQLs:

```go
where := map[string]interface{}{
	"city": []string{"beijing", "shanghai"},
	// The in operator can be omitted by default,
	// which is equivalent to:
	// "city in": []string{"beijing", "shanghai"},
	"score": 5,
	"age >": 35,
	"address": builder.IsNotNull,
	"_or": []map[string]interface{}{
		{
			"x1":    11,
			"x2 >=": 45,
		},
		{
			"x3":    "234",
			"x4 <>": "tx2",
		},
	},
	"_orderby": "bonus desc",
	"_groupby": "department",
}
table := "some_table"
selectFields := []string{"name", "age", "sex"}
cond, values, err := builder.BuildSelect(table, where, selectFields)

//cond = SELECT name,age,sex FROM some_table WHERE (((x1=? AND x2>=?) OR (x3=? AND x4!=?)) AND score=? AND city IN (?,?) AND age>? AND address IS NOT NULL) GROUP BY department ORDER BY bonus DESC
//values = []interface{}{11, 45, "234", "tx2", 5, "beijing", "shanghai", 35}

rows, err := db.Query(cond, values...)
```

In the `where` param, `in` operator is automatically added by value type(reflect.Slice).

```go
where := map[string]interface{}{
	"city": []string{"beijing", "shanghai"},
}
```
the same as
```go
where := map[string]interface{}{
	"city in": []string{"beijing", "shanghai"},
}
```

Besides, the library provide a useful API for executing aggregate queries like count, sum, max, min, avg

```go
where := map[string]interface{}{
    "score > ": 100,
    "city": []interface{}{"Beijing", "Shijiazhuang", }
}
// AggregateSum, AggregateMax, AggregateMin, AggregateCount, AggregateAvg are supported
result, err := AggregateQuery(ctx, db, "tableName", where, AggregateSum("age"))
sumAge := result.Int64()
result, err = AggregateQuery(ctx, db, "tableName", where, AggregateCount("*")) 
numberOfRecords := result.Int64()
result, err = AggregateQuery(ctx, db, "tableName", where, AggregateAvg("score"))
averageScore := result.Float64()
```

multi `or` condition can use multi `_or` prefix string mark 
``` go
where := map[string]interface{}{
    // location
    "_or_location": []map[string]interface{}{{
        "subway": "beijing_15", 
    }, {
        "district": "Chaoyang", 
    }},
    // functions
    "_or_functions": []map[string]interface{}{{
         "has_gas": true,
    }, {
        "has_lift": true,
}}}

// query = (((subway=?) OR (district=?)) AND ((has_gas=?) OR (has_lift=?)))
// args = ["beijing_15", "Chaoyang", true, true]
```

If you want to clear the value '0' in the where map, you can use builder.OmitEmpty

``` go
where := map[string]interface{}{
		"score": 0,
		"age >": 35,
	}
finalWhere := builder.OmitEmpty(where, []string{"score", "age"})
// finalWhere = map[string]interface{}{"age >": 35}

// support: Bool, Array, String, Float32, Float64, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr, Map, Slice, Interface, Struct
```

For complex queries, `NamedQuery` may be helpful:
```go
cond, vals, err := builder.NamedQuery("select * from tb where name={{name}} and id in (select uid from anothertable where score in {{m_score}})", map[string]interface{}{
	"name": "caibirdme",
	"m_score": []float64{3.0, 5.8, 7.9},
})

assert.Equal("select * from tb where name=? and id in (select uid from anothertable where score in (?,?,?))", cond)
assert.Equal([]interface{}{"caibirdme", 3.0, 5.8, 7.9}, vals)
```
slice type can be expanded automatically according to its length. Thus, these SQLs are very convenient for DBA to review.  
**For critical system, this is recommended**

For more detail, see [builder's doc](builder/README.md) or just use `godoc`

<h3 id="scanner">Scanner</h3>
For each response from MySQL, you want to map it with your well-defined structure.
Scanner provides a straightforward API to do this, it's based on reflection:

##### standard library
```go
type Person struct {
	Name string
	Age int
}

rows, err := db.Query("SELECT age as m_age, name from g_xxx where xxx")
defer rows.Close()

var students []Person

for rows.Next() {
	var student Person
	rows.Scan(student.Age, student.Name)
	students = append(students, student)
}
```
##### using scanner
```go
type Person struct {
	Name string `ddb:"name"`
	Age int `ddb:"m_age"`
}

rows, err := db.Query("SELECT age as m_age, name from g_xxx where xxx")
defer rows.Close()

var students []Person

scanner.Scan(rows, &students)
```
Types that implement the interface
```go
type ByteUnmarshaler interface {
	UnmarshalByte(data []byte) error
}
```
will take over the corresponding unmarshal work.

```go
type human struct {
	Age   int       `ddb:"ag"`
	Extra *extraInfo `ddb:"ext"`
}

type extraInfo struct {
	Hobbies     []string `json:"hobbies"`
	LuckyNumber int      `json:"ln"`
}

func (ext *extraInfo) UnmarshalByte(data []byte) error {
	return json.Unmarshal(data, ext)
}

//if the type of ext column in a table is varchar(stored legal json string) or json(mysql5.7)
var student human
err := scanner.Scan(rows, &student)
// ...
```

The extra tag of the struct will be used by scanner resolve data from response.The default tag name is `ddb:"tagname"`, but you can specify your own such as:

``` go
scanner.SetTagName("json")
type Person struct {
	Name string `json:"name"`
	Age int `json:"m_age"`
}

// ...
var student Person
scanner.Scan(rows, &student)
```

**scanner.SetTagName is a global setting and it can be invoked only once**

#### ScanMap
```go
rows, _ := db.Query("select name, age as m_age from person")
result, err := scanner.ScanMap(rows)
for _, record := range result {
	fmt.Println(record["name"], record["m_age"])
}
```
ScanMap scans data from rows and returns a `[]map[string]interface{}`  
int, float, string type may be stored as []uint8 by MySQL driver. ScanMap copies those values into the map. If you're sure that there's no binary data type in your MySQL table(in most cases, this is true), you can use ScanMapDecode instead which will convert []uint8 to int, float64 or string

For more detail, see [scanner's doc](scanner/README.md)

PS：

* Don't forget close rows if you don't use ScanXXXClose
* The second parameter of Scan must be a reference

<h3 id="tools">Tools</h3>
Besides APIs above, Gendry provides a [CLI tool](https://github.com/caibirdme/gforge) to help generating codes.





