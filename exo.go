package exo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type changeset struct {
	changes map[string]interface{}
	params  map[string]interface{}
	errors  map[string]error
	data    interface{}
	isValid bool
}

func New(data interface{}, params map[string]interface{}) changeset {
	var c changeset
	c.params = params
	c.data = data
	c.isValid = true
	c.errors = make(map[string]error)
	return c
}

func (c changeset) Cast(fields []string) changeset {
	if len(c.params) < 1 {
		return c
	}

	c.changes = c.filterFields(fields)
	return c
}

func (c changeset) filterFields(fields []string) map[string]interface{} {
	out := make(map[string]interface{})

	for k, v := range c.params {
		if !slices.Contains(fields, k) {
			continue
		}
		out[k] = v
	}

	return out
}

func Apply[T any](c changeset) (T, error) {
	var s T

	v := reflect.TypeOf(s)
	if v.Kind() != reflect.Struct {
		return s, fmt.Errorf("argument is not a struct")
	}

	if !c.isValid {
		var msg strings.Builder
		msg.WriteString("changeset has errors:\n\t")
		for k, v := range c.errors {
			error := fmt.Sprintf("%s: %s\n", k, v)
			msg.WriteString(error)
		}
		return s, errors.New(msg.String())
	}

	dict, err := json.Marshal(c.changes)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(dict, &s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c changeset) ValidateLength(field string, length int) changeset {
	sameLength := func(field string, curr interface{}) (bool, error) {
		switch c := curr.(type) {
		case string:
			c = curr.(string)
			if l := len(c); !(l == length) {
				msg := fmt.Sprintf("current length is %d, expected %d", l, length)
				return false, errors.New(msg)
			}

			return true, nil

		case []interface{}:
			c = curr.([]interface{})
			if l := len(c); !(l == length) {
				msg := fmt.Sprintf("current length is %d, expected %d", l, length)
				return false, errors.New(msg)
			}

			return true, nil
		default:
			return false, errors.New("Field isn't a string or slice")
		}
	}

	return c.ValidateChange(field, sameLength)
}

func (c changeset) ValidateChange(field string, validator func(string, interface{}) (bool, error)) changeset {
	val, ok := c.GetChange(field)

	if !ok {
		msg := fmt.Sprintf("%s doesn't exist", field)
		c.errors[field] = errors.New(msg)
		c.isValid = false
		return c
	}

	if ok, error := validator(field, val); !ok {
		c.errors[field] = error
		c.isValid = false
		return c
	}

	return c
}

func (c changeset) GetChange(field string) (interface{}, bool) {
	v, ok := c.changes[field]

	return v, ok
}

func (c changeset) GetChanges() map[string]interface{} {
	return c.changes
}

func (c changeset) GetParams() map[string]interface{} {
	return c.params
}

func (c changeset) GetErrors() map[string]error {
	return c.errors
}

func (c changeset) GetError(field string) error {
	return c.errors[field]
}

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
