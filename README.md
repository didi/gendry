## Gendry
[![Build Status](https://www.travis-ci.org/didi/gendry.svg?branch=master)](https://www.travis-ci.org/didi/gendry)
[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://github.com/didi/gendry/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/didi/gendry?status.svg)](https://godoc.org/github.com/didi/gendry)

**gendry** is a Go library that helps you operate database. Based on `go-sql-driver/mysql`, it provides a series of simple but useful tools to prepare parameters for calling methods in standard library `database/sql`.

The name **gendry** comes from the role in the hottest drama `The Game of Throne`, in which Gendry is not only the bastardy of the late king Robert Baratheon but also a skilled blacksmith. Like the one in drama,this library also forge something which is called `SQL`.

**gendry** consists of three isolated parts, and you can use each one of them partially:

* [manager](#manager)
* [builder](#builder)
* [scanner](#scanner)
* [CLI tool](#tools)

### Translation
* [中文](translation/zhcn/README.md)



<h3 id="manager">Manager</h3>

manager is used for initializing database connection pool(i.e `sql.DB`),
you can set almost all parameters for those mysql driver supported.For example, initializing a database connection pool:

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
In fact, all things manager does is just for concatting the `dataSourceName`

the format of a `dataSourceName` is：

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
```

manager is based on `go-mysql-driver/mysql`,and if you don't know some of the manager.SetXXX series functions,see it on [mysql driver's github home page](https://github.com/go-sql-driver/mysql).And for more details see [manager's doc](manager/README.md)

<h3 id="builder">Builder</h3>
builder as its name says, is for building sql.Writing sql mannually is intuitive but somewhat difficult to maintain.And for `where in`,if you have huge amount of elements in the `in` set,it's very hard to write.

builder isn't an ORM, in fact one of the most important reasons we create Gendry is we don't like ORM. So Gendry just provides some simple APIs to help you building sqls:

```go
where := map[string]interface{}{
	"city in": []string{"beijing", "shanghai"},
	"score": 5,
	"age >": 35,
	"address": builder.IsNotNull,
	"_orderby": "bonus desc",
	"_groupby": "department",
}
table := "some_table"
selectFields := []string{"name", "age", "sex"}
cond, values, err := builder.BuildSelect(table, where, selectFields)

//cond = SELECT name,age,sex FROM g_xxx WHERE (score=? AND city IN (?,?) AND age>? AND address IS NOT NULL) GROUP BY department ORDER BY bonus DESC
//values = []interface{}{"beijing", "shanghai", 5, 35}

rows,err := db.Query(cond, values...)
```
And, the library provide a useful API for executing aggregate queries like count,sum,max,min,avg

```go
where := map[string]interface{}{
    "score > ": 100,
    "city in": []interface{}{"Beijing", "Shijiazhuang",}
}
// AggregateSum,AggregateMax,AggregateMin,AggregateCount,AggregateAvg are supported
result, err := AggregateQuery(ctx, db, "tableName", where, AggregateSum("age"))
sumAge := result.Int64()
result,err = AggregateQuery(ctx, db, "tableName", where, AggregateCount("*")) 
numberOfRecords := result.Int64()
result,err = AggregateQuery(ctx, db, "tableName", where, AggregateAvg("score"))
averageScore := result.Float64()
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
slice type can be expanded automatically according to its length, thus these sqls are very convenient for DBA to review.  
**For critical system, this is recommended**

For more detail, see [builder's doc](builder/README.md) or just use `godoc`

<h3 id="scanner">Scanner</h3>
For each response from mysql,you want to map it with your well-defined structure.
Scanner provides a very easy API to do this,it's based on reflection:

##### standard library
```go
type Person struct {
	Name string
	Age int
}

rows,err := db.Query("SELECT age as m_age,name from g_xxx where xxx")
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

rows,err := db.Query("SELECT age as m_age,name from g_xxx where xxx")
defer rows.Close()

var students []Person

scanner.Scan(rows, &students)
```
Types which implement the interface
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

func (ext *extraInfo) UnmarshalDB(data []byte) error {
	return json.Unmarshal(data, ext)
}

//if the type of ext column in a table is varchar(stored legal json string) or json(mysql5.7)
var student human
err := scanner.Scan(rows, &student)
// ...
```

The extra tag of the struct will be used by scanner resolve data from response.The default tag name is `ddb:"tagname"`,but you can specify your own such as:

``` go
scaner.SetTagName("json")
type Person struct {
	Name string `json:"name"`
	Age int `json:"m_age"`
}

// ...
var student Person
scaner.Scan(rows, &student)
```

**scanner.SetTagName is a global setting and it can be invoked only once**

#### ScanMap
```go
rows,_ := db.Query("select name,age as m_age from person")
result,err := scaner.ScanMap(rows)
for _,record := range result {
	fmt.Println(record["name"], record["m_age"])
}
```
For more detail,see [scanner's doc](scanner/README.md)

PS：

* Don't forget close rows if you don't use ScanXXXClose
* The second parameter of Scan must be a reference

<h3 id="tools">Tools</h3>
Besides APIs above, Gendry provide a [CLI tool](https://github.com/caibirdme/gforge) to help generating codes.





