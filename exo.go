package exo

import "reflect"

func ToMap(s interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	fields := StructFields(s)

	for _, field := range fields {
		name := field.Name
		val := toValue(s).FieldByName(name)
		out[name] = val.Interface()
	}

	return out
}

func StructFields(s interface{}) []reflect.StructField {
	t := toValue(s).Type()
	f := make([]reflect.StructField, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		f[i] = field
	}

	return f
}

func toValue(s interface{}) reflect.Value {
	return reflect.ValueOf(s)
}
