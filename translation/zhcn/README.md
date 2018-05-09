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

对于比较复杂的查询, `NamedQuery`将会派上用场:
```go
cond, vals, err := builder.NamedQuery("select * from tb where name={{name}} and id in (select uid from anothertable where score in {{m_score}})", map[string]interface{}{
	"name": "caibirdme",
	"m_score": []float64{3.0, 5.8, 7.9},
})

assert.Equal("select * from tb where name=? and id in (select uid from anothertable where score in (?,?,?))", cond)
assert.Equal([]interface{}{"caibirdme", 3.0, 5.8, 7.9}, vals)
```
slice类型的值会根据slice的长度自动展开  
这种方式基本上就是手写sql，非常便于DBA review同时也方便开发者进行复杂sql的调优  
**对于关键系统，推荐使用这种方式**

具体文档看[builder](../../builder/README.md)

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

除了以上API，Gendry还提供了一个命令行工具来进行代码生成，可以显著减少你的开发量。详见[gforge](https://github.com/caibirdme/gforge)

---

* 如果有任何问题，乐意为你解答
* 有任何功能上的需求，也欢迎提出来
