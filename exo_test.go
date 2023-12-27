package exo

import (
	"reflect"
	"regexp"
	"testing"
)

func inMap(a map[string]interface{}, val interface{}) bool {
	for _, v := range a {
		if reflect.DeepEqual(v, val) {
			return true
		}
	}

	return false
}

func TestPutChange(t *testing.T) {
	var T = struct{ A string }{A: ""}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(T, attrs).Cast([]string{"A"})

	c = c.PutChange("A", "oi")

	if a, ok := c.GetChange("A"); ok && a != "oi" {
		t.Errorf("PutChange should overwrite the current value")
	}

	c = c.PutChange("B", "ixe")
}

func TestAddError(t *testing.T) {
	var T = struct{ A string }{A: ""}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(T, attrs).Cast([]string{"A"})

	c = c.AddError("A", "WE HAVE AN ERROR")

	err := c.GetError("A")

	if !(err != nil) {
		t.Errorf("AddError should return an error on a existing key")
	}
}

func TestValidateRequired(t *testing.T) {
	var T = struct {
		A string
		B int
	}{A: ""}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(T, attrs).Cast([]string{"A", "B"})

	c = c.ValidateRequired([]string{"A", "B"})

	if c.isValid {
		t.Errorf("ValidateRequired should add error on non existing keys")
	}

	err := c.GetError("B")

	if !(err != nil) {
		t.Errorf("ValidateRequired should return an error on a existing key")
	}
}

func TestUpdateChange(t *testing.T) {
	var T = struct{ A string }{A: ""}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(T, attrs).Cast([]string{"A"})

	c = c.UpdateChange("A", func(a interface{}) interface{} {
		return "foo"
	})

	if a, ok := c.GetChange("A"); ok && a == "hello" {
		t.Errorf("UpdateChange should modify the current change value on specified key")
	}

}

func TestValidateAcceptance(t *testing.T) {
	var T = struct{ A bool }{A: false}
	attrs := map[string]interface{}{"A": false}
	c := New(T, attrs).Cast([]string{"A"})

	c = c.ValidateAcceptance("A")

	if c.isValid {
		t.Errorf("ValidateAcceptance should add error on non tru keys")
	}

	err := c.GetError("A")

	if !(err != nil) {
		t.Errorf("ValidateAcceptance should return an error on a non true key")
	}
}

func TestValidateFormat(t *testing.T) {
	var T = struct{ A string }{A: "hello"}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(T, attrs).Cast([]string{"A"})

	re := regexp.MustCompile("hello")
	c = c.ValidateFormat("A", re)

	if !c.isValid {
		t.Errorf("ValidateFormat shouldn't add error on a valid format keys")
	}

	err := c.GetError("A")

	if err != nil {
		t.Errorf("ValidateFormat shouldn't return an error on a valid format key")
	}
}

type T struct {
	A string
}

func TestApply(t *testing.T) {
	s := T{""}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(s, attrs).Cast([]string{"A"})
	r, err := Apply[T](c)

	if err != nil {
		t.Errorf("Apply should returna a valid T struct")
	}

	if r.A != "hello" {
		t.Errorf("Apply should return the struct with updates")
	}
}

func TestValidateLength(t *testing.T) {
	var T = struct{ A string }{A: ""}
	attrs := map[string]interface{}{"A": "hello"}
	c := New(T, attrs).Cast([]string{"A"})

	if c := c.ValidateLength("A", 5); !c.isValid {
		t.Errorf("ValidateChange should return truthy on correct length")
	}

	if c := c.ValidateLength("A", 10); c.isValid {
		t.Errorf("ValidateChange should return falsy on wrong length")
	}

	var R = struct{ A int }{A: 10}
	attrs = map[string]interface{}{"A": 20}
	c = New(R, attrs).Cast([]string{"A"})

	if c := c.ValidateLength("A", 5); c.isValid {
		t.Errorf("ValidateChange should return error on non string and slices types")
	}
}

func TestCast(t *testing.T) {
	var T = struct {
		A string
		B int
	}{A: "", B: 0}

	attrs := map[string]interface{}{"foo": 123, "A": "hello", "B": 2}
	c := New(T, attrs).Cast([]string{"A", "B"})

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
	var T = struct {
		A string
		B int
	}{A: "", B: 0}

	attrs := map[string]interface{}{"foo": 123, "A": "hello", "B": 2}
	c := New(T, attrs).Cast([]string{"A", "B"})

	if v, ok := c.GetChange("A"); !ok || !reflect.DeepEqual(v, "hello") {
		t.Errorf("GetChange should return the correct value from key, got: %v", v)
	}

	if v, ok := c.GetChange("B"); !ok || !reflect.DeepEqual(v, 2) {
		t.Errorf("GetChange should return the correct value from key, got: %v", v)
	}
}

func TestStructFields(t *testing.T) {
	var T = struct {
		A string
		B int
	}{A: "hello", B: 2}

	a := StructFields(T)

	if typ := reflect.TypeOf(a).Kind(); typ != reflect.Slice {
		t.Errorf("StructFields should return an array of fields, got: %v", typ)
	}

	if l := len(a); l != 2 {
		t.Errorf("StructFields should return an array of length 2, got: %d", l)
	}
}

func TestMap(t *testing.T) {
	var T = struct {
		A string
		B int
	}{A: "hello", B: 2}

	a := ToMap(T)

	if typ := reflect.TypeOf(a).Kind(); typ != reflect.Map {
		t.Errorf("Map should return aa map, got: %v", typ)
	}

	if l := len(a); l != 2 {
		t.Errorf("Mapshould return a map of length 2, got: %d", l)
	}

	for _, val := range []interface{}{"hello", 2} {
		if !inMap(a, val) {
			t.Errorf("Map should have the value %v", val)
		}
	}
}
