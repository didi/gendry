package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHaving(t *testing.T) {
	type inStruct struct {
		table       string
		where       map[string]interface{}
		selectField []string
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"age > ": 23,
				},
				selectField: []string{"count(*) as total"},
			},
			out: outStruct{
				cond: "SELECT count(*) as total FROM tb WHERE (age>?)",
				vals: []interface{}{23},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"age > ":   23,
					"_groupby": "name",
					"_having": map[string]interface{}{
						"total >=": 1000,
						"total <":  50000,
					},
				},
				selectField: []string{"name, count(price) as total"},
			},
			out: outStruct{
				cond: "SELECT name, count(price) as total FROM tb WHERE (age>?) GROUP BY name HAVING (total>=? AND total<?)",
				vals: []interface{}{23, 1000, 50000},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_groupby": "name",
					"_having": map[string]interface{}{
						"total >=": 1000,
						"total <":  50000,
					},
				},
				selectField: []string{"name, count(price) as total"},
			},
			out: outStruct{
				cond: "SELECT name, count(price) as total FROM tb GROUP BY name HAVING (total>=? AND total<?)",
				vals: []interface{}{1000, 50000},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_having": map[string]interface{}{
						"total >=": 1000,
						"total <":  50000,
					},
					"age in": []interface{}{1, 2, 3},
				},
				selectField: []string{"name, age"},
			},
			out: outStruct{
				cond: "SELECT name, age FROM tb WHERE (age IN (?,?,?))",
				vals: []interface{}{1, 2, 3},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := BuildSelect(tc.in.table, tc.in.where, tc.in.selectField)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}
}

func Test_BuildInsert(t *testing.T) {
	ass := assert.New(t)
	type inStruct struct {
		table   string
		setData []map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				setData: []map[string]interface{}{
					map[string]interface{}{
						"foo": "bar",
						"age": 23,
					},
				},
			},
			out: outStruct{
				cond: "INSERT INTO tb (age,foo) VALUES (?,?)",
				vals: []interface{}{23, "bar"},
				err:  nil,
			},
		},
	}
	for _, tc := range data {
		cond, vals, err := BuildInsert(tc.in.table, tc.in.setData)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}
}

func Test_BuildDelete(t *testing.T) {
	ass := assert.New(t)
	type inStruct struct {
		table string
		where map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"age >=":   21,
					"sex in":   []interface{}{"male", "female"},
					"hobby in": []interface{}{"soccer", "basketball", "tenis"},
				},
			},
			out: outStruct{
				cond: "DELETE FROM tb WHERE (hobby IN (?,?,?) AND sex IN (?,?) AND age>=?)",
				vals: []interface{}{"soccer", "basketball", "tenis", "male", "female", 21},
				err:  nil,
			},
		},
	}
	for _, tc := range data {
		cond, vals, err := BuildDelete(tc.in.table, tc.in.where)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}
}

func Test_BuildUpdate(t *testing.T) {
	type inStruct struct {
		table   string
		where   map[string]interface{}
		setData map[string]interface{}
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":    "bar",
					"age >=": 23,
					"sex in": []interface{}{"male", "female"},
				},
				setData: map[string]interface{}{
					"score":    50,
					"district": "010",
				},
			},
			out: outStruct{
				cond: "UPDATE tb SET district=?,score=? WHERE (foo=? AND sex IN (?,?) AND age>=?)",
				vals: []interface{}{"010", 50, "bar", "male", "female", 23},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := BuildUpdate(tc.in.table, tc.in.where, tc.in.setData)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}
}

func Test_BuildSelect(t *testing.T) {
	type inStruct struct {
		table  string
		where  map[string]interface{}
		fields []string
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":      "bar",
					"qq":       "tt",
					"age in":   []interface{}{1, 3, 5, 7, 9},
					"faith <>": "Muslim",
					"_orderby": "age desc",
					"_groupby": "department",
					"_limit":   []uint{0, 100},
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=? AND qq=? AND age IN (?,?,?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC LIMIT 0,100",
				vals: []interface{}{"bar", "tt", 1, 3, 5, 7, 9, "Muslim"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"name like": "%123",
				},
				fields: nil,
			},
			out: outStruct{
				cond: `SELECT * FROM tb WHERE (name LIKE ?)`,
				vals: []interface{}{"%123"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"name": "caibirdme",
				},
				fields: nil,
			},
			out: outStruct{
				cond: "SELECT * FROM tb WHERE (name=?)",
				vals: []interface{}{"caibirdme"},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := BuildSelect(tc.in.table, tc.in.where, tc.in.fields)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}
}

func BenchmarkBuildSelect_Sequelization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildSelect("tb", map[string]interface{}{
			"foo":      "bar",
			"qq":       "tt",
			"age in":   []interface{}{1, 3, 5, 7, 9},
			"faith <>": "Muslim",
			"_orderby": "age desc",
			"_groupby": "department",
			"_limit":   []uint{0, 100},
		}, []string{"a", "b", "c"})
	}
}

func BenchmarkBuildSelect_Parallel(b *testing.B) {
	expectCond := "SELECT * FROM tb WHERE (foo=? AND qq=? AND age IN (?,?,?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC LIMIT 0,100"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cond, _, _ := BuildSelect("tb", map[string]interface{}{
				"foo":      "bar",
				"qq":       "tt",
				"age in":   []interface{}{1, 3, 5, 7, 9},
				"faith <>": "Muslim",
				"_orderby": "age desc",
				"_groupby": "department",
				"_limit":   []uint{0, 100},
			}, nil)
			if cond != expectCond {
				b.Errorf("should be %s but %s\n", expectCond, cond)
			}
		}
	})
}

func TestLike(t *testing.T) {
	type inStruct struct {
		table  string
		where  map[string]interface{}
		fields []string
	}
	type outStruct struct {
		cond string
		vals []interface{}
		err  error
	}
	var data = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"bar like": "haha%",
					"baz like": "%some",
					"foo":      1,
				},
				fields: nil,
			},
			out: outStruct{
				cond: `SELECT * FROM tb WHERE (foo=? AND bar LIKE ? AND baz LIKE ?)`,
				vals: []interface{}{1, "haha%", "%some"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"bar like": "haha%",
					"baz like": "%some",
					"foo":      1,
					"age in":   []interface{}{1, 3, 5, 7, 9},
				},
				fields: nil,
			},
			out: outStruct{
				cond: `SELECT * FROM tb WHERE (foo=? AND age IN (?,?,?,?,?) AND bar LIKE ? AND baz LIKE ?)`,
				vals: []interface{}{1, 1, 3, 5, 7, 9, "haha%", "%some"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"name like": "%James",
				},
				fields: []string{"name"},
			},
			out: outStruct{
				cond: `SELECT name FROM tb WHERE (name LIKE ?)`,
				vals: []interface{}{"%James"},
				err:  nil,
			},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := BuildSelect(tc.in.table, tc.in.where, tc.in.fields)
		ass.Equal(tc.out.err, err)
		ass.Equal(tc.out.cond, cond)
		ass.Equal(tc.out.vals, vals)
	}
}
