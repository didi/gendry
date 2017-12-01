package schema

import (
	"strings"
)

const (
	cTypeInt64   = "int64"
	cTypeInt     = "int"
	cTypeUInt    = "uint"
	cTypeString  = "string"
	cTypeFloat64 = "float64"
	cTypeTime    = "time.Time"
	cTypeInt8    = "int8"
	cTypeUInt64  = "uint64"
	cTypeByte    = "byte"
	cUnsigned    = "unsigned"
)

//Typer holds two methods allow user easilly get the information of a type
type typer interface {
	Type() string
	Match() bool
}

type typeWrapper func(string) typer

func i64TypeWrapper(s string) typer {
	s = strings.ToLower(s)
	u := uint64Type(s)
	if u.Match() {
		return u
	}
	return int64Type(s)
}

func byteTypeWrapper(s string) typer {
	s = strings.ToLower(s)
	b := byteType(s)
	if b.Match() {
		return b
	}
	return int8Type(s)
}

func intTypeWrapper(s string) typer {
	s = strings.ToLower(s)
	u := uintType(s)
	if u.Match() {
		return u
	}
	return intType(s)
}

func stringTypeWrapper(s string) typer {
	return stringType(strings.ToLower(s))
}

func float64TypeWrapper(s string) typer {
	return float64Type(strings.ToLower(s))
}

func timeTypeWrapper(s string) typer {
	return timeType(s)
}

type int64Type string

func (i64 int64Type) Type() string {
	return cTypeInt64
}

func (i64 int64Type) Match() bool {
	if strings.Contains(string(i64), "bigint") {
		return true
	}
	return false
}

type uint64Type string

func (ui64 uint64Type) Type() string {
	return cTypeUInt64
}

func (ui64 uint64Type) Match() bool {
	s := string(ui64)
	return strings.Contains(s, cUnsigned) && int64Type(s).Match()
}

type byteType string

func (b byteType) Type() string {
	return cTypeByte
}

func (b byteType) Match() bool {
	s := string(b)
	return strings.Contains(s, cUnsigned) && int8Type(s).Match()
}

type int8Type string

func (b int8Type) Type() string {
	return cTypeInt8
}

func (b int8Type) Match() bool {
	return strings.Contains(string(b), "tinyint")
}

type uintType string

func (ui uintType) Type() string {
	return cTypeUInt
}

func (ui uintType) Match() bool {
	s := string(ui)
	return strings.Contains(s, cUnsigned) && intType(s).Match()
}

type intType string

func (i intType) Type() string {
	return cTypeInt
}

func (i intType) Match() bool {
	return strings.Contains(string(i), "int")
}

type stringType string

func (s stringType) Type() string {
	return cTypeString
}

func (s stringType) Match() bool {
	var supportType = []string{"char", "text"}
	ss := string(s)
	for _, t := range supportType {
		if strings.Contains(ss, t) {
			return true
		}
	}
	return false
}

type float64Type string

func (f64 float64Type) Type() string {
	return cTypeFloat64
}

func (f64 float64Type) Match() bool {
	return strings.Contains(string(f64), "float") || strings.Contains(string(f64), "decimal")
}

type timeType string

func (t timeType) Type() string {
	return cTypeTime
}

func (t timeType) Match() bool {
	return t == "timestamp" || t == "date" || t == "datetime"
}
