### Builder

It's only a tool helping you build your queries.You should also use the `database/sql` to operate database

complex sql always need special optimization,which is hard to do it here.So, for very comlex sql, I suggest you write it manually, Exported WhereIn Helper will be added soon

### QuickStart

#### example_1

``` go
package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/didi/gendry/builder"
)

func main() {
    db,err := sql.Open("mysql", "xxxxxxxxxxx")
    if nil != err {
        panic(err)
    }
    where := map[string]interface{}{
        "_custom_0": builder.Custom("name=?", "name0"),
        "_custom_1": builder.JsonContains("my_json->'$.list'", 1),
        "country": "China",
        "role": "driver",
        "age >": 45,
        "gmt_create <": builder.Raw("gmt_modified"),
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
        "_groupby": "name",
        "_having": map[string]interface{}{
            "total >": 1000,
            "total <=": 50000,
        },
    	"_orderby": "age desc",
    }
    cond,vals,err := builder.BuildSelect("tableName", where, []string{"name", "count(price) as total", "age"})
    
    //cond: SELECT name,count(price) as total,age FROM tableName WHERE (name=? AND (? MEMBER OF(my_json->'$.list')) AND ((x1=? AND x2>=?) OR (x3=? AND x4!=?)) AND country=? AND role=? AND age>? AND gmt_create<gmt_modified) GROUP BY name HAVING (total>? AND total<=?) ORDER BY age desc
    //vals: []interface{}{"name0", 1, 11, 45, "234", "tx2", "China", "driver", 45, 1000, 50000}
    
	notDelWhere := map[string]interface{}{
		"is_del": 0,
	}

	bllWhere := map[string]interface{}{
		"_custom_0":    builder.Custom("name=?", "name0"),
		"_custom_1":    builder.JsonContains("my_json->'$.list'", 1),
		"country":      "China",
		"role":         "driver",
		"age >":        45,
		"gmt_create <": builder.Raw("gmt_modified"),
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
	}

	aopWhere := map[string]interface{}{
		"_and": []map[string]interface{}{
			notDelWhere,
			bllWhere,
		},
		"_groupby": "name",
		"_having": map[string]interface{}{
			"total >":  1000,
			"total <=": 50000,
		},
		"_orderby": "age desc",
	}
    
    condAop,valsAop,err := builder.BuildSelect("tableName", where, []string{"name", "count(price) as total", "age"})
    
    //condAop:  SELECT name,count(price) as total,age FROM tableName WHERE (((is_del=?) AND (name=? AND (? MEMBER OF(my_json->'$.list')) AND ((x1=? AND x2>=?) OR (x3=? AND x4!=?)) AND country=? AND role=? AND age>? AND gmt_create<gmt_modified))) GROUP BY name HAVING (total>? AND total<=?) ORDER BY age desc
    //valsAop:  [0 name0 1 11 45 234 tx2 China driver 45 1000 50000]

	if nil != err {
		panic(err)
	}	

    rows,err := db.Query(cond, vals...)
    if nil != err {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        var id int
        var name,phone string
        rows.Scan(&id, &name, &phone)
        fmt.Println(id, name, phone)
    }

    //have fun !!
}
```

---

## API

#### `BuildSelect`

sign: `BuildSelect(table string, where map[string]interface{}, field []string) (string,[]interface{},error)`

operators supported(case-insensitive):

* =
* &gt;
* &lt;
* =
* &lt;=
* &gt;=
* !=
* &lt;&gt;
* in
* not in
* like
* not like
* between
* not between

``` go
where := map[string]interface{}{
    "foo <>": "aha",
    "bar <=": 45,
    "sex in": []interface{}{"girl", "boy"},
    "name like": "%James",
}
```

others supported:

* _or
* _orderby
* _groupby
* _having
* _limit
* _lockMode
* _custom_xxx
* _and

``` go
where := map[string]interface{}{
    "age >": 100,
    "_custom_1":    builder.JsonContains("my_json->'$.list'", 1),
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
    "_orderby": "fieldName asc",
    "_groupby": "fieldName",
    "_having": map[string]interface{}{"foo":"bar",},
    "_limit": []uint{offset, row_count},
    "_lockMode": "share",
}

aopWhere := map[string]interface{}{
  "_and_soft_deleted": []map[string]interface{}{
    {
        "is_del": 1,
    },
    where,
  }
}
```
Note:
* _having will be ignored if _groupby isn't setted
* value of _limit could be:
    * `"_limit": []uint{a,b}` => `LIMIT a,b`
    * `"_limit": []uint{a}` => `LIMIT 0,a`
* value of _lockMode only supports `share` and `exclusive` temporarily:
    * `share` representative `SELECT ... LOCK IN SHARE MODE`. Unfortunately, the current version does not support `SELECT ... FOR SHARE`, It'll be supported in the future.
    * `exclusive` representative `SELECT ... FOR UPDATE`
* if key starts with `_custom_`, the corresponding value must be a `builder.Comparable`. We provide builtin type such as `Custom` and `JsonContains`. You can also provide your own implementation if you want
* `JsonSet`,`JsonArrayAppend`,`JsonArrayInsert`,`JsonRemove` should be used in update map rather than where map

#### Aggregate

sign: `AggregateQuery(ctx context.Context, db *sql.DB, table string, where map[string]interface{}, aggregate AggregateSymbleBuilder) (ResultResolver, error)`

Aggregate is a helper function to help executing some aggregate queries such as:
* sum
* avg
* max
* min
* count

example:
```go
where := map[string]interface{}{
    "score > ": 100,
    "city in": []interface{}{"Beijing", "Shijiazhuang",}
}
// supported: AggregateSum,AggregateMax,AggregateMin,AggregateCount,AggregateAvg
result, err := AggregateQuery(ctx, db, "tableName", where, AggregateSum("age"))
sumAge := result.Int64()

result,err = AggregateQuery(ctx, db, "tableName", where, AggregateCount("*")) 
numberOfRecords := result.Int64()

result,err = AggregateQuery(ctx, db, "tableName", where, AggregateAvg("score"))
averageScore := result.Float64()
```

#### `BuildUpdate`

sign: `BuildUpdate(table string, where map[string]interface{}, update map[string]interface{}) (string, []interface{}, error)`

BuildUpdate is very likely to BuildSelect but it **doesn't support**:

* _orderby
* _groupby
* _having

``` go
where := map[string]interface{}{
    "foo <>": "aha",
    "bar <=": 45,
    "sex in": []interface{}{"girl", "boy"},
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
    "_limit": uint(10),
}
update := map[string]interface{}{
	"role": "primaryschoolstudent",
	"rank": 5,
	"_custom_0": qb.JsonArrayAppend("my_json", "$", 0, "$", 1),
}
cond,vals,err := qb.BuildUpdate("table_name", where, update)

db.Exec(cond, vals...)
```

#### `BuildInsert`

sign: `BuildInsert(table string, data []map[string]interface{}) (string, []interface{}, error)`

data is a slice and every element(map) in it must have the same keys:

``` go
var data []map[string]interface{}
data = append(data, map[string]interface{}{
    "name": "deen",
    "age":  23,
})
data = append(data, map[string]interface{}{
    "name": "Tony",
    "age":  30,
})
cond, vals, err := qb.BuildInsert(table, data)
db.Exec(cond, vals...)
```

#### `BuildInsertIgnore`

sign: `BuildInsertIgnore(table string, data []map[string]interface{}) (string, []interface{}, error)`

data is a slice and every element(map) in it must have the same keys:

``` go
var data []map[string]interface{}
data = append(data, map[string]interface{}{
    "name": "deen",
    "age":  23,
})
data = append(data, map[string]interface{}{
    "name": "Tony",
    "age":  30,
})
cond, vals, err := qb.BuildInsertIgnore(table, data)
db.Exec(cond, vals...)
```

#### `BuildReplaceInsert`

sign: `BuildReplaceInsert(table string, data []map[string]interface{}) (string, []interface{}, error)`

data is a slice and every element(map) in it must have the same keys:

``` go
var data []map[string]interface{}
data = append(data, map[string]interface{}{
    "name": "deen",
    "age":  23,
})
data = append(data, map[string]interface{}{
    "name": "Tony",
    "age":  30,
})
cond, vals, err := qb.BuildReplaceInsert(table, data)
db.Exec(cond, vals...)
```

#### `BuildInsertOnDuplicate`

sign: `BuildInsertOnDuplicate(table string, data []map[string]interface{}, update map[string]interface{}) (string, []interface{}, error)`

data is a slice and every element(map) in it must have the same keys:

``` go
data := []map[string]interface{}{
    {
        "name": "deen",
        "age":  23,
    },
    {
        "name": "Tony",
        "age":  30,
    },
}
update := map[string]interface{}{
    "role": "primaryschoolstudent",
    "rank": 5,
}
cond, vals, err := qb.BuildInsertOnDuplicate(table, data, update)
db.Exec(cond, vals...)


// update support builder.Raw to update when duplicate with value in insert data
update = map[string]interface{}{
    "code": builder.Raw("VALUES(code)"), // mysql 8.x  builder.Raw("new.code")
    "name": builder.Raw("VALUES(name)"), // mysql 8.x  builder.Raw("new.name")
}
cond, values, err := builder.BuildInsertOnDuplicate(table, data, update)
// INSERT INTO country (id, code, name) VALUES (?,?,?),(?,?,?),(?,?,?) 
// ON DUPLICATE KEY UPDATE code=VALUES(code),name=VALUES(name)
```

#### `NamedQuery`

sign: `func NamedQuery(sql string, data map[string]interface{}) (string, []interface{}, error)`

For very complex query, this might be helpful. And for critical system, this is recommended.


```go
cond, vals, err := builder.NamedQuery("select * from tb where name={{name}} and id in (select uid from anothertable where score in {{m_score}})", map[string]interface{}{
	"name": "caibirdme",
	"m_score": []float64{3.0, 5.8, 7.9},
})

assert.Equal("select * from tb where name=? and id in (select uid from anothertable where score in (?,?,?))", cond)
assert.Equal([]interface{}{"caibirdme", 3.0, 5.8, 7.9}, vals)
```

#### `BuildDelete`

sign: `BuildDelete(table string, where map[string]interface{}) (string, []interface{}, error)`

------

## Safety
If you use `Prepare && stmt.SomeMethods` then You have no need to worry about the safety.
Prepare is a safety mechanism backed by `mysql`, it makes sql injection out of work.

So `builder` **doesn't** escape the string values it received -- it's unnecessary

If you call `db.Query(cond, vals...)` directly, and you **don't** set `interpolateParams` which is one of the driver's variables to `true`, the driver actually will still prepare a stmt.So it's safe.

Remember:
* don't assemble raw sql yourself,use `builder` instead.
* don't set `interpolateParams` to `true`(default false) if you're not aware of the consequence.

Obey instructions above there's no safety issues for most cases.
