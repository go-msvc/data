package data

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

// Get() and all its variants split names on [] for slices and . for maps
// so you cannot use it with a map containing this:
//
//	map[string]interface{}{
//		"test.young": true,
//		"system[4]":false,
//	}
//
// because both those names will be split into "test" and "system" which will
// not be found in the map...
//
// it also only supports map[string]... and not any other type of index on a map
// which is typically used in JSON data.
//
// data generally is either a map or a slice, but could also be a scalar
// value when name is ""
func Get(data interface{}, name string) (interface{}, error) {
	v, err := get(reflect.ValueOf(data), nameParts(name))
	if err != nil {
		log.Errorf("get(%s) failed: %+v", name, err)
	} else {
		log.Debugf("get(%s) -> (%T)%+v", name, v, v)
	}
	return v, err
}

// GetInto() does Get() and
//
//	then makes a copy of tmplValue and JSON parse the value over that
//		(i.e. merge maps with existing maps, overwrite others)
//	then validates the final value (if tmplValue is a Validator)
//	then returns the valid value
func GetInto(data interface{}, name string, tmplValue interface{}) (interface{}, error) {
	//get the specified value
	namedValue, err := Get(data, name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get")
	}
	//marshal named value to JSON
	jsonNamedValue, err := json.Marshal(namedValue)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal to JSON")
	}
	return JsonInto(jsonNamedValue, tmplValue)
}

func JsonInto(jsonValue []byte, tmplValue interface{}) (interface{}, error) {
	//make new copy of defaultValue
	outPtrValue := reflect.New(reflect.TypeOf(tmplValue))
	outPtrValue.Elem().Set(reflect.ValueOf(tmplValue))

	//unmarshal over the tmplValue to merge
	if err := json.Unmarshal(jsonValue, outPtrValue.Interface()); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal named JSON value")
	}

	//validate if implemented in tmplValue type
	if validtor, ok := outPtrValue.Interface().(Validator); ok {
		if err := validtor.Validate(); err != nil {
			return nil, errors.Wrapf(err, "invalid value")
		}
	}
	return outPtrValue.Elem().Interface(), nil
} //GetInto()

// GetOr() does GetInto() and on any error, returns the defaultValue
// it is useful if you want to use the value regardless of wrong values in data
func GetOr(data interface{}, name string, defaultValue interface{}) interface{} {
	v, err := GetInto(data, name, defaultValue)
	if err == nil {
		return v
	}
	return defaultValue
} //GetOr()

// func GetAll(data interface{}, names []string) (map[string]interface{}, error) {
// 	return nil, errors.Errorf("NYI")
// }

func get(value reflect.Value, nameParts []string) (interface{}, error) {
	log.Debugf("get(%v, name=%s)", value.Type(), strings.Join(nameParts, "|"))
	switch value.Kind() {
	//simple scalar types
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.String, reflect.Bool:
		if len(nameParts) > 0 {
			return nil, errors.Errorf("cannot get(%+v) from %v", nameParts, value.Kind())
		}
		return value.Interface(), nil
	case reflect.Interface:
		if len(nameParts) == 0 {
			return value.Interface(), nil
		}
		if obj, ok := value.Interface().(map[string]interface{}); ok {
			return get(reflect.ValueOf(obj), nameParts)
		}
		if arr, ok := value.Interface().([]interface{}); ok {
			return get(reflect.ValueOf(arr), nameParts)
		}
		return nil, errors.Errorf("cannot get(%+v) from %v", nameParts, value.Kind())
	case reflect.Struct:
		return getFromStruct(value, nameParts)
	case reflect.Map:
		return getFromMap(value, nameParts)
	case reflect.Slice:
		return getFromSlice(value, nameParts)
	default:
		return nil, errors.Errorf("cannot get(%+v) from %v", nameParts, value.Kind())
	}
}

func getFromStruct(value reflect.Value, nameParts []string) (interface{}, error) {
	structType := value.Type()
	log.Debugf("getFromStruct(%v, name=%s)", structType, strings.Join(nameParts, "|"))
	if len(nameParts) < 1 {
		return value.Interface(), nil //no more names, return the whole struct value
	}
	//if value is nil, cannot go into struct
	if !value.IsValid() {
		return nil, errors.Errorf("no value to get(%s)", nameParts[0])
	}

	//try to find by struct field name first
	if structField, ok := structType.FieldByName(nameParts[0]); ok {
		//found by struct field name
		return get(value.FieldByIndex(structField.Index), nameParts[1:])
	}

	//try to find by json tag
	numFields := structType.NumField()
	for i := 0; i < numFields; i++ {
		jsonTag := strings.SplitN(structType.Field(i).Tag.Get("json"), ",", 2)[0]
		if jsonTag == nameParts[0] {
			return get(value.Field(i), nameParts[1:]) //matched json tag
		}
	}
	return nil, errors.Errorf("struct %v.%s does not exist", structType, nameParts[0])
} //getFromStruct()

func getFromMap(value reflect.Value, nameParts []string) (interface{}, error) {
	log.Debugf("getFromMap(%v, name=%s)", value.Type(), strings.Join(nameParts, "|"))
	if len(nameParts) < 1 {
		return value.Interface(), nil //no more names, return the whole map value
	}
	//cannot get if map is not keyed with string values
	if value.Type().Key().Kind() != reflect.String {
		return nil, errors.Errorf("map key is %v != %v", value.Type().Key().Kind(), reflect.String)
	}
	//if value is nil, cannot go into map
	if value.IsNil() {
		return nil, errors.Errorf("no value to get(%s)", nameParts[0])
	}
	//get the named item
	elemValue := value.MapIndex(reflect.ValueOf(nameParts[0]))
	if !elemValue.IsValid() {
		return nil, errors.Errorf("not found(%s)", nameParts[0])
	}
	return get(elemValue, nameParts[1:])
} //getFromMap()

func getFromSlice(value reflect.Value, nameParts []string) (interface{}, error) {
	log.Debugf("getFromSlice(%v, name=%s)", value.Type(), strings.Join(nameParts, "|"))
	if len(nameParts) < 1 {
		return value.Interface(), nil //no more names, return the whole slice value
	}
	//if value is nil, cannot go into slice
	if value.IsNil() {
		return nil, errors.Errorf("no value to get(%s)", nameParts[0])
	}
	//key must be integer value in range of slice len
	if i64, err := strconv.ParseInt(nameParts[0], 10, 64); err != nil {
		return nil, errors.Errorf("get(%s) is not integer slice index", nameParts[0])
	} else {
		if i64 < 0 || i64 >= int64(value.Len()) {
			return nil, errors.Errorf("get(%s) is out of slice range 0..%d", nameParts[0], value.Len()-1)
		}
		//get the index slice elem
		elemValue := value.Index(int(i64))
		if !elemValue.IsValid() {
			return nil, errors.Errorf("not found(%s)", nameParts[0])
		}
		return get(elemValue, nameParts[1:])
	}
} //getFromSlice()

// split jq formatted name into parts
// e.g. "system[6].port" -> []string{"system", "6", "port"}
func nameParts(name string) []string {
	f := func(c rune) bool {
		return c == '.' || c == '[' || c == ']'
	}
	log.Debugf("name(%s) -> \"%s\"", name, strings.Join(strings.FieldsFunc(name, f), "|"))
	return strings.FieldsFunc(name, f)
}
