## Manager

Manager is used for initializing a database connection pool(sql.DB in go's standard library) 

The original way to initialize a sql.DB is something like：

```go
pool,err := sql.Open("mysql", dataSourceName)
```
And the format of `dataSourceName` is：

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
```

What the manager does is providing a series of simple methods which help you setting these parameters.

### Example
```go

import (
	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error
db,err = manager.New(dbName, user, password, host)
		.Set(
			SetCharset("utf8"),
			SetParseTime(true),
			SetLocal("UTC")
		)
		.Port(3307)
		.Open(true)
```

> `godoc` is a great tool for scanning the API documentation

### Basic API

`func New(dbName, user, password, host string) *Option`

New returns an Option

`func (o *Option) Driver(driver string) *Option`

Driver sets the driver, default mysql

`func (o *Option) Open(ping bool) (*sql.DB, error)`

Open is used for creating a *sql.DB.If ping=true,it will exec the db.Ping() after creating the db object

`func (o *Option) Port(port int) *Option`

Port sets the server port,default 3306

`func (o *Option) Set(sets ...Setting) *Option`

Set recieves a series of Set*-like functions

---

### Setting APIs
For more details see [DSN-Data-Source-Name](https://github.com/go-sql-driver/mysql#dsn-data-source-name)

* `SetAllowAllFiles`: allowAllFiles=true disables the file Whitelist for LOAD DATA LOCAL INFILE and allows all files. Might be insecure!
* `SetAllowCleartextPasswords`: allowCleartextPasswords=true allows using thecleartext client side plugin if required by an account, such as one defined with the PAM authentication plugin. Sending passwords in clear text may be a security problem in some configurations.
* `SetAllowNativePasswords`: Allows the usage of the mysql native password method
* `SetAutoCommit`: Set it to true if you know what you are doing
* `SetCharset`: Sets the charset used for client-server interaction
* `SetClientFoundRows`: clientFoundRows=true causes an UPDATE to return the number of matching rows instead of the number of rows changed.
* `SetCollation`: Sets the collation used for client-server interaction on connection. In contrast to charset, collation does not issue additional queries. If the specified collation is unavailable on the target server,the connection will fail.
* `SetColumnsWithAlias`: When columnsWithAlias is true, calls to sql.Rows.Columns() will return the table alias and the column name separated by a dot.
* `SetInterpolateParams`: If interpolateParams is true, placeholders (?) in calls to db.Query() and db.Exec() are interpolated into a single query string with given parameters. This reduces the number of roundtrips, since the driver has to prepare a statement, execute it with given parameters and close the statement again with interpolateParams=false.
* `SetLoc`: SetLoc Sets the location for time.Time values (when using parseTime=true). "Local" sets the system's location. See time.LoadLocation in standard library for detail.
* `SetParseTime`: SetParseTime if set true changes the DATE and DATETIME type in mysql will be stored in a time.Time object rather than the []byte / string
* `SetReadTimeout`: SetReadTimeout I/O read timeout. timeout ∈ [1ms, 24h) 
* `SetStrict`: SetStrict strict=true enables the strict mode in which MySQL warnings are treated as errors.
* `SetTimeout`: SetTimeout Driver side connection timeout. timeout ∈ [1ms, 24h)
* `SetWriteTimeout`: SetWriteTimeout I/O write timeout. timeout ∈ [1ms, 24h) 