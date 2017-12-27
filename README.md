## Gendry
[![Build Status](https://www.travis-ci.org/didi/gendry.svg?branch=master)](https://www.travis-ci.org/didi/gendry)
[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://github.com/didi/gendry/blob/master/LICENSE)

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
	"city in": []interface{}{"beijing", "shanghai"},
	"score": 5,
	"age >": 35,
	"_orderby": "bonus desc",
	"_grouoby": "department",
}
table := "some_table"
selectFields := []string{"name", "age", "sex"}
cond, values, err := builder.BuildSelect(table, where, selectFields)

//cond = SELECT name,age,sex FROM g_xxx WHERE (city IN (?,?) AND score=? AND age>?) GROUP BY department ORDER BY bonus DESC
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

sqls with join or subquery always need pay more attention to optimize and so,they're **not** supported yet.For more detail, see [builder's doc](builder/README.md) or just use `godoc`

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
Besides APIs above, Gendry provide a CLI tool to help generating codes.

#### install
`go get -u github.com/didichuxing/gendry/tools`

#### usage

```
> tools -h
A collection of tools to generate code for operating database supported by Gendry

Options:

  -h, --help   display help information
  -v           version

Commands:

  help    display help information
  table   schema could generate go struct code for given table
  dao     dao generates code of dao layer by given table name
```

Get the subcommand help information

```
> tools help table
schema could generate go struct code for given table

Options:

  -d               database name
  -t               table name
  -u               user name
  -p               password
  -h[=localhost]   host
  -P[=3306]        port
```

Generate a struct for table

```
> tools table -uusername -ppassword -hip -dinformation_schema -tCOLUMNS

// COLUMNS is a mapping object for COLUMNS
type COLUMNS struct {
	TABLECATALOG string `json:"TABLE_CATALOG"
	TABLESCHEMA string `json:"TABLE_SCHEMA"
	TABLENAME string `json:"TABLE_NAME"
	COLUMNNAME string `json:"COLUMN_NAME"
	ORDINALPOSITION uint64 `json:"ORDINAL_POSITION"
	COLUMNDEFAULT string `json:"COLUMN_DEFAULT"
	ISNULLABLE string `json:"IS_NULLABLE"
	DATATYPE string `json:"DATA_TYPE"
	CHARACTERMAXIMUMLENGTH uint64 `json:"CHARACTER_MAXIMUM_LENGTH"
	CHARACTEROCTETLENGTH uint64 `json:"CHARACTER_OCTET_LENGTH"
	NUMERICPRECISION uint64 `json:"NUMERIC_PRECISION"
	NUMERICSCALE uint64 `json:"NUMERIC_SCALE"
	DATETIMEPRECISION uint64 `json:"DATETIME_PRECISION"
	CHARACTERSETNAME string `json:"CHARACTER_SET_NAME"
	COLLATIONNAME string `json:"COLLATION_NAME"
	COLUMNTYPE string `json:"COLUMN_TYPE"
	COLUMNKEY string `json:"COLUMN_KEY"
	EXTRA string `json:"EXTRA"
	PRIVILEGES string `json:"PRIVILEGES"
	COLUMNCOMMENT string `json:"COLUMN_COMMENT"
	GENERATIONEXPRESSION string `json:"GENERATION_EXPRESSION"
}
```

The produced struct could pass the examine of golint and govet

Generate codes of dao layer about one table

```
> tools dao -uusername -ppassword -hip -dinformation_schema -tCOLUMNS | gofmt
package COLUMNS

import (
	"database/sql"
	"errors"
	"github.com/didichuxing/gendry/builder"
	"github.com/didichuxing/gendry/scanner"
)

/*
	This code is generated by ddtool
*/

// COLUMNS is a mapping object for COLUMNS
type COLUMNS struct {
	TABLECATALOG           string `json:"TABLE_CATALOG"`
	TABLESCHEMA            string `json:"TABLE_SCHEMA"`
	TABLENAME              string `json:"TABLE_NAME"`
	COLUMNNAME             string `json:"COLUMN_NAME"`
	ORDINALPOSITION        uint64 `json:"ORDINAL_POSITION"`
	COLUMNDEFAULT          string `json:"COLUMN_DEFAULT"`
	ISNULLABLE             string `json:"IS_NULLABLE"`
	DATATYPE               string `json:"DATA_TYPE"`
	CHARACTERMAXIMUMLENGTH uint64 `json:"CHARACTER_MAXIMUM_LENGTH"`
	CHARACTEROCTETLENGTH   uint64 `json:"CHARACTER_OCTET_LENGTH"`
	NUMERICPRECISION       uint64 `json:"NUMERIC_PRECISION"`
	NUMERICSCALE           uint64 `json:"NUMERIC_SCALE"`
	DATETIMEPRECISION      uint64 `json:"DATETIME_PRECISION"`
	CHARACTERSETNAME       string `json:"CHARACTER_SET_NAME"`
	COLLATIONNAME          string `json:"COLLATION_NAME"`
	COLUMNTYPE             string `json:"COLUMN_TYPE"`
	COLUMNKEY              string `json:"COLUMN_KEY"`
	EXTRA                  string `json:"EXTRA"`
	PRIVILEGES             string `json:"PRIVILEGES"`
	COLUMNCOMMENT          string `json:"COLUMN_COMMENT"`
	GENERATIONEXPRESSION   string `json:"GENERATION_EXPRESSION"`
}

//GetOne gets one record from table COLUMNS by condition "where"
func GetOne(db *sql.DB, where map[string]interface{}) (*COLUMNS, error) {
	if nil == db {
		return nil, errors.New("sql.DB object couldn't be nil")
	}
	cond, vals, err := builder.BuildSelect("COLUMNS", where, nil)
	if nil != err {
		return nil, err
	}
	row, err := db.Query(cond, vals...)
	if nil != err || nil == row {
		return nil, err
	}
	defer row.Close()
	var res *COLUMNS
	err = scanner.Scan(row, &res)
	return res, err
}

//GetMulti gets multiple records from table COLUMNS by condition "where"
func GetMulti(db *sql.DB, where map[string]interface{}) ([]*COLUMNS, error) {
	if nil == db {
		return nil, errors.New("sql.DB object couldn't be nil")
	}
	cond, vals, err := builder.BuildSelect("COLUMNS", where, nil)
	if nil != err {
		return nil, err
	}
	row, err := db.Query(cond, vals...)
	if nil != err || nil == row {
		return nil, err
	}
	defer row.Close()
	var res []*COLUMNS
	err = scanner.Scan(row, &res)
	return res, err
}

//Insert inserts an array of data into table COLUMNS
func Insert(db *sql.DB, data []map[string]interface{}) (int64, error) {
	if nil == db {
		return nil, errors.New("sql.DB object couldn't be nil")
	}
	cond, vals, err := builder.BuildInsert("COLUMNS", data)
	if nil != err {
		return 0, err
	}
	result, err := db.Exec(cond, vals...)
	if nil != err || nil == result {
		return 0, err
	}
	return result.LastInsertId()
}

//Update updates the table COLUMNS
func Update(db *sql.DB, where, data map[string]interface{}) (int64, error) {
	if nil == db {
		return 0, errors.New("sql.DB object couldn't be nil")
	}
	cond, vals, err := builder.BuildUpdate("COLUMNS", where, data)
	if nil != err {
		return 0, err
	}
	result, err := db.Exec(cond, vals...)
	if nil != err {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete deletes matched records in COLUMNS
func Delete(db *sql.DB, where,data map[string]interface{}) (int64, error) {
	if nil == db {
		return 0, errors.New("sql.DB object couldn't be nil")
	}
	cond,vals,err := builder.BuildDelete("{{.TableName}}", where)
	if nil != err {
		return 0, err
	}
	result,err := db.Exec(cond, vals...)
	if nil != err {
		return 0, err
	}
	return result.RowsAffected()
}
```
