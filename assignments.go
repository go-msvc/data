package data

import "github.com/go-msvc/errors"

//NewAssignments ...
func NewAssignments() Assignments {
	return make([]Assignment, 0)
}

//Assignments is a set of instructions to set data values
//the sequence of the assignments matters, hence an array and not a map
type Assignments []Assignment

//With ...
func (a Assignments) With(name string, value interface{}) Assignments {
	a = append(a, Assignment{name, value})
	return a
}

//Merge another set of assignments into this set,
//overwriting duplicate names
func (a Assignments) Merge(b Assignments) Assignments {
	a = append(a, b...)
	return a
}

//Validate ...
func (a Assignments) Validate() error {
	for idx, nv := range a {
		if len(nv.Name) == 0 {
			return errors.Errorf("Assignment[%d] has no name", idx)
		}
	}
	return nil
}

//Assignment ...
type Assignment struct {
	Name  string      `json:"name" doc:"Name to set, e.g. \"a\" or \"a.b\""`
	Value interface{} `json:"value" doc:"Value to set, may be constance or include {{...}} references to names."`
}
