package data_test

import (
	"reflect"
	"testing"

	"github.com/go-msvc/data"
	"github.com/go-msvc/errors"
)

func init() {
	//	data.Debug()
}

func Test1(t *testing.T) {
	d := data.NewData()
	assert(t, d.Assign(nil, nil))
	assert(t, verify(d, map[string]interface{}{}))

	t.Logf("d=%+v", d)
}

func TestConstStrInt1(t *testing.T) {
	//simple const int and string assignments
	a1 := data.NewAssignments().With("a", 1).With("b", "2").With("c", nil)
	d := data.NewData()
	assert(t, d.Assign(a1, []data.IGetter{d}))
	assert(t, verify(d, map[string]interface{}{"a": 1, "b": "2", "c": nil}))

	t.Logf("d=%+v", d)
}

func TestNestedName(t *testing.T) {
	//simple const int and string assignments
	a1 := data.NewAssignments().With("a.b", 1).With("a.c", "2")
	d := data.NewData()
	assert(t, d.Assign(a1, []data.IGetter{d}))
	assert(t, verify(d["a"].(data.Data), map[string]interface{}{"b": 1, "c": "2"}))

	t.Logf("d=%+v", d)
}

func TestRef1(t *testing.T) {
	//simple const int and string assignments
	a1 := data.NewAssignments().With("a", 123).With("b", "{{a}}").With("c", "[{{a}}]")
	d := data.NewData()
	assert(t, d.Assign(a1, []data.IGetter{d}))
	assert(t, verify(d, map[string]interface{}{"a": 123, "b": 123, "c": "[123]"}))

	t.Logf("d=%+v", d)
}

func TestComplexValue(t *testing.T) {
	//simple const int and string assignments
	a1 := data.NewAssignments().
		With("a", map[string]interface{}{"A": 1, "B": 2}).
		With("b", "{{a}}").
		With("c", "[{{a}}]").
		With("d", map[string]interface{}{"C": map[string]interface{}{"one": 1, "two": "two"}})
	d := data.NewData()
	assert(t, data.Assign(d, a1, []data.IGetter{data.Data(d)}))
	t.Logf("d=%+v", d)
	assert(t, verify(d,
		map[string]interface{}{
			"a": map[string]interface{}{"A": 1, "B": 2},
			"b": map[string]interface{}{"A": 1, "B": 2},
			"c": "[map[A:1 B:2]]", //...go value printed with %v
			"d": map[string]interface{}{"C": map[string]interface{}{"one": 1, "two": "two"}},
		}))

	t.Logf("d=%+v", d)
}

func verify(d data.Data, e data.Data) error {
	//all items in e must be in d
	for n, ev := range e {
		dv, ok := d[n]
		if !ok {
			return errors.Errorf("%s does not exist in data", n)
		}
		if dm, ok := dv.(data.Data); ok {
			if em, ok := ev.(map[string]interface{}); ok {
				if err := verify(dm, em); err != nil {
					return errors.Errorf("%s:(%T)%v != expected (%T)%v", n, dv, dv, ev, ev)
				}
			} else {
				return errors.Errorf("cannot convert ev(%T) to obj", ev)
			}
		} else {
			if !reflect.DeepEqual(dv, ev) {
				return errors.Errorf("%s:(%T)%v != expected (%T)%v", n, dv, dv, ev, ev)
			}
		}
	}

	//no items in d may not be in e
	for n, _ := range d {
		_, ok := e[n]
		if !ok {
			return errors.Errorf("%s exists but was not expected", n)
		}
	}
	return nil
}

func assert(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Failed: %+v", err)
	}
}
