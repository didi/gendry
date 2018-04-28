package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type person struct {
	Name string `ddb:"name" json:"name,omitempty"`
	Age  int    `ddb:"age" json:"age,asd,wdwas,string"`
	foo  byte   `ddb:"foo"`
	C    *int   `ddb:"cc"`
}

func TestMap(t *testing.T) {
	var q = 5
	p := &q
	a := person{"deen", 22, 1, p}

	mapA, err := Map(a, DefaultTagName)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal("deen", mapA["name"])
	ass.Equal(22, mapA["age"])
	var ok bool
	_, ok = mapA["foo"]
	ass.False(ok)
	_, ok = mapA["cc"]
	ass.False(ok)
}

func TestStructWithPointer(t *testing.T) {
	var q = 5
	p := &q
	b := &person{"caibirdme", 23, 1, p}
	c := &b
	mapB, err := Map(c, "")
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal("caibirdme", mapB["Name"])
	ass.Equal(23, mapB["Age"])
}

func TestStructWithMuiltiSubTag(t *testing.T) {
	var q = 5
	p := &q
	a := person{"deen", 22, 1, p}

	mapA, err := Map(a, "json")
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal("deen", mapA["name"])
	ass.Equal(22, mapA["age"])
	var ok bool
	_, ok = mapA["foo"]
	ass.False(ok)
	_, ok = mapA["cc"]
	ass.False(ok)
}

func TestNil(t *testing.T) {
	m, err := Map(nil, "")
	ass := assert.New(t)
	ass.Nil(m)
	ass.Nil(err)
}

func TestNonStructInput(t *testing.T) {
	ass := assert.New(t)
	m, err := Map(10, "")
	ass.Nil(m)
	ass.Equal(ErrNoneStructTarget, err)
}
