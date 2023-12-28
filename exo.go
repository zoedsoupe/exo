package exo

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type changeset[T struct{}] struct {
	changes map[string]interface{}
	params  map[string]interface{}
	errors  map[string]error
	data    T
	isValid bool
}

func Cast[T struct{}](params map[string]interface{}) changeset[T] {
	var s T
	var c changeset[T]
	c.params = params
	c.data = s
	c.isValid = true
	c.errors = make(map[string]error)

	for _, f := range StructFields(s) {
		field := f.Name

		change, ok := params[field]

		if !ok {
			continue
		}

		c.changes[field] = change
	}

	return c
}

func Apply[T struct{}](c changeset[T]) (T, error) {
	var s T

	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Struct {
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

	r := reflect.ValueOf(&s).Elem()
	for key, value := range c.changes {
		f := r.FieldByName(key)

		if !(f.IsValid() && f.CanSet()) {
			continue
		}

		val := reflect.ValueOf(value)

		if !val.Type().AssignableTo(f.Type()) {
			msg := fmt.Sprintf("type mismatch for field %s, expected %s found %s", key, f.Type().String(), val.Type().String())
			return s, errors.New(msg)
		}

		f.Set(val)
	}

	return s, nil
}

func (c changeset[T]) AddError(field string, err string) changeset[T] {
	c.errors[field] = errors.New(err)
	return c
}

func (c changeset[T]) PutChange(field string, change interface{}) changeset[T] {
	sfs := StructFields(c.data)
	var fields = make([]string, len(sfs))

	for i, f := range sfs {
		fields[i] = f.Name
	}

	for i, f := range fields {
		if field == f {
			val := reflect.ValueOf(change)
			sf := sfs[i]

			if !val.Type().AssignableTo(sf.Type) {
				c.isValid = false
				c.errors[field] = fmt.Errorf("type mismatch for field %s, expected %s found %s", field, sf.Type.String(), val.Type().String())
				return c
			}

			c.changes[field] = change
			return c
		}
	}

	c.isValid = false
	c.errors[field] = fmt.Errorf("%s is invalid", field)
	return c
}

func (c changeset[T]) UpdateChange(field string, cb func(interface{}) interface{}) changeset[T] {
	v := cb(c.changes[field])
	return c.PutChange(field, v)
}

func (c changeset) ValidateRequired(need []string) changeset {
	keys := make([]string, len(c.changes))

	for k := range c.changes {
		keys = append(keys, k)
	}

	slices.Sort[[]string](keys)
	slices.Sort[[]string](need)

	diff := difference(need, keys)

	if len(diff) > 0 {
		c.isValid = false

		for _, key := range diff {
			msg := fmt.Sprintf("%s is required", key)
			c = c.AddError(key, msg)
		}
		return c
	}

	return c
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (c changeset) ValidateAcceptance(field string) changeset {
	isTrue := func(field string, v interface{}) (bool, error) {
		if v == false {
			msg := fmt.Sprintf("%s is not true", field)
			return false, errors.New(msg)
		}

		return true, nil
	}

	return c.ValidateChange(field, isTrue)
}

func (c changeset) ValidateFormat(field string, re *regexp.Regexp) changeset {
	hasFormat := func(field string, curr interface{}) (bool, error) {
		switch curr.(type) {
		case string:
			if re.FindString(curr.(string)) == "" {
				return false, errors.New("password must match alphabetic chars")
			}

			return true, nil

		default:
			return false, errors.New("Field isn't a string")
		}
	}

	return c.ValidateChange(field, hasFormat)
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

func (c changeset[T]) GetChange(field string) (interface{}, bool) {
	v, ok := c.changes[field]

	return v, ok
}

func (c changeset[T]) GetChanges() map[string]interface{} {
	return c.changes
}

func (c changeset[T]) GetParams() map[string]interface{} {
	return c.params
}

func (c changeset[T]) GetErrors() map[string]error {
	return c.errors
}

func (c changeset[T]) GetError(field string) error {
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
