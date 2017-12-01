package builder

import (
	"context"
	"database/sql"
)

// AggregateQuery is a helper function to execute the aggregate query and return the result
func AggregateQuery(ctx context.Context, db *sql.DB, table string, where map[string]interface{}, aggregate AggregateSymbleBuilder) (ResultResolver, error) {
	cond, vals, err := BuildSelect(table, where, []string{aggregate.Symble()})
	if nil != err {
		return resultResolve{0}, err
	}
	rows, err := db.QueryContext(ctx, cond, vals...)
	if nil != err {
		return resultResolve{0}, err
	}
	var result interface{}
	for rows.Next() {
		err = rows.Scan(&result)
	}
	rows.Close()
	return resultResolve{result}, err
}

// ResultResolver is a helper for retrieving data
// caller should know the type and call the responding method
type ResultResolver interface {
	Int64() int64
	Float64() float64
}

type resultResolve struct {
	data interface{}
}

func (r resultResolve) Int64() int64 {
	i64, ok := r.data.(int64)
	if ok {
		return i64
	}
	i32, ok := r.data.(int32)
	if ok {
		return int64(i32)
	}
	i, ok := r.data.(int)
	if ok {
		return int64(i)
	}
	f64, ok := r.data.(float64)
	if ok {
		return int64(f64)
	}
	f32, _ := r.data.(float32)
	return int64(f32)
}

// from go-mysql-driver/mysql the value returned could be int64 float64 float32

func (r resultResolve) Float64() float64 {
	f64, ok := r.data.(float64)
	if ok {
		return f64
	}
	f32, ok := r.data.(float32)
	if ok {
		return float64(f32)
	}
	return float64(r.Int64())
}

// AggregateSymbleBuilder need to be implemented so that executor can
// get what should be put into `select Symble() from xxx where yyy`
type AggregateSymbleBuilder interface {
	Symble() string
}

type agBuilder string

func (a agBuilder) Symble() string {
	return string(a)
}

// AggregateCount count(col)
func AggregateCount(col string) AggregateSymbleBuilder {
	return agBuilder("count(" + col + ")")
}

// AggregateSum sum(col)
func AggregateSum(col string) AggregateSymbleBuilder {
	return agBuilder("sum(" + col + ")")
}

// AggregateAvg avg(col)
func AggregateAvg(col string) AggregateSymbleBuilder {
	return agBuilder("avg(" + col + ")")
}

// AggregateMax max(col)
func AggregateMax(col string) AggregateSymbleBuilder {
	return agBuilder("max(" + col + ")")
}

// AggregateMin min(col)
func AggregateMin(col string) AggregateSymbleBuilder {
	return agBuilder("min(" + col + ")")
}
