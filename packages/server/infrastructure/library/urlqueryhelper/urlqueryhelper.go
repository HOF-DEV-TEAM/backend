package urlqueryhelper

import (
	"errors"
	"net/http"
	"reflect"
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
