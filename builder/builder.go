package builder

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var (
	errSplitEmptyKey = errors.New("[builder] couldn't split a empty string")
	errSplitOrderBy  = errors.New(`[builder] the value of _orderby should be "fieldName direction [,fieldName direction]"`)
	// ErrUnsupportedOperator reports there's unsupported operators in where-condition
	ErrUnsupportedOperator       = errors.New("[builder] unsupported operator")
	errWhereInType               = errors.New(`[builder] the value of "xxx in" must be of []interface{} type`)
	errGroupByValueType          = errors.New(`[builder] the value of "_groupby" must be of string type`)
	errLimitValueType            = errors.New(`[builder] the value of "_limit" must be of []uint type`)
	errLimitValueLength          = errors.New(`[builder] the value of "_limit" must contain one or two uint elements`)
	errEmptyINCondition          = errors.New(`[builder] the value of "in" must contain at least one element`)
	errHavingValueType           = errors.New(`[builder] the value of "_having" must be of map[string]interface{}`)
	errHavingUnsupportedOperator = errors.New(`[builder] "_having" contains unsupported operator`)
)

type whereMapSet struct {
	set map[string]map[string]interface{}
}

func (w *whereMapSet) add(op, field string, val interface{}) {
	if nil == w.set {
		w.set = make(map[string]map[string]interface{})
	}
	s, ok := w.set[op]
	if !ok {
		s = make(map[string]interface{})
		w.set[op] = s
	}
	s[field] = val
}

type eleOrderBy struct {
	field, order string
}

type eleLimit struct {
	begin, step uint
}

// BuildSelect work as its name says.
// supported operators including: =,in,>,>=,<,<=,<>,!=.
// key without operator will be regarded as =.
// special key begin with _: _orderby,_groupby,_limit,_having.
// the value of _orderby must be a string separated by a space(ie:map[string]interface{}{"_orderby": "fieldName desc"}).
// the value of _limit must be a slice whose type should be []uint and must contain two uints(ie: []uint{0, 100}).
// the value of _having must be a map just like where but only support =,in,>,>=,<,<=,<>,!=
// for more examples,see README.md or open a issue.
func BuildSelect(table string, where map[string]interface{}, selectField []string) (cond string, vals []interface{}, err error) {
	var orderBy []eleOrderBy
	var limit *eleLimit
	var groupBy string
	var having map[string]interface{}
	copiedWhere := copyWhere(where)
	if val, ok := copiedWhere["_orderby"]; ok {
		eleOrderBy, e := splitOrderBy(val.(string))
		if e != nil {
			err = e
			return
		}
		orderBy = eleOrderBy
		delete(copiedWhere, "_orderby")
	}
	if val, ok := copiedWhere["_groupby"]; ok {
		s, ok := val.(string)
		if !ok {
			err = errGroupByValueType
			return
		}
		groupBy = s
		delete(copiedWhere, "_groupby")
		if h, ok := copiedWhere["_having"]; ok {
			having, err = resolveHaving(h)
			if nil != err {
				return
			}
		}
	}
	if _, ok := copiedWhere["_having"]; ok {
		delete(copiedWhere, "_having")
	}
	if val, ok := copiedWhere["_limit"]; ok {
		arr, ok := val.([]uint)
		if !ok {
			err = errLimitValueType
			return
		}
		if len(arr) != 2 {
			if len(arr) == 1 {
				arr = []uint{0, arr[0]}
			} else {
				err = errLimitValueLength
				return
			}
		}
		begin, step := arr[0], arr[1]
		limit = &eleLimit{
			begin: begin,
			step:  step,
		}
		delete(copiedWhere, "_limit")
	}
	conditions, release, err := getWhereConditions(copiedWhere)
	if nil != err {
		return
	}
	defer release()
	if having != nil {
		havingCondition, release1, err1 := getWhereConditions(having)
		if nil != err1 {
			err = err1
			return
		}
		defer release1()
		conditions = append(conditions, nilComparable(0))
		conditions = append(conditions, havingCondition...)
	}
	return buildSelect(table, selectField, groupBy, orderBy, limit, conditions...)
}

func copyWhere(src map[string]interface{}) (target map[string]interface{}) {
	target = make(map[string]interface{})
	for k, v := range src {
		target[k] = v
	}
	return
}

func resolveHaving(having interface{}) (map[string]interface{}, error) {
	var havingMap map[string]interface{}
	var ok bool
	if havingMap, ok = having.(map[string]interface{}); !ok {
		return nil, errHavingValueType
	}
	copiedMap := make(map[string]interface{})
	for key, val := range havingMap {
		_, operator, err := splitKey(key)
		if nil != err {
			return nil, err
		}
		if !isStringInSlice(operator, opOrder) {
			return nil, errHavingUnsupportedOperator
		}
		copiedMap[key] = val
	}
	return copiedMap, nil
}

// BuildUpdate work as its name says
func BuildUpdate(table string, where map[string]interface{}, update map[string]interface{}) (string, []interface{}, error) {
	conditions, release, err := getWhereConditions(where)
	if nil != err {
		return "", nil, err
	}
	defer release()
	return buildUpdate(table, update, conditions...)
}

// BuildDelete work as its name says
func BuildDelete(table string, where map[string]interface{}) (string, []interface{}, error) {
	conditions, release, err := getWhereConditions(where)
	if nil != err {
		return "", nil, err
	}
	defer release()
	return buildDelete(table, conditions...)
}

// BuildInsert work as its name says
func BuildInsert(table string, data []map[string]interface{}) (string, []interface{}, error) {
	return buildInsert(table, data, commonInsert)
}

// BuildInsertIgnore work as its name says
func BuildInsertIgnore(table string, data []map[string]interface{}) (string, []interface{}, error) {
	return buildInsert(table, data, ignoreInsert)
}

// BuildReplaceInsert work as its name says
func BuildReplaceInsert(table string, data []map[string]interface{}) (string, []interface{}, error) {
	return buildInsert(table, data, replaceInsert)
}

var (
	cpPool = sync.Pool{
		New: func() interface{} {
			return make([]Comparable, 0)
		},
	}
)

func getCpPool() ([]Comparable, func()) {
	obj := cpPool.Get().([]Comparable)
	return obj[:0], func() { cpPool.Put(obj) }
}

func emptyFunc() {}

func isStringInSlice(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func getWhereConditions(where map[string]interface{}) ([]Comparable, func(), error) {
	if len(where) == 0 {
		return nil, emptyFunc, nil
	}
	wms := &whereMapSet{}
	var field, operator string
	var err error
	for key, val := range where {
		field, operator, err = splitKey(key)
		if !isStringInSlice(operator, opOrder) {
			return nil, emptyFunc, ErrUnsupportedOperator
		}
		if nil != err {
			return nil, emptyFunc, err
		}
		if _, ok := val.(NullType); ok {
			operator = opNull
		}
		wms.add(operator, field, val)
	}

	return buildWhereCondition(wms)
}

const (
	opEq         = "="
	opNe1        = "!="
	opNe2        = "<>"
	opIn         = "in"
	opNotIn      = "not in"
	opGt         = ">"
	opGte        = ">="
	opLt         = "<"
	opLte        = "<="
	opLike       = "like"
	opNotLike    = "not like"
	opBetween    = "between"
	opNotBetween = "not between"
	// special
	opNull = "null"
)

type compareProducer func(m map[string]interface{}) (Comparable, error)

var op2Comparable = map[string]compareProducer{
	opEq: func(m map[string]interface{}) (Comparable, error) {
		return Eq(m), nil
	},
	opNe1: func(m map[string]interface{}) (Comparable, error) {
		return Ne(m), nil
	},
	opNe2: func(m map[string]interface{}) (Comparable, error) {
		return Ne(m), nil
	},
	opIn: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m)
		if nil != err {
			return nil, err
		}
		return In(wp), nil
	},
	opNotIn: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m)
		if nil != err {
			return nil, err
		}
		return NotIn(wp), nil
	},
	opBetween: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m)
		if nil != err {
			return nil, err
		}
		return Between(wp), nil
	},
	opNotBetween: func(m map[string]interface{}) (Comparable, error) {
		wp, err := convertWhereMapToWhereMapSlice(m)
		if nil != err {
			return nil, err
		}
		return NotBetween(wp), nil
	},
	opGt: func(m map[string]interface{}) (Comparable, error) {
		return Gt(m), nil
	},
	opGte: func(m map[string]interface{}) (Comparable, error) {
		return Gte(m), nil
	},
	opLt: func(m map[string]interface{}) (Comparable, error) {
		return Lt(m), nil
	},
	opLte: func(m map[string]interface{}) (Comparable, error) {
		return Lte(m), nil
	},
	opLike: func(m map[string]interface{}) (Comparable, error) {
		return Like(m), nil
	},
	opNotLike: func(m map[string]interface{}) (Comparable, error) {
		return NotLike(m), nil
	},
	opNull: func(m map[string]interface{}) (Comparable, error) {
		return nullCompareble(m), nil
	},
}

var opOrder = []string{opEq, opIn, opNe1, opNe2, opNotIn, opGt, opGte, opLt, opLte, opLike, opNotLike, opBetween, opNotBetween, opNull}

func buildWhereCondition(mapSet *whereMapSet) ([]Comparable, func(), error) {
	cpArr, release := getCpPool()
	for _, operator := range opOrder {
		whereMap, ok := mapSet.set[operator]
		if !ok {
			continue
		}
		f, ok := op2Comparable[operator]
		if !ok {
			release()
			return nil, emptyFunc, ErrUnsupportedOperator
		}
		cp, err := f(whereMap)
		if nil != err {
			release()
			return nil, emptyFunc, err
		}
		cpArr = append(cpArr, cp)
	}
	return cpArr, release, nil
}

func convertWhereMapToWhereMapSlice(where map[string]interface{}) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})
	for key, val := range where {
		vals, ok := convertInterfaceToMap(val)
		if !ok {
			return nil, errWhereInType
		}
		if 0 == len(vals) {
			return nil, errEmptyINCondition
		}
		result[key] = vals
	}
	return result, nil
}

func convertInterfaceToMap(val interface{}) ([]interface{}, bool) {
	s := reflect.ValueOf(val)
	if s.Kind() != reflect.Slice {
		return nil, false
	}
	interfaceSlice := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		interfaceSlice[i] = s.Index(i).Interface()
	}
	return interfaceSlice, true
}

func splitKey(key string) (field string, operator string, err error) {
	key = strings.Trim(key, " ")
	if "" == key {
		err = errSplitEmptyKey
		return
	}
	idx := strings.IndexByte(key, ' ')
	if idx == -1 {
		field = key
		operator = "="
	} else {
		field = key[:idx]
		operator = strings.Trim(key[idx+1:], " ")
		operator = removeInnerSpace(operator)
	}
	return
}

func removeInnerSpace(operator string) string {
	n := len(operator)
	firstSpace := strings.IndexByte(operator, ' ')
	if firstSpace == -1 {
		return operator
	}
	lastSpace := firstSpace
	for i := firstSpace+1; i<n; i++ {
		if operator[i] == ' ' {
			lastSpace = i
		} else {
			break
		}
	}
	return operator[:firstSpace] + operator[lastSpace:]
}

func splitOrderBy(orderby string) ([]eleOrderBy, error) {
	var err error
	var eleOrder []eleOrderBy
	for _, val := range strings.Split(orderby, ",") {
		val = strings.Trim(val, " ")
		idx := strings.IndexByte(val, ' ')
		if idx == -1 {
			err = errSplitOrderBy
			return eleOrder, err
		}
		field := val[:idx]
		direction := strings.Trim(val[idx+1:], " ")
		eleOrder = append(eleOrder, eleOrderBy{
			field: field,
			order: direction,
		})
	}
	return eleOrder, err
}

const (
	paramPlaceHolder = "?"
)

var searchHandle = regexp.MustCompile(`{{\S+?}}`)

// NamedQuery is used for expressing complex query
func NamedQuery(sql string, data map[string]interface{}) (string, []interface{}, error) {
	length := len(data)
	if length == 0 {
		return sql, nil, nil
	}
	vals := make([]interface{}, 0, length)
	var err error
	cond := searchHandle.ReplaceAllStringFunc(sql, func(paramName string) string {
		paramName = strings.TrimRight(strings.TrimLeft(paramName, "{"), "}")
		val, ok := data[paramName]
		if !ok {
			err = fmt.Errorf("%s not found", paramName)
			return ""
		}
		v := reflect.ValueOf(val)
		if v.Type().Kind() != reflect.Slice {
			vals = append(vals, val)
			return paramPlaceHolder
		}
		length := v.Len()
		for i := 0; i < length; i++ {
			vals = append(vals, v.Index(i).Interface())
		}
		return createMultiPlaceholders(length)
	})
	if nil != err {
		return "", nil, err
	}
	return cond, vals, nil
}

func createMultiPlaceholders(num int) string {
	if 0 == num {
		return ""
	}
	length := (num << 1) | 1
	buff := make([]byte, length)
	buff[0], buff[length-1] = '(', ')'
	ll := length - 2
	for i := 1; i <= ll; i += 2 {
		buff[i] = '?'
	}
	ll = length - 3
	for i := 2; i <= ll; i += 2 {
		buff[i] = ','
	}
	return string(buff)
}
