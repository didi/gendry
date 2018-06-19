# Scaner

### Scan

```go
import (
    "database/sql"
    "fmt"
    "github.com/didi/gendry/scanner"
)

type Person struct {
    Name string `ddb:"name"`
    Age int `ddb:"m_age"`
}


rows,_ := db.Query("select name,m_age from person")
var students []Person
err := scanner.Scan(rows, &students)
for _,student := range students {
	fmt.Println(student)
}
```

*Make sure the second param of Scan should be a reference*

### ScanClose
`ScanClose` is the same as the Scan but it also close the rows so you dont't need to worry about closing the rows yourself.

### ScanMap
ScanMap returns the result in the form of []map[string]interface{}, sometimes this could be more convenient.

```go
rows,_ := db.Query("select name,m_age from person")
result,err := scanner.ScanMap(rows)
for _,record := range result {
	fmt.Println(record["name"], record["m_age"])
}
```
If you don't want to define a struct,ScanMap may be useful.But the returned the map is `map[string]interface{}`, and `interface{}` is pretty unclear like the `void *` in `C` or `Object` in `JAVA`, it'll suck you sooner or later.

### ScanMapClose
ScanMapClose is the same as ScanMap but it also close the rows

### Map
`Map` convert a struct into a map which could easily be used to insert

Test cases blow may make sense

```go
package scaner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	type Person struct {
		Name string `ddb:"name"`
		Age  int    `ddb:"age"`
		foo  byte   `ddb:"foo"`
	}
	a := Person{"deen", 22, 1}
	b := &Person{"caibirdme", 23, 1}
	c := &b
	mapA, err := Map(a, DefaultTagName)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal("deen", mapA["name"])
	ass.Equal(22, mapA["age"])
	_, ok := mapA["foo"]
	ass.False(ok)
	mapB, err := Map(c, "")
	ass.NoError(err)
	ass.Equal("caibirdme", mapB["Name"])
	ass.Equal(23, mapB["Age"])
}
```
* Unexported fields will be ignored
* Ptr type will be ignored
* Resolve pointer automatically
* The second param specify what tagName you used in defining your struct.If passed an empty string, FieldName will be returned as the key of the map