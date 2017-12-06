package schema

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"text/template"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
)

const (
	cDefaultTable = "COLUMNS"
	cTimeFormat   = "2006-01-02 15:04:05"
)

type columnSlice []column

func readTableStruct(db *sql.DB, tableName string, dbName string) (columnSlice, error) {
	var where = map[string]interface{}{
		"TABLE_NAME":   tableName,
		"TABLE_SCHEMA": dbName,
	}
	var selectFields = []string{"COLUMN_NAME", "COLUMN_TYPE", "COLUMN_COMMENT"}
	cond, vals, err := builder.BuildSelect(cDefaultTable, where, selectFields)
	if nil != err {
		return nil, err
	}
	rows, err := db.Query(cond, vals...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()
	var ts columnSlice
	scanner.SetTagName("json")
	err = scanner.Scan(rows, &ts)
	if nil != err {
		return nil, err
	}
	return ts, nil
}

func createStructSourceCode(cols columnSlice, tableName string) (io.Reader, string, error) {
	structName := convertUnderScoreToCammel(tableName)
	fillData := sourceCode{
		StructName: structName,
		TableName:  tableName,
		FieldList:  make([]sourceColumn, len(cols)),
	}
	for idx, col := range cols {
		colType, err := col.GetType()
		if nil != err {
			continue
		}
		fillData.FieldList[idx] = sourceColumn{
			Name:      col.GetName(),
			Type:      colType,
			StructTag: fmt.Sprintf("`json:\"%s\"`", col.Name),
		}
	}
	var buff bytes.Buffer
	err := template.Must(template.New("struct").Parse(codeTemplate)).Execute(&buff, fillData)
	if nil != err {
		return nil, "", err
	}
	return &buff, structName, nil
}

type sourceCode struct {
	StructName string
	TableName  string
	FieldList  []sourceColumn
}

type sourceColumn struct {
	Name      string
	Type      string
	StructTag string
}

const codeTemplate = `
// {{ .StructName }} is a mapping object for {{ .TableName }} table in mysql
type {{.StructName}} struct {
{{- range .FieldList }}
	{{ .Name }} {{ .Type }} {{ .StructTag }}
{{- end}}
}
`
