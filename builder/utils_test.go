package builder

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestResultResolver(t *testing.T) {
	var testData = []struct {
		origin   interface{}
		intout   int64
		floatout float64
	}{
		{
			origin:   10,
			intout:   10,
			floatout: 10.0,
		},
		{
			origin:   4.5,
			intout:   4,
			floatout: 4.5,
		},
	}
	ass := assert.New(t)
	for _, tc := range testData {
		rr := resultResolve{tc.origin}
		ass.Equal(tc.intout, rr.Int64())
		ass.True(math.Abs(tc.floatout-rr.Float64()) < 1e-5)
	}
}

func TestAggregateQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if nil != err {
		t.Error(err)
	}
	var testData = []struct {
		rows     *sqlmock.Rows
		reg      string
		ag       AggregateSymbleBuilder
		intout   int64
		floatout float64
		where    map[string]interface{}
	}{
		{
			sqlmock.NewRows([]string{"count(*)"}).AddRow(12),
			"SELECT count\\(\\*\\) FROM tb1",
			AggregateCount("*"),
			12,
			12,
			nil,
		},
		{
			sqlmock.NewRows([]string{"sum(age)"}).AddRow(math.MaxInt64),
			"SELECT sum\\(age\\) FROM tb1",
			AggregateSum("age"),
			math.MaxInt64,
			float64(math.MaxInt64),
			nil,
		},
		{
			sqlmock.NewRows([]string{"avg(age)"}).AddRow(100.957),
			"SELECT avg\\(age\\) FROM tb1",
			AggregateAvg("age"),
			100,
			100.957,
			nil,
		},
		{
			sqlmock.NewRows([]string{"sum(age)"}).AddRow(100.957),
			"sum\\(age\\)",
			AggregateSum("age"),
			100,
			100.957,
			nil,
		},
		{
			sqlmock.NewRows([]string{"max(age)"}).AddRow(math.Pi),
			"max\\(age\\)",
			AggregateMax("age"),
			3,
			math.Pi,
			nil,
		},
		{
			sqlmock.NewRows([]string{"min(age)"}).AddRow(math.Pi),
			"min\\(age\\) .+foo",
			AggregateMin("age"),
			3,
			math.Pi,
			map[string]interface{}{"foo": "hello", "bar": 123},
		},
	}
	ass := assert.New(t)
	ctx := context.Background()
	for _, tc := range testData {
		mock.ExpectQuery(tc.reg).WillReturnRows(tc.rows)
		result, err := AggregateQuery(ctx, db, "tb1", tc.where, tc.ag)
		ass.NoError(err)
		ass.NoError(mock.ExpectationsWereMet())
		ass.Equal(tc.intout, result.Int64())
		ass.True(math.Abs(result.Float64()-tc.floatout) < 1e6)
	}
}
