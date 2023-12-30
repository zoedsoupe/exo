package exo_test

import (
	"reflect"
	"testing"

	"github.com/zoedsoupe/exo"
)

func inMap(a map[string]interface{}, val interface{}) bool {
	for _, v := range a {
		if reflect.DeepEqual(v, val) {
			return true
		}
	}

	return false
}

func TestStructFields(t *testing.T) {
	var T = struct {
		A string
		B int
	}{A: "hello", B: 2}

	a := exo.StructFields(T)

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

	a := exo.ToMap(T)

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
