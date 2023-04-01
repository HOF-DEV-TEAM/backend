package urlqueryhelper

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gofrs/uuid"
)

func Bind(structValue interface{}, r *http.Request) error {
	fields := reflect.TypeOf(structValue)
	values := reflect.ValueOf(structValue)
	num := fields.Elem().NumField()

	for i := 0; i < num; i++ {
		field := fields.Elem().Field(i)

		key := field.Tag.Get("filter")
		kind := field.Type.Kind()

		val := r.URL.Query().Get(key)

		value := values.Elem().Field(i)
		switch kind {
		case reflect.String:
			value.SetString(val)
		default:
			return errors.New("Does not support type of struct")
		}
	}
	return nil
}

type QueryHelper interface {
	SetQueryHelper(structValue interface{}) string
	WhereQueryHelper(structValue interface{}) string
}

type queryHandler struct {
}

var _ QueryHelper = queryHandler{}

func NewQueryHelper() QueryHelper {
	return queryHandler{}
}

func (handler queryHandler) SetQueryHelper(structValue interface{}) string {
	fields := reflect.TypeOf(structValue)

	values := reflect.ValueOf(structValue)

	fieldNumbers := values.NumField()
	var setQuery string
	for i := 0; i < fieldNumbers; i++ {
		fieldProperties := fields.Field(i)
		value := values.Field(i)
		goTag := fieldProperties.Tag.Get("sql")
		goType := fieldProperties.Type.Kind()

		var val string

		switch goType {
		case reflect.String:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				val = value.Interface().(string)
			}

		case reflect.Int:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				val = string(rune(value.Interface().(int)))
			}

		case reflect.Float64:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				val = string(rune(value.Interface().(float64)))
			}

		case reflect.Array:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				if v, ok := value.Interface().(uuid.UUID); ok {
					val = v.String()
				}
			}
		}

		if val != "" {
			queryString := goTag + "=" + "'" + val + "'"

			setQuery = setQueryIndex(i, fieldNumbers, setQuery, queryString)
		}

	}
	setQuery = strings.TrimSuffix(setQuery, ", ")
	return setQuery
}

func (handler queryHandler) WhereQueryHelper(structValue interface{}) string {
	fields := reflect.TypeOf(structValue)

	values := reflect.ValueOf(structValue)

	fieldNumbers := values.NumField()
	var whereQuery string
	for i := 0; i < fieldNumbers; i++ {
		fieldProperties := fields.Field(i)
		value := values.Field(i)
		goTag := fieldProperties.Tag.Get("sql")
		goType := fieldProperties.Type.Kind()

		var val string

		switch goType {
		case reflect.String:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				val = value.Interface().(string)
			}

		case reflect.Int:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				val = string(rune(value.Interface().(int)))
			}

		case reflect.Float64:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				val = string(rune(value.Interface().(float64)))
			}

		case reflect.Array:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				if v, ok := value.Interface().(uuid.UUID); ok {
					val = v.String()
				}
			}
		}

		if val != "" {
			queryString := goTag + "=" + "'" + val + "'"
			whereQuery = whereQueryIndex(i, fieldNumbers, whereQuery, queryString)

		}
	}
	whereQuery = strings.TrimSuffix(whereQuery, " ")
	return whereQuery
}

func setQueryIndex(i, fieldNumbers int, setQuery string, val string) string {
	switch i {
	case 0:
		setQuery += val + ","
	case fieldNumbers - 1:
		setQuery += val
	default:
		setQuery += val + ", "
	}

	return setQuery
}

func whereQueryIndex(i, fieldNumbers int, goTag string, val string) string {
	var whereQuery string
	switch i {
	case 0:
		whereQuery += val
	case fieldNumbers - 1:
		whereQuery += " AND " + val
	default:
		whereQuery += " AND " + val + " "
	}

	return whereQuery
}
