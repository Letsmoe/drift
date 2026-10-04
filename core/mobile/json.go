package mobile

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// encode marshals value for the Kotlin side, which decodes lists as non-null:
// nil slices anywhere inside value become [] instead of null.
func encode(value any) (string, error) {
	pointer := reflect.New(reflect.TypeOf(value))
	pointer.Elem().Set(reflect.ValueOf(value))
	fillNilSlices(pointer.Elem())
	encoded, err := json.Marshal(pointer.Elem().Interface())
	if err != nil {
		return "", fmt.Errorf("encode result: %w", err)
	}
	return string(encoded), nil
}

// encodeResult passes err through so callers can return encode(value, err) directly.
func encodeResult(value any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return encode(value)
}

// mustEncode is for values that cannot fail to marshal (plain structs of strings and numbers).
func mustEncode(value any) string {
	encoded, err := encode(value)
	if err != nil {
		return "null"
	}
	return encoded
}

func decode(encoded string, target any) error {
	err := json.Unmarshal([]byte(encoded), target)
	if err != nil {
		return fmt.Errorf("decode argument: %w", err)
	}
	return nil
}

// fillNilSlices walks value and replaces nil slices with empty ones in place.
func fillNilSlices(value reflect.Value) {
	switch value.Kind() {
	case reflect.Slice:
		fillSlice(value)
	case reflect.Struct:
		fillStruct(value)
	case reflect.Map:
		fillMap(value)
	case reflect.Pointer, reflect.Interface:
		if !value.IsNil() {
			fillNilSlices(value.Elem())
		}
	default:
	}
}

func fillSlice(value reflect.Value) {
	if !value.CanSet() {
		return
	}
	if value.IsNil() {
		value.Set(reflect.MakeSlice(value.Type(), 0, 0))
		return
	}
	for index := range value.Len() {
		fillNilSlices(value.Index(index))
	}
}

func fillStruct(value reflect.Value) {
	for index := range value.NumField() {
		field := value.Field(index)
		if field.CanSet() {
			fillNilSlices(field)
		}
	}
}

// fillMap copies each element out, fixes it and stores it back, since map elements are not addressable.
func fillMap(value reflect.Value) {
	for _, key := range value.MapKeys() {
		element := reflect.New(value.Type().Elem()).Elem()
		element.Set(value.MapIndex(key))
		fillNilSlices(element)
		value.SetMapIndex(key, element)
	}
}
