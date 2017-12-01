package schema

import (
	"fmt"
	"os"
)

const (
	errFormat = "Schema Error:[%s]\n"
)

func errUnknownType(columnName, columnType string) error {
	return schemaError(fmt.Sprintf("unknown datatype: columnName:%s, columnType:[%s]", columnName, columnType))
}

func schemaError(errmsg string) error {
	return fmt.Errorf(errFormat, errmsg)
}

func checkError(err error, exit bool) {
	if nil != err {
		fmt.Printf("%s\n", schemaError(err.Error()))
	}
	if exit {
		os.Exit(1)
	}
}
