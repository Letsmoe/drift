package main

import (
	"reflect"
	"regexp"
	"strings"
)

// jsonField is one property of a struct as encoding/json writes it.
type jsonField struct {
	name     string
	goType   reflect.Type
	optional bool
	asString bool
}

var plainIdentifier = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

// declaredName quotes names TypeScript does not accept bare, and marks optional ones.
func (field jsonField) declaredName() string {
	name := field.name
	if !plainIdentifier.MatchString(name) {
		name = "'" + name + "'"
	}
	if field.optional {
		return name + "?"
	}
	return name
}

// jsonFields lists goType's JSON properties in declaration order. Fields of
// untagged embedded structs are flattened; a field declared directly wins over
// an embedded one of the same name, as in encoding/json.
func jsonFields(goType reflect.Type) []jsonField {
	direct := make(map[string]bool)
	for index := range goType.NumField() {
		field, include := directField(goType.Field(index))
		if include {
			direct[field.name] = true
		}
	}

	var fields []jsonField
	for index := range goType.NumField() {
		structField := goType.Field(index)
		if embedded := embeddedStruct(structField); embedded != nil {
			fields = append(fields, promotedFields(embedded, direct)...)
			continue
		}
		field, include := directField(structField)
		if include {
			fields = append(fields, field)
		}
	}
	return fields
}

func promotedFields(embedded reflect.Type, direct map[string]bool) []jsonField {
	var fields []jsonField
	for _, field := range jsonFields(embedded) {
		if !direct[field.name] {
			fields = append(fields, field)
		}
	}
	return fields
}

// embeddedStruct is the struct type an untagged embedded field promotes, or nil.
func embeddedStruct(structField reflect.StructField) reflect.Type {
	if !structField.Anonymous {
		return nil
	}
	name, _, _ := strings.Cut(structField.Tag.Get("json"), ",")
	if name != "" {
		return nil
	}
	fieldType := structField.Type
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() != reflect.Struct {
		return nil
	}
	return fieldType
}

func directField(structField reflect.StructField) (jsonField, bool) {
	if embeddedStruct(structField) != nil || !structField.IsExported() {
		return jsonField{}, false
	}
	tag := structField.Tag.Get("json")
	if tag == "-" {
		return jsonField{}, false
	}
	name, options, _ := strings.Cut(tag, ",")
	if name == "" {
		name = structField.Name
	}
	return jsonField{
		name:     name,
		goType:   structField.Type,
		optional: hasOption(options, "omitempty") || hasOption(options, "omitzero"),
		asString: hasOption(options, "string"),
	}, true
}

func hasOption(options string, option string) bool {
	for _, candidate := range strings.Split(options, ",") {
		if candidate == option {
			return true
		}
	}
	return false
}
