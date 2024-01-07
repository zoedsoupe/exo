// changeset is an attempt to port the Elixir's library
// for data structure validations and perform changes
// in a lazy and reactive structure called Changeset.
package changeset

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/zoedsoupe/exo"
)

// `Changeset[T]` represents the structure to
// perform lazy validations and changes into
// an already defined data structure.
// It holds the `changes` normally called `attrs`
// it `errors` for each field and a field that
// you can always check if a `Changeset[T]` is valid.
type Changeset[T interface{}] struct {
	changes     map[string]interface{}
	params      map[string]interface{}
	errors      map[string]error
	validations map[string]Validator
	data        T
	IsValid     bool
}

func (c *Changeset[T]) Error() string {
	var out strings.Builder

	out.WriteString("Changeset has errors:\n\t")

	for field, err := range c.GetErrors() {
		msg := fmt.Sprintf("%s: %s\n\t", field, err)
		out.WriteString(msg)
	}

	return out.String()
}

// Convenience to transform a `Changeset[T]` to
// a String JSON ready to be sent as HTTP server
// response.
func (c *Changeset[T]) ErrorJSON() map[string]string {
	var final = make(map[string]string)
	for field, err := range c.GetErrors() {
		final[field] = err.Error()
	}

	return final
}

// Fiven a data type and a map of attributes, filter
// parameters that exists as field on the data type.
// If the value of the parameter mismatch the data type field,
// an error is added to the Changeset and it is amrked as invalid.
func Cast[T interface{}](params map[string]interface{}) Changeset[T] {
	var s T

	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("argument is not a struct"))
	}

	var c Changeset[T]
	c.params = params
	c.data = s
	c.IsValid = true
	c.errors = make(map[string]error)
	c.validations = make(map[string]Validator)
	c.changes = make(map[string]interface{})

	for _, f := range exo.StructFields(s) {
		field := f.Name
		change, ok := params[field]
		if !ok {
			continue
		}

		sType := f.Type.String()
		cType := reflect.TypeOf(change).String()
		if cType != sType {
			c.IsValid = false
			msg := fmt.Errorf("type mismatch: expect %s got %s", sType, cType)
			c.AddError(field, msg)
		} else {
			c.changes[field] = change
		}
	}

	return c
}

// Same as Apply but handle a new instance of the desired
// data structure.
func ApplyNew[T interface{}](c Changeset[T]) (T, error) {
	var s = c.data
	err := Apply(&s, c)
	return s, err
}

// Given an already existence instance of the data type used
// to generate the Changeset as a pointer, and the Changeset
// it self, apply all changes to the instance.
// Note that this function panic if given an invalid data type.
func Apply[T interface{}](s *T, c Changeset[T]) error {
	t := reflect.ValueOf(s)
	if t.Kind() != reflect.Ptr {
		panic(fmt.Errorf("argument to Apply is not a pointer to a struct"))
	}
	if t.Elem().Type().Kind() != reflect.Struct {
		panic(fmt.Errorf("argument to Apply is not a struct"))
	}

	if !c.IsValid {
		return &c
	}

	r := reflect.ValueOf(s).Elem()
	for key, value := range c.changes {
		f := r.FieldByName(key)
		if !(f.IsValid() && f.CanSet()) {
			continue
		}

		val := reflect.ValueOf(value)
		if !val.Type().AssignableTo(f.Type()) {
			msg := fmt.Errorf("type mismatch expected %s got %s", key, val.Type().String())
			c.AddError(key, msg)
			return &c
		}

		f.Set(val)
	}

	return nil
}

// Adds a new error on the given field. Note that if
// already exists an error on the given field, it will
// be overwritten.
func (c Changeset[T]) AddError(field string, err error) Changeset[T] {
	c.errors[field] = err
	return c
}

// Writes a change into the given field. The only validation that
// is made is the type matching for the given data type field.
// Note that if a change is already present of the changes map,
// it will be overwritten.
// This function is more suited for internal usage into an application.
func (c Changeset[T]) PutChange(field string, change interface{}) Changeset[T] {
	sfs := exo.StructFields(c.data)
	var fields = make([]string, len(sfs))

	for i, f := range sfs {
		fields[i] = f.Name
	}

	for i, f := range fields {
		if field == f {
			val := reflect.ValueOf(change)
			sf := sfs[i]

			if !val.Type().AssignableTo(sf.Type) {
				c.IsValid = false
				c.errors[field] = fmt.Errorf("type mismatch, expected %s got %s", sf.Type.String(), val.Type().String())
				return c
			}

			c.changes[field] = change
			return c
		}
	}

	c.IsValid = false
	c.errors[field] = fmt.Errorf("%s is invalid", field)
	return c
}

// Given an callback receives the current change and would
// return a change and an optional error, updates or transform
// the change for the given field.
// Note that the current change can be possibly a zero value
// or `nil`. Also note that this function will behave like
// `PutChange`, only make type assertions for fields and no
// additional validation.
func (c Changeset[T]) UpdateChange(field string, cb func(interface{}) (interface{}, error)) Changeset[T] {
	v, err := cb(c.changes[field])
	if err != nil {
		c.AddError(field, err)
		c.IsValid = false
		return c
	}
	return c.PutChange(field, v)
}

// Interface to define custom validations for changesets.
// Check `ValidateChange` for more information.
type Validator interface {
	Validate(field string, value interface{}) (bool, error)
}

// Validates that a given change has the desired length.
// It works on string, map and slice types.
// If you want an **exact** length, give the `Min` and `Max`
// the same value.
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
		msg = "should be %s %d characters"
	case "slice":
		l = len(v.([]interface{}))
		msg = "should have %s %d items"
	case "map":
		l = len(v.(map[interface{}]interface{}))
		msg = "should have %s %d elements"
	default:
		l = -1
	}

	if lv.Min == lv.Max && l != lv.Min {
		return false, fmt.Errorf(msg, "", lv.Min)
	}

	if l < lv.Min {
		return false, fmt.Errorf(msg, "at least", lv.Min)
	}

	if l > lv.Max {
		return false, fmt.Errorf(msg, "at most", lv.Max)
	}

	return true, nil
}

// Validates if a string field would match the given Regexp pattern.
type FormatValidator struct {
	Pattern *regexp.Regexp
}

func (fv FormatValidator) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(string)

	if !ok {
		return false, fmt.Errorf("is not a string")
	}

	if fv.Pattern.FindString(v) == "" {
		return false, fmt.Errorf("has invalid format")
	}

	return true, nil
}

// Validates if a boolean field is true.
type AcceptanceValidator struct{}

func (av AcceptanceValidator) Validate(field string, val interface{}) (bool, error) {
	accepted, ok := val.(bool)

	if !ok {
		return false, fmt.Errorf("isn't a boolean")
	}

	if !accepted {
		return false, fmt.Errorf("must be accepted")
	}

	return true, nil
}

// Given a list of disallowed values, validates if the
// value of a field isn't included into this list.
type ExclusionValidator struct {
	Disallowed []interface{}
}

func (ev ExclusionValidator) Validate(field string, value interface{}) (bool, error) {
	for _, disallowed := range ev.Disallowed {
		if !reflect.DeepEqual(value, disallowed) {
			return true, nil
		}
	}

	return false, fmt.Errorf("is reserved")
}

// Given a slice of desired values, validates if the
// value of a field is included on this slice.
// It can act like a type of "Enum".
type InclusionValidator struct {
	Allowed []interface{}
}

func (iv InclusionValidator) Validate(field string, value interface{}) (bool, error) {
	for _, allowed := range iv.Allowed {
		if reflect.DeepEqual(value, allowed) {
			return true, nil
		}
	}

	return false, errors.New("is invalid")
}

// Interface to define a possible numeric value.
type Number interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64
}

// Validates that a `Number` is less than a given a max value.
type LessThanValidator[T Number] struct {
	MaxValue T
}

func (ltv LessThanValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(T)

	if !ok {
		return false, fmt.Errorf("isn't a Number %v", v)
	}

	if v > ltv.MaxValue {
		return false, fmt.Errorf("must be less than %v", v)
	}

	return true, nil
}

// Validates if a `Number` field is less than or equal to a max value.
type LessThanOrEqualValidator[T Number] struct {
	MaxValue T
}

func (ltv LessThanOrEqualValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(T)

	if !ok {
		return false, fmt.Errorf("isn't a Number %v", v)
	}

	if v >= ltv.MaxValue {
		return false, fmt.Errorf("must be less than or equal to %v", v)
	}

	return true, nil
}

// Validates that a `Number` field is greater than a given minimal value.
type GreaterThanValidator[T Number] struct {
	MinValue T
}

func (gtv GreaterThanValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(T)

	if !ok {
		return false, fmt.Errorf("isn't a Number %v", v)
	}

	if v < gtv.MinValue {
		return false, fmt.Errorf("must be greater than %v", v)
	}

	return true, nil
}

// Validates if a `Number` field is greater than or equal to a given minimal value.
type GreaterThanOrEqualValidator[T Number] struct {
	MinValue T
}

func (gtv GreaterThanOrEqualValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(T)

	if !ok {
		return false, fmt.Errorf("isn't a Number %v", v)
	}

	if v <= gtv.MinValue {
		return false, fmt.Errorf("must be greater than or equal to %v", v)
	}

	return true, nil
}

// Validates if a `Number` value is equal to a given exact value.
type EqualToValidator[T Number] struct {
	Value T
}

func (ev EqualToValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(T)

	if !ok {
		return false, fmt.Errorf("isn't a Number %v", v)
	}

	if v == ev.Value {
		return true, nil
	}

	return false, fmt.Errorf("must be equal to %v", v)
}

// Validates if a `Number` value is different to a given exact value.
type NotEqualToValidator[T Number] struct {
	Value T
}

func (nev NotEqualToValidator[T]) Validate(field string, val interface{}) (bool, error) {
	v, ok := val.(T)

	if !ok {
		return false, fmt.Errorf("isn't a Number %v", v)
	}

	if v != nev.Value {
		return true, nil
	}

	return false, fmt.Errorf("must be not equal to %v", v)
}

// Given a slice of fields names, validates if all of them
// are present on the `changes` Changeset field, ensuring
// their existence.
func (c Changeset[T]) ValidateRequired(need []string) Changeset[T] {
	for _, field := range need {
		fieldValue, exists := c.changes[field]

		if !exists || !reflect.ValueOf(fieldValue).IsValid() {
			c.IsValid = false
			c.errors[field] = errors.New("is required")
		}
	}

	return c
}

// Given a field and a instance of a `Validator`, apply the
// validation on the changeset and if any error is present,
// add it to the `errors` Changeset field, marking it as invalid.
func (c Changeset[T]) ValidateChange(field string, v Validator) Changeset[T] {
	val, ok := c.GetChange(field)
	c.validations[field] = v

	if !ok {
		c.errors[field] = errors.New("doesn't exist")
		c.IsValid = false
		return c
	}

	if ok, error := v.Validate(field, val); !ok {
		c.errors[field] = error
		c.IsValid = false
		return c
	}

	return c
}

// Get the value of a `changes` entry.
func (c Changeset[T]) GetChange(field string) (interface{}, bool) {
	v, ok := c.changes[field]

	return v, ok
}

// Return all current changes that may be applied to the Changeset.
func (c Changeset[T]) GetChanges() map[string]interface{} {
	return c.changes
}

// Return the raw map that was gaved to `Cast`.
func (c Changeset[T]) GetParams() map[string]interface{} {
	return c.params
}

// Return a map of fields and their errors.
func (c Changeset[T]) GetErrors() map[string]error {
	return c.errors
}

// Return a specific error for a field.
func (c Changeset[T]) GetError(field string) error {
	return c.errors[field]
}

// Applies a callback on each error and return a map
// of fields and the transformed errors.
// The callback will receive a reference to the changeset
// the current error and the `Validator` that it failed.
func (c Changeset[T]) TraverseErrors(cb func(*Changeset[T], error, Validator) interface{}) map[string]interface{} {
	var result = make(map[string]interface{}, len(c.errors))

	for field, err := range c.errors {
		final := cb(&c, err, c.validations[field])
		result[field] = final
	}

	return result
}

// Return a map of fields and their applied `Validators`.
func (c Changeset[T]) Validations() map[string]Validator {
	return c.validations
}

// Check if a given field is present of the `changes`
// Changeset field and returns a boolean of the result.
// The behaviour is similar to `ValidateRequired`
// although it only operates for a single field.
// This is useful when performing complex validations that are
// not possible with `ValidateRequired`.
// For example, evaluating whether at least one field from
// a list is present or evaluating that exactly one field from a list is present.
func (c Changeset[T]) IsFieldMissing(field string) bool {
	curr, exists := c.changes[field]

	if !exists || !reflect.ValueOf(curr).IsValid() {
		return true
	}

	return false
}
