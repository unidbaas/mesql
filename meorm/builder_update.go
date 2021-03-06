package meorm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ipiao/mesql/medb"
)

// UpdateBuilder 更新构造器
// 只支持单个更新
type UpdateBuilder struct {
	*where
	builder    *BaseBuilder
	connname   string
	table      string
	columns    []string
	values     []interface{}
	orderbys   []string
	limit      int64
	limitvalid bool
	err        error
	sql        string
	args       []interface{}
}

// set项
type setClause struct {
	column string
	value  interface{}
}

// reset
func (u *UpdateBuilder) reset() *UpdateBuilder {
	u.table = ""
	u.columns = u.columns[:0]
	u.values = u.values[:0]
	u.where = new(where)
	u.where.dialect = u.dialect
	u.orderbys = u.orderbys[:0]
	u.limit = 0
	u.limitvalid = false
	u.err = nil
	u.sql = ""
	u.args = u.args[:0]
	return u
}

// Set 设置值
func (u *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	u.columns = append(u.columns, column)
	u.values = append(u.values, value)
	return u
}

// Colunms 设置更新列
func (u *UpdateBuilder) Colunms(column ...string) *UpdateBuilder {
	u.columns = append(u.columns, column...)
	return u
}

// Models 插入结构体
// models必须为结构体、结构体数组，或者相应的指针
func (u *UpdateBuilder) Models(models interface{}) *UpdateBuilder {
	var t = reflect.TypeOf(models)
	var v = reflect.ValueOf(models)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	var cols = getColumns(v)
	var vals = getValues(v)
	if len(u.columns) == 0 && len(cols) == 0 {
		u.err = errors.New("columns can not be null")
		return u
	}
	//
	if len(u.columns) == 0 {
		u.columns = cols
	}
	if len(u.table) == 0 {
		u.table = getTableName(v)
	}
	// 获取列名和结构体字段列的映射
	var tempMap = make(map[int]int, len(u.columns))
	for i, column := range u.columns {
		flag := 0
		for j, col := range cols {
			if column == col {
				tempMap[i] = j
				flag++
				break
			}
		}
		if flag == 0 {
			u.err = fmt.Errorf("can not find column %s in models", column)
		}
	}
	// 拼接值
	for _, val := range vals {
		var value = make([]interface{}, len(u.columns))
		for i, v := range tempMap {
			value[i] = val[v]
		}
		u.values = append(u.values, value...)
	}
	return u
}

// Where where 条件
func (u *UpdateBuilder) Where(condition string, args ...interface{}) *UpdateBuilder {
	u.conds = append(u.conds, &condValues{
		condition: condition,
		values:    args,
	})
	return u
}

// OrderBy orderby 条件
func (u *UpdateBuilder) OrderBy(order string) *UpdateBuilder {
	u.orderbys = append(u.orderbys, order)
	return u
}

// Limit limit
func (u *UpdateBuilder) Limit(limit int64) *UpdateBuilder {
	u.limitvalid = true
	u.limit = limit
	return u
}

// 生成sql
func (u *UpdateBuilder) tosql() (string, []interface{}) {
	if u.where.err != nil {
		u.err = u.where.err
		return "", nil
	}

	holder := u.builder.dialect.Holder()
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	var args []interface{}
	buf.WriteString("UPDATE ")
	buf.WriteString(u.table)
	buf.WriteString(" SET ")
	for i, s := range u.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s + "=")
		buf.WriteByte(holder)
	}
	args = append(args, u.values...)

	if len(u.conds) > 0 {
		buf.WriteString(" WHERE ")
		for i, cond := range u.conds {
			if i > 0 {
				buf.WriteString(" AND (")
			} else {
				buf.WriteByte('(')
			}
			buf.WriteString(cond.condition)
			buf.WriteByte(')')
			if len(cond.values) > 0 {
				args = append(args, cond.values...)
			}
		}
	}

	if len(u.orderbys) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, s := range u.orderbys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(s)
		}
	}
	if u.limitvalid {
		buf.WriteString(" LIMIT ")
		buf.WriteByte(holder)
		args = append(args, u.limit)
	}

	u.sql = buf.String()
	u.args = args
	return u.sql, u.args
}

// ToSQL tosql
func (u *UpdateBuilder) ToSQL() (string, []interface{}) {
	if len(u.sql) > 0 {
		return u.sql, u.args
	}
	return u.tosql()
}

// Exec 执行
func (u *UpdateBuilder) Exec() *medb.Result {
	if u.err != nil {
		var res = new(medb.Result).SetErr(u.err)
		return res
	}
	return u.builder.Exec(u)
}
