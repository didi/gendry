package manager

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcatDSN_None(t *testing.T) {
	var setting []Setting
	ass := assert.New(t)
	ass.Equal("", concatDSN(setting), `should return "" when passed a nil slice`)
}

func TestConcatDSN_One(t *testing.T) {
	var setting []Setting
	ass := assert.New(t)
	setting = append(setting, SetAllowAllFiles(true))
	ass.Equal("allowAllFiles=true", concatDSN(setting), "trim rear &")
}

func TestConcatDSN_Multi(t *testing.T) {
	var setting []Setting
	ass := assert.New(t)
	setting = append(setting,
		SetAllowAllFiles(true),
		SetAutoCommit(true),
		SetCharset("utf8"),
		SetAllowNativePasswords(false),
		SetClientFoundRows(true),
		SetColumnsWithAlias(true),
		SetParseTime(false),
	)
	expected := "allowAllFiles=true&autocommit=true&charset=utf8&allowNativePasswords=false&clientFoundRows=true&columnsWithAlias=true&parseTime=false"
	ass.Equal(expected, concatDSN(setting), "multi dsn")
}

func TestConcatDSN_Time(t *testing.T) {
	var setting []Setting
	ass := assert.New(t)
	setting = append(setting, SetReadTimeout(time.Millisecond), SetTimeout(time.Second), SetWriteTimeout(time.Minute), SetReadTimeout(time.Hour))
	expected := "readTimeout=1ms&timeout=1s&writeTimeout=1m0s&readTimeout=1h0m0s"
	ass.Equal(expected, concatDSN(setting), "time unit")
}

func TestConcatDSN_Time_overflow(t *testing.T) {
	var setting []Setting
	ass := assert.New(t)
	setting = append(setting, SetReadTimeout(time.Microsecond), SetTimeout(24*time.Hour))
	ass.Equal("", concatDSN(setting), "duration <1ms or >=24h should be invalid")
}

func TestConcatDSN_String_null(t *testing.T) {
	var setting []Setting
	ass := assert.New(t)
	setting = append(setting, SetCollation(""), SetLoc(""), SetStrict(false), SetInterpolateParams(true))
	ass.Equal("strict=false&interpolateParams=true", concatDSN(setting), `null value should be ignored`)
}

func TestNew(t *testing.T) {
	dbName := "dbname"
	user := "user"
	password := "password"
	host := "localhost"
	o := New(dbName, user, password, host)
	ass := assert.New(t)
	ass.Equal(dbName, o.dbName)
	ass.Equal(user, o.user)
	ass.Equal(password, o.password)
	ass.Equal(host, o.host)
	o = o.Driver("oracle")
	ass.Equal("oracle", o.driver)
	o = o.Port(1234)
	ass.Equal(1234, o.port)
	_, err := o.Open(true)
	ass.Error(err)
}

func TestRealDSN_NoneSetting(t *testing.T) {
	dbName := "dbname"
	user := "user"
	password := "password"
	host := "localhost"
	o := New(dbName, user, password, host)
	ass := assert.New(t)
	ass.Equal(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, defaultPort, dbName), realDSN(o))
}

func TestRealDSN_WithSetting(t *testing.T) {
	dbName := "dbname"
	user := "user"
	password := "password"
	host := "localhost"
	o := New(dbName, user, password, host)
	o.Set(SetTimeout(time.Second), SetAllowCleartextPasswords(false))
	ass := assert.New(t)
	ass.Equal(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=1s&allowCleartextPasswords=false", user, password, host, defaultPort, dbName), realDSN(o))
}
