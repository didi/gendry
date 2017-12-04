## Gendry
[![Build Status](https://www.travis-ci.org/didi/Gendry.svg?branch=master)](https://www.travis-ci.org/didi/Gendry)
[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://github.com/didi/Gendry/blob/master/LICENSE)

**Gendry**是一个用于辅助操作数据库的Go包。基于`go-sql-driver/mysql`，它提供了一系列的方法来为你调用标准库`database/sql`中的方法准备参数。

**Gendery**主要分为3个独立的部分，你可以单独使用任何一个部分：

* [manager](#manager)
* [builder](#builder)
* [scanner](#scanner)
* [CLI Tool](#tools)

<h3 id="manager">Manager</h3>
manager主要用来初始化连接池(也就是`sql.DB`对象)，设置各种参数，因此叫manager。你可以设置任何`go-sql-driver/mysql`驱动支持的参数。
初始化连接池时，代码如下:

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
事实上，manager做的事情就是就是生成`dataSouceName`

dataSourceName的一般格式为：

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
```
manager是基于数据库驱动`go-mysql-driver/mysql`而开发的，manager支持了几乎所有该驱动支持的参数设置。具体用法看manager的README。

<h3 id="builder">Builder</h3>
builder顾名思义，就是构建生成sql语句。手写sql虽然直观简单，但是可维护性差，最主要的是硬编码容易出错。而且如果遇到大where in查询，而in的集合内容又是动态的，这就非常麻烦了。

builder不是一个ORM（我们开发Gendry的重要原因之一就是不喜欢ORM）,它只是提供简单的API帮你生成sql语句，如：

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

//cond = SELECT name,age,sex FROM g_xxx WHERE cityIN (?,?) AND score=? AND age>? GROUP BY department ORDER BY bonus DESC
//values = []interface{}{"beijing", "shanghai", 5, 35}

rows,err := db.Query(cond, values...)
```

同时，builder还提供一个便捷方法来进行聚合查询，比如：count,sum,max,min,avg

```go
where := map[string]interface{}{
    "score > ": 100,
    "city in": []interface{}{"Beijing", "Shijiazhuang",}
}
// AggregateSum,AggregateMax,AggregateMin,AggregateCount,AggregateAvg is supported
result, err := AggregateQuery(ctx, db, "tableName", where, AggregateSum("age"))
sumAge := result.Int64()

result,err = AggregateQuery(ctx, db, "tableName", where, AggregateCount("*")) 
numberOfRecords := result.Int64()

result,err = AggregateQuery(ctx, db, "tableName", where, AggregateAvg("score"))
averageScore := result.Float64()
```

连表查询和子查询通常都会是性能的瓶颈，这些sql大多需要手动优化，因此builder并**不支持**连表查询和子查询，具体文档看[builder](../builder/README.md)

<h3 id="scanner">Scanner</h3>
执行了数据库操作之后，要把返回的结果集和自定义的struct进行映射。Scanner提供一个简单的接口通过反射来进行结果集和自定义类型的绑定:

```go
type Person struct {
	Name string `ddb:"name"`
	Age int `ddb:"m_age"`
}

rows,err := db.Query("SELECT age as m_age,name from g_xxx where xxx")
defer rows.Close()

var students []Person

scanner.Scan(rows, &students)

for _,student := range students {
	fmt.Println(student)
}
```

scanner进行反射时会使用结构体的tag，如上所示，scanner会把结果集中的 m_age 绑定到结构体的Age域上。默认使用的tagName是`ddb:"xxx"`，你也可以自定义。

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

**scaner.SetTagName是全局设置，为了避免歧义，只允许设置一次，一般在初始化DB阶段进行此项设置**

### ScanMap
ScanMap方法返回的是一个map，有时候你可能不太像定义一个结构体去存你的中间结果，那么ScanMap或许比较有帮助

```go
rows,_ := db.Query("select name,m_age from person")
result,err := scanner.ScanMap(rows)
for _,record := range result {
	fmt.Println(record["name"], record["m_age"])
}
```

注意：

* 如果是使用Scan或者ScanMap的话，你必须在之后手动close rows
* 传给Scan的必须是引用
* ScanClose和ScanMapClose不需要手动close rows

<h3 id="tools">CLI Tool</h3>

除了以上API，Gendry还提供了一个命令行工具来进行代码生成，可以显著减少你的开发量

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

查看子命令的帮助

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

生成对应于某个表的结构体，本例中是information_schema库中的COLUMNS表

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

生成的结构体可以通过golint和govet的扫描

同时也可以根据某个表，生成对这个表进行增删改查的所有dao层代码

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

// COLUMNS is a mapping object for COLUMNS table in mysql
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

---

* 如果有任何问题，乐意为你解答
* 有任何功能上的需求，也欢迎提出来