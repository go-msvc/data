package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/log"
)

//IData supports both IGetter and ISetter
type IData interface {
	IGetter
	ISetter
	String(name string) string
}

//Data is name-value pairs of data that implements both IDataGetter and IDataSetter
type Data map[string]interface{}

//NewData ...
func NewData() Data {
	d := Data(map[string]interface{}{})
	return d
}

//Get implements IDataGetter.Get()
func (d Data) Get(name string) (interface{}, error) {
	//split nested name on '.' and eliminate any empty names
	log.Debugf("Get(%s)", name)
	names := []string{}
	for _, n := range strings.Split(name, ".") {
		if len(n) > 0 {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return d, nil
	}

	return get(d, names)
}

func (d Data) String(name string) string {
	v, err := d.Get(name)
	if err != nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func get(d map[string]interface{}, names []string) (interface{}, error) {
	log.Debugf("get(%v) from %+v", names, d)
	if len(names) == 0 {
		return d, nil
	}

	//work on first name only, then recurse to next
	name := names[0]
	if value, ok := d[name]; ok {
		if m, ok := value.(map[string]interface{}); ok {
			log.Debugf("value[%s]=%T is a map", name, value)
			return get(m, names[1:])
		}
		//not a map, not expecting sub-names
		log.Debugf("value[%s]=%T is not a map", name, value)
		if len(names) > 1 {
			//the reference is looking for a object attr,
			//but this match is just a value, so skip this as not found
			//it may refer to another field in another source
			log.Debugf("found %s type(%T) but does not match nested name [%v]", name, value, names)
			return nil, nil
		}
		log.Debugf("get(%s)=%T=%+v", name, value, value)
		return value, nil
	}
	log.Debugf("name=\"%s\" not found in d=%+v", name, d)
	return nil, nil
}

//Set ...
func (d Data) Set(name string, value interface{}) {
	d[name] = value
}

//Assign applies a set of assignments configured as "<name>":<value> to the data set.
//The name may be complex as "<a>.<b>.<c>" to set nested items, creating parent objects as needed.
//The value can refer to named items in from or otherwise d using {{...}} around the name of the reference.
//References are resolved in the sets of data in from, which may include d
//Since from is an array of pointers to Data: when one of the items is d, it is possible
//for a next assignment to refer to the value set in a previous assignment of the same set of assignments.
func (d Data) Assign(assignments Assignments, from []IGetter) error {
	if d == nil {
		return errors.Errorf("nil.Set()")
	}

	if assignments == nil {
		return nil //nothing to do
	}

	if len(from) < 1 {
		return errors.Errorf("Data.Set(no from data)")
	}

	//this is a map - so sequence of config is not necessarily the same
	//as the sequence we make the assigments here, so if one assignment
	//uses the value in another assignment and they are executed out of
	//sequence, the assignments will not be good
	for _, assignment := range assignments {
		err := d.setCompoundNV(assignment.Name, assignment.Value, from)
		if err != nil {
			return errors.Wrapf(err, "failed to set \"%s\"", assignment.Name)
		}
	}
	return nil
} //Data.Set()

//Assign ...
func Assign(d map[string]interface{}, assignments Assignments, from []IGetter) error {
	if d == nil {
		return errors.Errorf("nil.Set()")
	}

	if assignments == nil {
		return nil //nothing to do
	}

	if len(from) < 1 {
		return errors.Errorf("Data.Set(no from data)")
	}

	//this is a map - so sequence of config is not necessarily the same
	//as the sequence we make the assigments here, so if one assignment
	//uses the value in another assignment and they are executed out of
	//sequence, the assignments will not be good
	for _, assignment := range assignments {
		err := setCompoundNV(d, assignment.Name, assignment.Value, from)
		if err != nil {
			return errors.Wrapf(err, "failed to set \"%s\"", assignment.Name)
		}
	}
	return nil
} //Assign()

//SetCompoundNV sets a name value spec in the data set, both may be complex/nested
func (d Data) setCompoundNV(nameSpec string, v interface{}, source []IGetter) error {
	if d == nil {
		return errors.Errorf("nil.SetNV()")
	}
	if len(nameSpec) < 1 {
		return errors.Errorf("Data.SetNV(name=\"\")")
	}

	//enter sub-item is specified compound name like a.b.c
	names := strings.SplitN(nameSpec, ".", 2)
	switch len(names) {
	case 0:
		return errors.Errorf("invalid name \"%s\"", nameSpec)
	case 1:
		return d.setNV(names[0], v, source)
	} //switch(len(names))

	//has two name parts:
	//define empty object if not exists
	if _, ok := d[names[0]]; !ok {
		//name does not exist: create
		d[names[0]] = NewData()
	} else {
		//named exists: check type
		if _, ok := d[names[0]].(Data); !ok {
			return errors.Errorf("data[%s]=%T (!=name-value type)", names[0], d[names[0]])
		}
	}

	sub := d[names[0]].(Data)
	return sub.setCompoundNV(names[1], v, source)
} //Data.setCompoundNV()

//SetCompoundNV sets a name value spec in the data set, both may be complex/nested
func setCompoundNV(d map[string]interface{}, nameSpec string, v interface{}, source []IGetter) error {
	if d == nil {
		return errors.Errorf("nil.SetNV()")
	}
	if len(nameSpec) < 1 {
		return errors.Errorf("Data.SetNV(name=\"\")")
	}

	//enter sub-item is specified compound name like a.b.c
	names := strings.SplitN(nameSpec, ".", 2)
	switch len(names) {
	case 0:
		return errors.Errorf("invalid name \"%s\"", nameSpec)
	case 1:
		return setNV(d, names[0], v, source)
	} //switch(len(names))

	//has two name parts:
	//define empty object if not exists
	if _, ok := d[names[0]]; !ok {
		//name does not exist: create
		d[names[0]] = NewData()
	} else {
		//named exists: check type
		if _, ok := d[names[0]].(Data); !ok {
			return errors.Errorf("data[%s]=%T (!=name-value type)", names[0], d[names[0]])
		}
	}

	sub := d[names[0]].(Data)
	return sub.setCompoundNV(names[1], v, source)
} //setCompoundNV()

//SetNV operates on simple names ...
func (d Data) setNV(name string, v interface{}, source []IGetter) error {
	log.Tracef("setNV(%s:%v) ...", name, v)

	//determine the value
	switch v.(type) {
	case string:
		return setNVString(d, name, v.(string), source)
	case map[string]interface{}:
		subData := NewData()
		for nn, vv := range v.(map[string]interface{}) {
			log.Tracef("setNV: %s.%s: v=%T=%+v ...", name, nn, vv, vv)
			subData.setNV(nn, vv, source)
		}
		d[name] = subData
		log.Tracef("  Set  %s = (%T)%+v", name, subData, subData)

	default:
		log.Tracef("  Set %s = (%T)%v", name, v, v)
		d[name] = v
	}
	return nil
}

func setNV(d map[string]interface{}, name string, v interface{}, source []IGetter) error {
	//determine the value
	switch v.(type) {
	case string:
		log.Tracef("  Set  %s = (%T)%+v", name, v, v)
		return setNVString(d, name, v.(string), source)
	case map[string]interface{}:
		subData := NewData()
		for nn, vv := range v.(map[string]interface{}) {
			log.Tracef("setNV: %s.%s: v=%T=%+v ...", name, nn, vv, vv)
			subData.setNV(nn, vv, source)
		}
		d[name] = subData
		log.Tracef("  Set  %s = (%T)%+v", name, subData, subData)

	default:
		log.Tracef("  Set %s = (%T)%v", name, v, v)
		d[name] = v
	}

	// if srcData.Result() != nil {
	// 	return errors.Wrapf(srcData.Result(), "Failed to populate %s", name)
	// }

	// //is CDR wasp_id not set and this is called wasp_id, store the value
	// if len(s.cdrWaspID) == 0 && n == "wasp_id" {
	// 	s.cdrWaspID = outputObj[n].(string)
	// }

	return nil
} //setNV()

func setNVString(d map[string]interface{}, name string, str string, source []IGetter) error {
	//replace any {{...}} in str with the referenced values
	matches := variableRegex.FindAllString(str, -1)
	log.Debugf("set \"%s\":\"%s\" which has %d references", name, str, len(matches))
	substituted := false
	if len(matches) > 0 {
		for _, itemName := range matches {
			//strip the {{ and }} from itemName
			itemName = itemName[2 : len(itemName)-2]
			log.Debugf("looking for referenced item:\"%s\" in %d sources", itemName, len(source))

			//get item from any source, stop when found
			gotValue := false
			for i := 0; i < len(source); i++ {
				itemValue, err := source[i].Get(itemName)
				if err != nil {
					return errors.Wrapf(err, "source[%d] failed to get(%s)", i, name)
				}

				if itemValue == nil {
					//not present in this source
					continue
				}

				//found referenced value in a source
				gotValue = true
				if len(matches) == 1 && str == "{{"+itemName+"}}" {
					//the {{...}} is the complete valueSpec,
					//copy the value directly to preserve its type
					//log.Debugf("  Set  %s = {{%s}} = (%T)%v", name, itemName, itemValue, itemValue)
					d[name] = itemValue
				} else {
					//the valueSpec is not only {{...}}
					//substitute value into s keeping the formatting around it
					str = strings.Replace(str, "{{"+itemName+"}}", fmt.Sprintf("%v", itemValue), 1)
					//log.Debugf("  Substituted {{%s}} = (%T)%v, Now: \"%s\"", itemName, itemValue, itemValue, str)
					substituted = true
				}
				break
			} //for each source

			if !gotValue {
				return errors.Errorf("%s=%s: {{%s}} not found in any source", name, str, itemName)
			}
		} //for each {{...}} reference
	} //if found {{...}} references in value spec

	//if value spec (s) was only this value "{{xxx}}", then we already copied the value above
	//but if it was formatted with more than just the value, we set the substituted string value here
	if substituted || len(matches) == 0 {
		log.Debugf("Just setting value: %s=%s", name, str)
		d[name] = str
	}
	return nil
} //Data.setNVString()

//MarshalJSON ...
func (d Data) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(d))
}
