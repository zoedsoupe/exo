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

type Validator interface {
	Validate(field string, value interface{}) (bool, error)
}

type LengthValidator struct {
	Min int
	Max int
}

func (lv LengthValidator) Validate(field string, v interface{}) (bool, error) {
	var l int
	t := reflect.TypeOf(v).String()
	var msg string

	switch t {
	case "string":
		l = len(v.(string))
		msg = "%s should be %s %d characters"
	case "slice":
		l = len(v.([]interface{}))
		msg = "%s should have %s %d items"
	case "map":
		l = len(v.(map[interface{}]interface{}))
		msg = "%s should have %s %d elements"
	default:
		l = -1
	}

	if lv.Min == lv.Max && l != lv.Min {
		return false, fmt.Errorf(msg, field, "", lv.Min)
	}

	if l < lv.Min {
		return false, fmt.Errorf(msg, field, "at least", lv.Min)
	}

	if l > lv.Max {
		return false, fmt.Errorf(msg, field, "at most", lv.Max)
	}

	return true, nil
}

type FormatValidator struct {
	Pattern *regexp.Regexp
}

func (fv FormatValidator) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(string)

	if !ok {
		return false, fmt.Errorf("%s is not a string", field)
	}

	if fv.Pattern.FindString(v) == "" {
		return false, fmt.Errorf("%s has invalid format", field)
	}

	return true, nil
}

type AcceptanceValidator struct{}

func (av AcceptanceValidator) Validate(field string, val interface{}) (bool, error) {
	accepted, ok := val.(bool)

	if !ok {
		return false, fmt.Errorf("%s isn't a boolean", field)
	}

	if !accepted {
		return false, fmt.Errorf("%s must be accepted", field)
	}

	return true, nil
}

type ExclusionValidator struct {
	Disallowed []interface{}
}

func (ev ExclusionValidator) Validate(field string, value interface{}) (bool, error) {
	for _, disallowed := range ev.Disallowed {
		if !reflect.DeepEqual(value, disallowed) {
			return true, nil
		}
	}

	return false, fmt.Errorf("%s is reserved", field)
}

type InclusionValidator struct {
	Allowed []interface{}
}

func (iv InclusionValidator) Validate(field string, value interface{}) (bool, error) {
	for _, allowed := range iv.Allowed {
		if reflect.DeepEqual(value, allowed) {
			return true, nil
		}
	}

	return false, fmt.Errorf("%s is invalid", field)
}

type Number interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64
}

type LessThanValidator[T Number] struct {
	MaxValue T
}

func (ltv LessThanValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v := val.(T)

	if v > ltv.MaxValue {
		return false, fmt.Errorf("%s must be less than %v", field, v)
	}

	return true, nil
}

type LessThanOrEqualValidator[T Number] struct {
	MaxValue T
}

func (ltv LessThanOrEqualValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v := val.(T)

	if v >= ltv.MaxValue {
		return false, fmt.Errorf("%s must be less than or equal to %v", field, v)
	}

	return true, nil
}

type GreaterThanValidator[T Number] struct {
	MinValue T
}

func (gtv GreaterThanValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v := val.(T)

	if v < gtv.MinValue {
		return false, fmt.Errorf("%s must be greater than %v", field, v)
	}

	return true, nil
}

type GreaterThanOrEqualValidator[T Number] struct {
	MinValue T
}

func (gtv GreaterThanOrEqualValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v := val.(T)

	if v <= gtv.MinValue {
		return false, fmt.Errorf("%s must be greater than or equal to %v", field, v)
	}

	return true, nil
}

type EqualToValidator[T Number] struct {
	Value T
}

func (ev EqualToValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v := val.(T)

	if v == ev.Value {
		return true, nil
	}

	return false, fmt.Errorf("%s must be equal to %v", field, v)
}

type NotEqualToValidator[T Number] struct {
	Value T
}

func (nev NotEqualToValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v := val.(T)

	if v != nev.Value {
		return true, nil
	}

	return false, fmt.Errorf("%s must be not equal to %v", field, v)
}

func (c changeset[T]) ValidateRequired(need []string) changeset[T] {
	for _, field := range need {
		fieldValue, exists := c.changes[field]

		if !exists || !reflect.ValueOf(fieldValue).IsValid() {
			c.isValid = false
			c.errors[field] = fmt.Errorf("%s is required", field)
			return c
		}
	}

	return c
}

func (c changeset[T]) ValidateChange(field string, v Validator) changeset[T] {
	val, ok := c.GetChange(field)

	if !ok {
		msg := fmt.Sprintf("%s doesn't exist", field)
		c.errors[field] = errors.New(msg)
		c.isValid = false
		return c
	}

	if ok, error := v.Validate(field, val); !ok {
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
