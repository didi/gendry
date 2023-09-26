package builder

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLockMode(t *testing.T) {
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
					"vx":       []interface{}{1, 3, 5},
					"faith <>": "Muslim",
					"_or": []map[string]interface{}{
						{
							"aa": 11,
							"bb": "xswl",
						},
						{
							"cc":    "234",
							"dd in": []interface{}{7, 8},
							"_or": []map[string]interface{}{
								{
									"neeest_ee <>": "dw42",
									"neeest_ff in": []interface{}{34, 59},
								},
								{
									"neeest_gg":        1259,
									"neeest_hh not in": []interface{}{358, 1245},
								},
							},
						},
					},
					"_orderby":  "age DESC,score ASC",
					"_groupby":  "department",
					"_limit":    []uint{0, 100},
					"_lockMode": "share",
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (((aa=? AND bb=?) OR (((neeest_ff IN (?,?) AND neeest_ee!=?) OR (neeest_gg=? AND neeest_hh NOT IN (?,?))) AND cc=? AND dd IN (?,?))) AND foo=? AND qq=? AND age IN (?,?,?,?,?) AND vx IN (?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC,score ASC LIMIT ?,? LOCK IN SHARE MODE",
				vals: []interface{}{11, "xswl", 34, 59, "dw42", 1259, 358, 1245, "234", 7, 8, "bar", "tt", 1, 3, 5, 7, 9, 1, 3, 5, "Muslim", 0, 100},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"name like": "%123",
					"_lockMode": "exclusive",
				},
				fields: nil,
			},
			out: outStruct{
				cond: `SELECT * FROM tb WHERE (name LIKE ?) FOR UPDATE`,
				vals: []interface{}{"%123"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"name":      "caibirdme",
					"_lockMode": "share",
				},
				fields: nil,
			},
			out: outStruct{
				cond: "SELECT * FROM tb WHERE (name=?) LOCK IN SHARE MODE",
				vals: []interface{}{"caibirdme"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":       "bar",
					"_orderby":  "  ",
					"_lockMode": "exclusive",
				},
				fields: nil,
			},
			out: outStruct{
				cond: "SELECT * FROM tb WHERE (foo=?) FOR UPDATE",
				vals: []interface{}{"bar"},
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
						"vx":       []interface{}{1, 3, 5},
					},
				},
				selectField: []string{"name, count(price) as total"},
			},
			out: outStruct{
				cond: "SELECT name, count(price) as total FROM tb WHERE (age>?) GROUP BY name HAVING (vx IN (?,?,?) AND total>=? AND total<?)",
				vals: []interface{}{23, 1, 3, 5, 1000, 50000},
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
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_limit": []uint{1},
					"age in": []interface{}{1, 2, 3},
				},
				selectField: []string{"name, age"},
			},
			out: outStruct{
				cond: "SELECT name, age FROM tb WHERE (age IN (?,?,?)) LIMIT ?,?",
				vals: []interface{}{1, 2, 3, 0, 1},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_limit": []uint{2, 1},
					"age in": []interface{}{1, 2, 3},
				},
				selectField: []string{"name, age"},
			},
			out: outStruct{
				cond: "SELECT name, age FROM tb WHERE (age IN (?,?,?)) LIMIT ?,?",
				vals: []interface{}{1, 2, 3, 2, 1},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_groupby": "  ",
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

func TestBuildHaving_1(t *testing.T) {
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
	var testCases = []struct {
		in  inStruct
		out outStruct
	}{
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_groupby": "name",
					"_having": map[string]interface{}{
						"total IN":     []interface{}{1000, 2000, 3000},
						"total NOT IN": []interface{}{2000},
					},
				},
				selectField: []string{"name", "COUNT(price) AS total"},
			},
			out: outStruct{
				cond: "SELECT name,COUNT(price) AS total FROM tb GROUP BY name HAVING (total IN (?,?,?) AND total NOT IN (?))",
				vals: []interface{}{1000, 2000, 3000, 2000},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"_groupby": "name",
					"_having": map[string]interface{}{
						"total BETWEEN ":        []interface{}{1000, 50000},
						"total NOT   BETWEEN  ": []interface{}{3000, 3500},
					},
				},
				selectField: []string{"name", "COUNT(price) AS total"},
			},
			out: outStruct{
				cond: "SELECT name,COUNT(price) AS total FROM tb GROUP BY name HAVING ((total BETWEEN ? AND ?) AND (total NOT BETWEEN ? AND ?))",
				vals: []interface{}{1000, 50000, 3000, 3500},
				err:  nil,
			},
		},
	}

	ass := assert.New(t)
	for _, tc := range testCases {
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
					{
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
					"_limit":   10,
				},
			},
			out: outStruct{
				cond: "DELETE FROM tb WHERE (hobby IN (?,?,?) AND sex IN (?,?) AND age>=?) LIMIT ?",
				vals: []interface{}{"soccer", "basketball", "tenis", "male", "female", 21, 10},
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
					"_or": []map[string]interface{}{
						{
							"x1":    11,
							"x2 >=": 45,
						},
						{
							"x3":    "234",
							"x4 <>": "tx2",
						},
					},
				},
				setData: map[string]interface{}{
					"score":    50,
					"district": "010",
				},
			},
			out: outStruct{
				cond: "UPDATE tb SET district=?,score=? WHERE (((x1=? AND x2>=?) OR (x3=? AND x4!=?)) AND foo=? AND sex IN (?,?) AND age>=?)",
				vals: []interface{}{"010", 50, 11, 45, "234", "tx2", "bar", "male", "female", 23},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":    "bar",
					"age >=": 23,
					"sex in": []interface{}{"male", "female"},
					"_limit": 10,
				},
				setData: map[string]interface{}{
					"score":    50,
					"district": "010",
				},
			},
			out: outStruct{
				cond: "UPDATE tb SET district=?,score=? WHERE (foo=? AND sex IN (?,?) AND age>=?) LIMIT ?",
				vals: []interface{}{"010", 50, "bar", "male", "female", 23, 10},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":    "bar",
					"age >=": 23,
					"sex in": []interface{}{"male", "female"},
					"_limit": 5.5,
				},
				setData: map[string]interface{}{
					"score":    50,
					"district": "010",
				},
			},
			out: outStruct{
				cond: "",
				vals: nil,
				err:  errLimitType,
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
					"vx":       []interface{}{1, 3, 5},
					"faith <>": "Muslim",
					"_or": []map[string]interface{}{
						{
							"aa": 11,
							"bb": "xswl",
						},
						{
							"cc":    "234",
							"dd in": []interface{}{7, 8},
							"_or": []map[string]interface{}{
								{
									"neeest_ee <>": "dw42",
									"neeest_ff in": []interface{}{34, 59},
								},
								{
									"neeest_gg":        1259,
									"neeest_hh not in": []interface{}{358, 1245},
								},
							},
						},
					},
					"_orderby": "age DESC,score ASC",
					"_groupby": "department",
					"_limit":   []uint{0, 100},
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (((aa=? AND bb=?) OR (((neeest_ff IN (?,?) AND neeest_ee!=?) OR (neeest_gg=? AND neeest_hh NOT IN (?,?))) AND cc=? AND dd IN (?,?))) AND foo=? AND qq=? AND age IN (?,?,?,?,?) AND vx IN (?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC,score ASC LIMIT ?,?",
				vals: []interface{}{11, "xswl", 34, 59, "dw42", 1259, 358, 1245, "234", 7, 8, "bar", "tt", 1, 3, 5, 7, 9, 1, 3, 5, "Muslim", 0, 100},
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
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":      "bar",
					"_orderby": "  ",
				},
				fields: nil,
			},
			out: outStruct{
				cond: "SELECT * FROM tb WHERE (foo=?)",
				vals: []interface{}{"bar"},
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

func Test_BuildSelectMultiOr(t *testing.T) {
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
					"a": 1,
					"_or": []map[string]interface{}{
						{
							"b": 2,
							"c": 3,
						},
						{
							"d": 4,
							"e": 5,
						},
					},
					"_or2": []map[string]interface{}{
						{
							"b2": 22,
							"c2": 33,
						},
						{
							"d2": 44,
							"e2": 55,
						},
					},
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (((b=? AND c=?) OR (d=? AND e=?)) AND ((b2=? AND c2=?) OR (d2=? AND e2=?)) AND a=?)",
				vals: []interface{}{2, 3, 4, 5, 22, 33, 44, 55, 1},
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
		_, _, err := BuildSelect("tb", map[string]interface{}{
			"foo":      "bar",
			"qq":       "tt",
			"age in":   []interface{}{1, 3, 5, 7, 9},
			"faith <>": "Muslim",
			"_orderby": "age DESC",
			"_groupby": "department",
			"_limit":   []uint{0, 100},
		}, []string{"a", "b", "c"})
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkBuildSelect_Parallel(b *testing.B) {
	expectCond := "SELECT * FROM tb WHERE (foo=? AND qq=? AND age IN (?,?,?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC LIMIT ?,?"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cond, _, _ := BuildSelect("tb", map[string]interface{}{
				"foo":      "bar",
				"qq":       "tt",
				"age in":   []interface{}{1, 3, 5, 7, 9},
				"faith <>": "Muslim",
				"_orderby": "age DESC",
				"_groupby": "department",
				"_limit":   []uint{0, 100},
			}, nil)
			if cond != expectCond {
				b.Fatalf("should be %s but %s\n", expectCond, cond)
			}
		}
	})
}

func TestNamedQuery(t *testing.T) {
	var testData = []struct {
		sql  string
		data map[string]interface{}
		cond string
		vals []interface{}
		err  error
	}{
		{
			sql: `select * from tb where name={{name}}`,
			data: map[string]interface{}{
				"age": 24,
			},
			cond: "",
			vals: nil,
			err:  errors.New("name not found"),
		},
		{
			sql:  `select * from tb where name=hello`,
			data: nil,
			cond: "select * from tb where name=hello",
			vals: nil,
			err:  nil,
		},
		{
			sql: `select * from tb where name={{name}} and age<{{age}}`,
			data: map[string]interface{}{
				"age": 24,
			},
			cond: "",
			vals: nil,
			err:  errors.New("name not found"),
		},
		{
			sql: `select * from tb where name={{name}} and age<>{{age}}`,
			data: map[string]interface{}{
				"name": "caibirdme",
				"age":  24,
			},
			cond: `select * from tb where name=? and age<>?`,
			vals: []interface{}{"caibirdme", 24},
			err:  nil,
		},
		{
			sql: `select * from tb where name={{name}} and age in {{age}}`,
			data: map[string]interface{}{
				"name": "caibirdme",
				"age":  []int{1, 2, 3},
			},
			cond: `select * from tb where name=? and age in (?,?,?)`,
			vals: []interface{}{"caibirdme", 1, 2, 3},
			err:  nil,
		},
		{
			sql: `select * from tb where name={{name}} and age in (select m_age from anothertb where m_age>{{m_age}})`,
			data: map[string]interface{}{
				"name":  "caibirdme",
				"m_age": 88.9,
			},
			cond: `select * from tb where name=? and age in (select m_age from anothertb where m_age>?)`,
			vals: []interface{}{"caibirdme", 88.9},
			err:  nil,
		},
		{
			sql: `select * from tb where age in {{some}} and other in {{some}}`,
			data: map[string]interface{}{
				"some": []float64{24.0, 28.7},
			},
			cond: "select * from tb where age in (?,?) and other in (?,?)",
			vals: []interface{}{24.0, 28.7, 24.0, 28.7},
			err:  nil,
		},
		{
			sql: `select a.name,a.age from tb1 as a join tb2 as b on a.id=b.id where a.age>{{age}} and b.age<{{foo}} order by a.name desc limit {{limit}}`,
			data: map[string]interface{}{
				"age":   20,
				"foo":   30,
				"limit": 40,
			},
			cond: "select a.name,a.age from tb1 as a join tb2 as b on a.id=b.id where a.age>? and b.age<? order by a.name desc limit ?",
			vals: []interface{}{20, 30, 40},
			err:  nil,
		},
		{
			sql: `select * from tb where age in {{age}}`,
			data: map[string]interface{}{
				"age": []int{1},
			},
			cond: `select * from tb where age in (?)`,
			vals: []interface{}{1},
			err:  nil,
		},
		{
			sql: `select {{foo}},{{bar}} from tb where age={{age}} and address in {{addr}}`,
			data: map[string]interface{}{
				"foo":  "f1",
				"bar":  "f2",
				"age":  10,
				"addr": []string{"beijing", "shanghai", "chengdu"},
			},
			cond: `select ?,? from tb where age=? and address in (?,?,?)`,
			vals: []interface{}{"f1", "f2", 10, "beijing", "shanghai", "chengdu"},
			err:  nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range testData {
		cond, vals, err := NamedQuery(tc.sql, tc.data)
		if !ass.Equal(tc.err, err) {
			return
		}
		ass.Equal(tc.cond, cond)
		ass.Equal(tc.vals, vals)
	}
}

func Test_BuildIN(t *testing.T) {
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
					"age in":   []int{1, 3, 5, 7, 9},
					"faith <>": "Muslim",
					"_orderby": "age DESC",
					"_groupby": "department",
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=? AND qq=? AND age IN (?,?,?,?,?) AND faith!=?) GROUP BY department ORDER BY age DESC",
				vals: []interface{}{"bar", "tt", 1, 3, 5, 7, 9, "Muslim"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":    "bar",
					"age IN": []int{1, 3, 5, 7, 9},
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=? AND age IN (?,?,?,?,?))",
				vals: []interface{}{"bar", 1, 3, 5, 7, 9},
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

func Benchmark_BuildIN(b *testing.B) {
	where := map[string]interface{}{
		"age": []uint64{1, 3, 5, 7, 9},
	}
	for i := 0; i < b.N; i++ {
		convertWhereMapToWhereMapSlice(where, opIn)
	}
}

func Test_BuildOrderBy(t *testing.T) {
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
					"_orderby": "age DESC,id ASC",
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=?) ORDER BY age DESC,id ASC",
				vals: []interface{}{"bar"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"foo":      "bar",
					"_orderby": "RAND()",
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=?) ORDER BY RAND()",
				vals: []interface{}{"bar"},
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

func Test_Where_Null(t *testing.T) {
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
					"aa": IsNotNull,
				},
				fields: []string{"id", "name"},
			},
			out: outStruct{
				cond: "SELECT id,name FROM tb WHERE (aa IS NOT NULL)",
				vals: nil,
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"aa":  IsNotNull,
					"foo": "bar",
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=? AND aa IS NOT NULL)",
				vals: []interface{}{"bar"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"aa":  IsNull,
					"foo": "bar",
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=? AND aa IS NULL)",
				vals: []interface{}{"bar"},
				err:  nil,
			},
		},
		{
			in: inStruct{
				table: "tb",
				where: map[string]interface{}{
					"aa":  IsNull,
					"foo": "bar",
					"bb":  IsNotNull,
				},
				fields: []string{"id", "name", "age"},
			},
			out: outStruct{
				cond: "SELECT id,name,age FROM tb WHERE (foo=? AND aa IS NULL AND bb IS NOT NULL)",
				vals: []interface{}{"bar"},
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

func TestBuildSelect_Limit(t *testing.T) {
	var testCase = []struct {
		limit  []uint
		err    error
		expect []interface{}
	}{
		{
			limit:  []uint{10, 20},
			err:    nil,
			expect: []interface{}{10, 20},
		},
		{
			limit:  []uint{0, 1},
			err:    nil,
			expect: []interface{}{0, 1},
		},
		{
			limit:  []uint{1},
			err:    nil,
			expect: []interface{}{0, 1},
		},
		{
			limit:  []uint{20, 10},
			err:    nil,
			expect: []interface{}{20, 10},
		},
		{
			limit:  []uint{},
			err:    errLimitValueLength,
			expect: nil,
		},
		{
			limit:  []uint{1, 2, 3},
			err:    errLimitValueLength,
			expect: nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range testCase {
		cond, vals, err := BuildSelect("tb", map[string]interface{}{
			"_limit": tc.limit,
		}, nil)
		ass.Equal(tc.err, err)
		if tc.err == nil {
			ass.Equal(`SELECT * FROM tb LIMIT ?,?`, cond, "where=%+v", tc.limit)
			ass.Equal(tc.expect, vals)
		}
	}
}

func Test_NotIn(t *testing.T) {
	table := "some_table"
	fields := []string{"name", "age", "sex"}
	where := []map[string]interface{}{
		{
			"city in":            []string{"beijing", "shanghai"},
			"age >":              35,
			"address":            IsNotNull,
			" hobbies not in   ": []string{"baseball", "swim", "running"},
			"_groupby":           "department",
			"_orderby":           "bonus DESC",
		},
		{
			"city IN":            []string{"beijing", "shanghai"},
			"age >":              35,
			"address":            IsNotNull,
			" hobbies NOT IN   ": []string{"baseball", "swim", "running"},
			"_groupby":           "department",
			"_orderby":           "bonus DESC",
		},
	}

	expectCond := `SELECT name,age,sex FROM some_table WHERE (city IN (?,?) AND hobbies NOT IN (?,?,?) AND age>? AND address IS NOT NULL) GROUP BY department ORDER BY bonus DESC`
	expectVals := []interface{}{"beijing", "shanghai", "baseball", "swim", "running", 35}

	ass := assert.New(t)
	for _, w := range where {
		cond, vals, err := BuildSelect(table, w, fields)
		ass.NoError(err)
		ass.Equal(expectCond, cond)
		ass.Equal(expectVals, vals)
	}
}

func TestBuildBetween(t *testing.T) {
	table := "tb"
	fields := []string{"foo", "bar"}
	where := []map[string]interface{}{
		{
			"city in ":    []string{"beijing", "chengdu"},
			"age between": []int{10, 30},
			"name":        "caibirdme",
		},
		{
			"city IN ":    []string{"beijing", "chengdu"},
			"age BETWEEN": []int{10, 30},
			"name":        "caibirdme",
		},
	}

	expectCond := "SELECT foo,bar FROM tb WHERE (name=? AND city IN (?,?) AND (age BETWEEN ? AND ?))"
	expectVals := []interface{}{"caibirdme", "beijing", "chengdu", 10, 30}

	ass := assert.New(t)
	for _, w := range where {
		cond, vals, err := BuildSelect(table, w, fields)
		ass.NoError(err)
		ass.Equal(expectCond, cond)
		ass.Equal(expectVals, vals)
	}
}

func TestBuildNotBetween(t *testing.T) {
	table := "tb"
	fields := []string{"foo", "bar"}
	where := []map[string]interface{}{
		{
			"city in ":        []string{"beijing", "chengdu"},
			"age not between": []int{10, 30},
			"name":            "caibirdme",
			"_limit":          []uint{10, 20},
		},
		{
			"city IN ":        []string{"beijing", "chengdu"},
			"age NOT BETWEEN": []int{10, 30},
			"name":            "caibirdme",
			"_limit":          []uint{10, 20},
		},
	}

	expectCond := "SELECT foo,bar FROM tb WHERE (name=? AND city IN (?,?) AND (age NOT BETWEEN ? AND ?)) LIMIT ?,?"
	expectVals := []interface{}{"caibirdme", "beijing", "chengdu", 10, 30, 10, 20}

	ass := assert.New(t)
	for _, w := range where {
		cond, vals, err := BuildSelect(table, w, fields)
		ass.NoError(err)
		ass.Equal(expectCond, cond)
		ass.Equal(expectVals, vals)
	}
}

func TestBuildCombinedBetween(t *testing.T) {
	table := "tb"
	fields := []string{"foo", "bar"}
	where := []map[string]interface{}{
		{
			"city in ":        []string{"beijing", "chengdu"},
			"age not between": []int{10, 30},
			"name":            "caibirdme",
			"score between":   []float64{3.5, 7.2},
			"_limit":          []uint{10, 20},
		},
		{
			"city IN ":        []string{"beijing", "chengdu"},
			"age NOT BETWEEN": []int{10, 30},
			"name":            "caibirdme",
			"score BETWEEN":   []float64{3.5, 7.2},
			"_limit":          []uint{10, 20},
		},
	}

	expectCond := "SELECT foo,bar FROM tb WHERE (name=? AND city IN (?,?) AND (score BETWEEN ? AND ?) AND (age NOT BETWEEN ? AND ?)) LIMIT ?,?"
	expectVals := []interface{}{"caibirdme", "beijing", "chengdu", 3.5, 7.2, 10, 30, 10, 20}

	ass := assert.New(t)
	for _, w := range where {
		cond, vals, err := BuildSelect(table, w, fields)
		ass.NoError(err)
		ass.Equal(expectCond, cond)
		ass.Equal(expectVals, vals)
	}
}

func TestLike(t *testing.T) {
	type inStruct struct {
		table  string
		where  []map[string]interface{}
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
				where: []map[string]interface{}{
					{
						"bar like": "haha%",
						"baz like": "%some",
						"foo":      1,
					},
					{
						"bar LIKE": "haha%",
						"baz LIKE": "%some",
						"foo":      1,
					},
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
				where: []map[string]interface{}{
					{
						"bar like": "haha%",
						"baz like": "%some",
						"foo":      1,
						"age in":   []interface{}{1, 3, 5, 7, 9},
					},
					{
						"bar LIKE": "haha%",
						"baz LIKE": "%some",
						"foo":      1,
						"age IN":   []interface{}{1, 3, 5, 7, 9},
					},
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
				where: []map[string]interface{}{
					{
						"name like": "%James",
					},
					{
						"name LIKE": "%James",
					},
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
		for _, w := range tc.in.where {
			cond, vals, err := BuildSelect(tc.in.table, w, tc.in.fields)
			ass.Equal(tc.out.err, err)
			ass.Equal(tc.out.cond, cond)
			ass.Equal(tc.out.vals, vals)
		}
	}
}

func TestNotLike(t *testing.T) {
	table := "tb"
	where := []map[string]interface{}{
		{
			"name  not    like  ": "%ny",
		},
		{
			"name  NOT    LIKE  ": "%ny",
		},
	}

	expectCond := "SELECT * FROM tb WHERE (name NOT LIKE ?)"
	expectVals := []interface{}{"%ny"}

	ass := assert.New(t)
	for _, w := range where {
		cond, vals, err := BuildSelect(table, w, nil)
		ass.NoError(err)
		ass.Equal(expectCond, cond)
		ass.Equal(expectVals, vals)
	}
}

func TestNotLike_1(t *testing.T) {
	table := "tb"
	where := []map[string]interface{}{
		{
			"name  not like  ": "%ny",
			"age":              20,
		},
		{
			"name  NOT LIKE  ": "%ny",
			"age":              20,
		},
	}

	expectCond := "SELECT * FROM tb WHERE (age=? AND name NOT LIKE ?)"
	expectVals := []interface{}{20, "%ny"}

	ass := assert.New(t)
	for _, w := range where {
		cond, vals, err := BuildSelect(table, w, nil)
		ass.NoError(err)
		ass.Equal(expectCond, cond)
		ass.Equal(expectVals, vals)
	}
}

func TestFixBug_insert_quote_field(t *testing.T) {
	cond, vals, err := BuildInsert("tb", []map[string]interface{}{
		{
			"id":      1,
			"`order`": 2,
			"`id`":    3, // I know this is forbidden, but just for test
		},
	})
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal("INSERT INTO tb (`id`,`order`,id) VALUES (?,?,?)", cond)
	ass.Equal([]interface{}{3, 2, 1}, vals)
}

func TestInsertOnDuplicate(t *testing.T) {
	cond, vals, err := BuildInsertOnDuplicate(
		"tb",
		[]map[string]interface{}{
			{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		map[string]interface{}{
			"c": 4,
		},
	)
	ass := assert.New(t)
	ass.NoError(err)
	ass.Equal("INSERT INTO tb (a,b,c) VALUES (?,?,?) ON DUPLICATE KEY UPDATE c=?", cond)
	ass.Equal([]interface{}{1, 2, 3, 4}, vals)
}
