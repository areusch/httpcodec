package httpcodec;

import(
	"errors"
	"net/url"
	"reflect"
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

func valueToString(v reflect.Value) string {
	switch (v.Kind()) {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return v.String()
}

func (u *URLEncoder) addField(key string, v reflect.Value) {
	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		for j := 0; j < v.Len(); j++ {
			u.base.Add(key, valueToString(v.Index(j)))
		}
	} else {
		u.base.Add(key, valueToString(v))
	}
}

func (u *URLEncoder) encodeMap(i interface{}) error {
	v := reflect.ValueOf(i)
	for _, k := range(v.MapKeys()) {
		value := v.MapIndex(k)
		u.addField(valueToString(k), value)
	}

	return nil
}

const kTagKey = "url"

func (u *URLEncoder) encodeStruct(t reflect.Type, i interface{}) error {
	v := reflect.ValueOf(i)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		urlKey := f.Name
		if tagValue := f.Tag.Get(kTagKey); tagValue != "" {
			urlKey = tagValue
		}

		value := v.Field(i)
		u.addField(urlKey, value)
	}
	return nil
}
