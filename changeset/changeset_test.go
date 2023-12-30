package changeset_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/zoedsoupe/exo/changeset"
)

type T struct {
	A string
	B int
}

func TestPutChange(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)

	c = c.PutChange("A", "oi")

	if a, ok := c.GetChange("A"); ok && a != "oi" {
		t.Errorf("PutChange should overwrite the current value")
	}

	c = c.PutChange("B", "ixe")
}

func TestAddError(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)

	c = c.AddError("A", "WE HAVE AN ERROR")

	err := c.GetError("A")

	if !(err != nil) {
		t.Errorf("AddError should return an error on a existing key")
	}
}

func TestValidateRequired(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)

	c = c.ValidateRequired([]string{"A", "B"})

	if c.IsValid {
		t.Errorf("ValidateRequired should add error on non existing keys")
	}

	err := c.GetError("B")

	if !(err != nil) {
		t.Errorf("ValidateRequired should return an error on a existing key")
	}
}

func TestUpdateChange(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)

	c = c.UpdateChange("A", func(a interface{}) interface{} {
		return "foo"
	})

	if a, ok := c.GetChange("A"); ok && a == "hello" {
		t.Errorf("UpdateChange should modify the current change value on specified key")
	}

}

func TestValidateAcceptance(t *testing.T) {
	attrs := map[string]interface{}{"A": false}
	c := changeset.Cast[T](attrs)

	c = c.ValidateChange("A", changeset.AcceptanceValidator{})

	if c.IsValid {
		t.Errorf("ValidateAcceptance should add error on non tru keys")
	}

	err := c.GetError("A")

	if !(err != nil) {
		t.Errorf("ValidateAcceptance should return an error on a non true key")
	}
}

func TestValidateFormat(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)

	re := regexp.MustCompile("hello")
	c = c.ValidateChange("A", changeset.FormatValidator{Pattern: re})

	if !c.IsValid {
		t.Errorf("ValidateFormat shouldn't add error on a valid format keys")
	}

	err := c.GetError("A")

	if err != nil {
		t.Errorf("ValidateFormat shouldn't return an error on a valid format key")
	}
}

func TestApply(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)
	r, err := changeset.Apply[T](c)

	if err != nil {
		t.Errorf("Apply should returna a valid T struct")
	}

	if r.A != "hello" {
		t.Errorf("Apply should return the struct with updates")
	}
}

type R struct{ A int }

func TestValidateLength(t *testing.T) {
	attrs := map[string]interface{}{"A": "hello"}
	c := changeset.Cast[T](attrs)

	lv1 := changeset.LengthValidator{Min: 5, Max: 5}
	if c := c.ValidateChange("A", lv1); !c.IsValid {
		t.Errorf("ValidateChange should return truthy on correct length")
	}

	lv2 := changeset.LengthValidator{Min: 10, Max: 10}
	if c := c.ValidateChange("A", lv2); c.IsValid {
		t.Errorf("ValidateChange should return falsy on wrong length")
	}

	attrs = map[string]interface{}{"A": 20}
	c2 := changeset.Cast[R](attrs)

	lvc2 := changeset.LengthValidator{Min: 5, Max: 5}
	if c2 := c2.ValidateChange("A", lvc2); c2.IsValid {
		t.Errorf("ValidateChange should return error on non string and slices types")
	}
}

func TestCast(t *testing.T) {
	attrs := map[string]interface{}{"foo": 123, "A": "hello", "B": 2}
	c := changeset.Cast[T](attrs)

	if p := c.GetParams(); !reflect.DeepEqual(p, attrs) {
		t.Errorf("Cast should parse params as raw attrs, got: %v", p)
	}

	if _, f := c.GetChange("foo"); f {
		t.Errorf("Cast should't return the 'foo' key")
	}

	if _, f := c.GetChange("A"); !f {
		t.Errorf("Cast should return the 'A' key")
	}

	if _, f := c.GetChange("B"); !f {
		t.Errorf("Cast should return the 'B' key")
	}
}

func TestGetChange(t *testing.T) {
	attrs := map[string]interface{}{"foo": 123, "A": "hello", "B": 2}
	c := changeset.Cast[T](attrs)

	if v, ok := c.GetChange("A"); !ok || !reflect.DeepEqual(v, "hello") {
		t.Errorf("GetChange should return the correct value from key, got: %v", v)
	}

	if v, ok := c.GetChange("B"); !ok || !reflect.DeepEqual(v, 2) {
		t.Errorf("GetChange should return the correct value from key, got: %v", v)
	}
}
