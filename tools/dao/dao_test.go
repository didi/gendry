package dao

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDao(t *testing.T) {
	r, err := GenerateDao("someTable", "somestruct")
	ass := assert.New(t)
	ass.NoError(err)
	ass.NotNil(r)
	_, err = ioutil.ReadAll(r)
	ass.NoError(err)
}
