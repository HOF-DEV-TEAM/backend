package urlqueryhelper

import (
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"reflect"
	"strconv"
	"strings"
)

// Todo: add support for joins, sorting, grouping and other boolean selects

type Query struct{}

type QueryBuilder struct {
	Val   reflect.Value
	GoTag string
	Query string
}

func (q *QueryBuilder) isEmpty(val reflect.Value) bool {
	return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}

func (q *QueryBuilder) formatQuery(delim string) string {
	return strings.Replace(strings.TrimSpace(q.Query), " ", delim, -1)
}

func (q *QueryBuilder) clearQuery() {
	q.Query = ""
}

func (q *QueryBuilder) setString(val string) {
	if val != "" {
		q.Query += fmt.Sprintf("%s='%s' ", q.GoTag, val)
	}
}

func (q *QueryBuilder) setNull() {
	q.Query += fmt.Sprintf("%s=NULL ", q.GoTag)
}

func (q *QueryBuilder) BuildString(val reflect.Value) {
	if !q.isEmpty(val) {
		q.setString(val.Interface().(string))
	}
}

func (q *QueryBuilder) BuildInt(val reflect.Value) {
	if !q.isEmpty(val) {
		q.setString(fmt.Sprintf("%d", val.Interface().(int)))
	}
}

func (q *QueryBuilder) BuildFloat(val reflect.Value) {
	if !q.isEmpty(val) {
		q.setString(string(rune(val.Interface().(float64))))
	}
}

func (q *QueryBuilder) BuildArray(val reflect.Value) {
	if !q.isEmpty(val) {
		if v, ok := val.Interface().(uuid.UUID); ok {
			q.setString(v.String())
		}
	}
}

func (q *QueryBuilder) BuildPtr(val reflect.Value) {
	if !q.isEmpty(val) {
		boolValue := val.Interface().(*bool)
		q.setString(strconv.FormatBool(*boolValue))
	}
}

func (q *QueryBuilder) BuildSqlNullString(val reflect.Value) {
	if !q.isEmpty(val) {
		v := val.Interface().(sql.NullString)

		if v.Valid {
			str := v.String
			if str == "" {
				q.setNull()
				return
			}
			q.setString(str)
		}
	}
}

func (qB *QueryBuilder) parseStruct(structValue interface{}) *QueryBuilder {
	fields := reflect.TypeOf(structValue)
	values := reflect.ValueOf(structValue)
	fieldNumbers := values.NumField()

	for i := 0; i < fieldNumbers; i++ {
		fieldProperties := fields.Field(i)
		value := values.Field(i)
		qB.GoTag = fieldProperties.Tag.Get("sql")

		goType := fieldProperties.Type.Kind()

		switch goType {
		case reflect.String:
			qB.BuildString(value)

		case reflect.Int:
			qB.BuildInt(value)

		case reflect.Float64:
			qB.BuildFloat(value)

		case reflect.Array:
			qB.BuildArray(value)
		case reflect.Ptr:
			qB.BuildPtr(value)

		default:
			switch value.Type().String() {
			case "sql.NullString":
				qB.BuildSqlNullString(value)
			}
		}
	}
	return qB
}

func (q *QueryBuilder) Where(structValue any) string {
	qb := q.parseStruct(structValue)
	defer q.clearQuery()

	return qb.formatQuery(" AND ")
}

func (q *QueryBuilder) Set(structValue any) string {
	qb := q.parseStruct(structValue)
	defer q.clearQuery()
	return qb.formatQuery(",")
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}
