package httpcodec;

import(
	"errors"
	"net/url"
	"reflect"
	"strings"
	"strconv"
)

type URLEncoder struct {
	base url.Values
}

func (u URLEncoder) Encode(i interface{}) error {
	if i == nil {
		return nil
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Map {
		return u.encodeMap(i)
	} else if t.Kind() == reflect.Struct {
		return u.encodeStruct(t, i)
	} else {
		return errors.New("Can only urlencode structs and maps!")
	}
	panic("Should not get here")
}

func valueToString(v reflect.Value) (val string) {
	val = v.String()
	switch (v.Kind()) {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		val = strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		val = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	default:
		if stringConvert := v.MethodByName("String"); stringConvert.IsValid() {
			defer func() {
				recover()
			}()
			retVal := stringConvert.Call([]reflect.Value{})
			if len(retVal) > 0 && retVal[0].Kind() == reflect.String {
				return retVal[0].String()
			}
		}
	}
	return val
}

// Stolen from json package
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (u *URLEncoder) addOneField(key string, v reflect.Value, omitEmpty bool) {
	if omitEmpty && isEmptyValue(v) {
		return
	}
	u.base.Add(key, valueToString(v))
}

func (u *URLEncoder) addField(key string, v reflect.Value, omitEmpty bool) {
	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		for j := 0; j < v.Len(); j++ {
			u.addOneField(key, v.Index(j), omitEmpty)
		}
	} else {
		u.addOneField(key, v, omitEmpty)
	}
}

func (u *URLEncoder) encodeMap(i interface{}) error {
	v := reflect.ValueOf(i)
	for _, k := range(v.MapKeys()) {
		value := v.MapIndex(k)
		u.addField(valueToString(k), value, false)
	}

	return nil
}

const kTagKey = "url"

func (u *URLEncoder) encodeStruct(t reflect.Type, i interface{}) error {
	v := reflect.ValueOf(i)
	for i := 0; i < t.NumField(); i++ {
		omitEmpty := false
		f := t.Field(i)
		urlKey := f.Name
		if tagValue := f.Tag.Get(kTagKey); tagValue != "" {
			params := strings.Split(tagValue, ",")
			urlKey = params[0]
			if len(params) > 0 {
				omitEmpty = params[1] == "omitempty"
			}
		}

		value := v.Field(i)
		u.addField(urlKey, value, omitEmpty)
	}
	return nil
}
