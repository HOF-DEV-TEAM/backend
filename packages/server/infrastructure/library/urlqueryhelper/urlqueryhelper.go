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

func SqlQueryHelper(where, set bool, structValue interface{}) (string, string) {
	fields := reflect.TypeOf(structValue)

	values := reflect.ValueOf(structValue)

	fieldNumbers := values.NumField()
	var whereQuery, setQuery string
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
			whereString := goTag + "=" + "'" + val + "'"
			whereQuery, setQuery = getIndex(where, set, i, fieldNumbers, whereQuery, setQuery, whereString)
			setQuery = strings.TrimSuffix(setQuery, ", ")
			whereQuery = strings.TrimSuffix(whereQuery, " ")
		}
	
	}
	return whereQuery, setQuery
}

func getIndex(where, set bool, i, fieldNumbers int, whereQuery, setQuery string, val string) (string, string) {
	switch i {
	case 0:
		if where {
			whereQuery += val
		}
		if set {
			setQuery += val + ", "
		}
	case fieldNumbers - 1:
		if where {
			whereQuery += " AND " + val
		}
		if set {
			setQuery += val
		}
	default:
		if where {
			whereQuery += " AND " + val + " "
		}
		if set {
			setQuery += val + ", "
		}
	}

	return whereQuery, setQuery
}
