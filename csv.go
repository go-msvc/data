package data

import (
	"fmt"
	"reflect"

	"github.com/go-msvc/errors"
)

func CSV(data interface{}) ([]string, error) {
	return csv(reflect.ValueOf(data))
}

func csv(value reflect.Value) ([]string, error) {
	switch value.Kind() {
	//simple scalar types
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.String, reflect.Bool:
		return []string{fmt.Sprintf("%v", value.Interface())}, nil
	case reflect.Struct:
		return structCSV(value)
	case reflect.Slice:
		return sliceCSV(value)
	default:
		v := value.Interface()
		if s, ok := v.(string); ok && s == "null" {
			return []string{""}, nil //this is where JSON unmarshalled (string)"null" into struct field of type interface{}
		}
		if v == nil {
			return []string{""}, nil //ensure we do not get "<nil>" from %s
		}
		//other types are simply printed using "%s" so any custom scalar type
		//with a String() method will be printed
		return []string{fmt.Sprintf("%s", v)}, nil
	}
} //csv()

func structCSV(value reflect.Value) ([]string, error) {
	values := []string{}
	t := value.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		csvValues, err := csv(value.Field(i))
		if err != nil {
			return values, errors.Wrapf(err, "failed on struct field(%s)", f.Name)
		}
		values = append(values, csvValues...)
	}
	return values, nil
} //structCSV()

func sliceCSV(value reflect.Value) ([]string, error) {
	values := []string{}
	for i := 0; i < value.Len(); i++ {
		csvValues, err := csv(value.Index(i))
		if err != nil {
			return values, errors.Wrapf(err, "failed on slice[%d]", i)
		}
		values = append(values, csvValues...)
	}
	return values, nil
} //sliceCSV()
