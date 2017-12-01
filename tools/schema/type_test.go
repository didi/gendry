package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	var dataSet = []struct {
		t      string
		expect string
	}{
		{"int(11)", cTypeInt},
		{"int(11) unsigned", cTypeUInt},
		{"smallint", cTypeInt},
		{"smallint unsigned", cTypeUInt},
		{"bigint", cTypeInt64},
		{"tinyint(2) unsigned", cTypeByte},
		{"float", cTypeFloat64},
		{"decimal(8,2)", cTypeFloat64},
		{"tinyint(2)", cTypeInt8},
		{"bigint unsigned", cTypeUInt64},
		{"bigint(20) unsigned", cTypeUInt64},
		{"timestamp", cTypeTime},
		{"date", cTypeTime},
		{"datetime", cTypeTime},
	}
	ass := assert.New(t)
	for _, tc := range dataSet {
		ass.Equal(tc.expect, getType(tc.t), "type: %s", tc.t)
	}
}
