package urlqueryhelper

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
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
			return errors.New("does not support type of struct")
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

		case reflect.Bool:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				boolVal := value.Interface().(bool)
				val = strconv.FormatBool(boolVal)
			}

		case reflect.Ptr:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				boolValue := value.Interface().(*bool)
				val = strconv.FormatBool(*boolValue)
			}

		case reflect.Struct:
			if !reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface()) {
				if value.Kind() == reflect.Struct && fieldProperties.Type != reflect.TypeOf(time.Time{}) {
					innerVal := fmt.Sprintf("%s: %s, ", fieldProperties.Name, handler.SetQueryHelper(value.Interface()))
					if fieldProperties.Type == reflect.TypeOf(sql.NullString{}) {
						switch {
						case strings.Contains(innerVal, "LastUpdated: ="):
							innerVal = strings.Replace(innerVal, "LastUpdated: =", "", 1)
							innerVal = strings.Replace(innerVal, ",='true',", "", 1)
						case strings.Contains(innerVal, "DeletedAt: ="):
							innerVal = strings.Replace(innerVal, "DeletedAt: =", "", 1)
							innerVal = strings.Replace(innerVal, ",='true',", "", 1)
						case strings.Contains(innerVal, "DateAdded: ="):
							innerVal = strings.Replace(innerVal, "DateAdded: =", "", 1)
							innerVal = strings.Replace(innerVal, ",='true',", "", 1)
						case strings.Contains(innerVal, "SeriesID: ="):
							innerVal = strings.Replace(innerVal, "SeriesID: =", "", 1)
							innerVal = strings.Replace(innerVal, ",='true',", "", 1)
						case strings.Contains(innerVal, "DateReleased: ="):
							innerVal = strings.Replace(innerVal, "DateReleased: =", "", 1)
							innerVal = strings.Replace(innerVal, ",='true',", "", 1)
						}

						val = strings.Trim(innerVal, "' ")

					}

				} else {
					val = fmt.Sprintf("%s: %v, ", fieldProperties.Name, value.Interface())
				}

			}
		}

		if val != "" {
			queryString := fmt.Sprintf("%s='%s'", goTag, val)

			setQuery = setQueryIndex(i, fieldNumbers, setQuery, queryString)
		}

	}
	setQuery = strings.TrimSuffix(setQuery, ",")

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
			queryString := fmt.Sprintf("%s='%s'", goTag, val)
			whereQuery += whereQueryIndex(i, fieldNumbers, whereQuery, queryString)

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
