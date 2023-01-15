package data_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-msvc/assert"
	"github.com/go-msvc/data"
)

func TestGetFromMap(t *testing.T) {
	testData := map[string]interface{}{
		"a": 1,
		"b": "two",
		"c": nil,
		"d": true,
	}

	for name, exp := range testData {
		if got, err := data.Get(testData, name); err != nil || got != exp {
			t.Errorf("name(%s) -> (%T)%v, exp=(%T)%v, err=%+v", name, got, got, exp, exp, err)
		}
	}
}

func TestGetFromSlice(t *testing.T) {
	testData := []interface{}{
		1,
		"two",
		nil,
		true,
	}

	for i := 0; i < len(testData); i++ {
		exp := testData[i]
		name := fmt.Sprintf("[%d]", i)
		if got, err := data.Get(testData, name); err != nil || got != exp {
			t.Errorf("name(%s) -> (%T)%v, exp=(%T)%v, err=%+v", name, got, got, exp, exp, err)
		}
	}
}

func TestGetFromSubs(t *testing.T) {
	type Writer struct {
		Format string `json:"format"`
		Limit  int    `json:"limit"`
	}
	type Config struct {
		//scalar values
		Enabled bool   `json:"enabled"`
		Name    string `json:"name,omitempty"`
		Size    int

		//map values
		Logger map[string]Writer `json:"logger,omitempty"`

		//slice values
		Auditors []Writer `json:"auditors"`
	}
	testData := Config{
		Enabled: true,
		Name:    "Test",
		Size:    1000,
		Logger: map[string]Writer{
			"file":   Writer{Format: "yaml", Limit: 30},
			"remote": Writer{Format: "asn1", Limit: 300},
		},
		Auditors: []Writer{
			{Format: "csv", Limit: 10},
			{Format: "json", Limit: 100},
		},
	}

	assert.Bool(t, "Enabled (field name)", true, data.GetOr(testData, "Enabled", false).(bool))
	assert.Bool(t, "Enabled (json tag name)", true, data.GetOr(testData, "enabled", false).(bool))

	assert.String(t, "Name (field name)", "Test", data.GetOr(testData, "Name", "").(string))
	assert.String(t, "Name (json tag name)", "Test", data.GetOr(testData, "name", "").(string))

	assert.Int(t, "Size (field name)", data.GetOr(testData, "Size", 0).(int), 1000)
	assert.Int(t, "Size (non-existing json tag name)", data.GetOr(testData, "size", 12).(int), 12)

	//an attr that does not exist in the struct should default
	assert.Int(t, "Funny (non-existing field)", data.GetOr(testData, "funne", 123).(int), 123)

	//check map elements using map field name "Logger"
	assert.String(t, "Logger[file].Format (field)", data.GetOr(testData, "Logger[file].format", "").(string), "yaml")
	assert.Int(t, "Logger[file].size (field)", data.GetOr(testData, "Logger[file].Limit", 0).(int), 30)

	//check map elements using map JSON tag name "logger"
	assert.String(t, "logger[remote].Format (field)", data.GetOr(testData, "logger[remote].Format", "").(string), "asn1")
	assert.Int(t, "logger[remote].size (field)", data.GetOr(testData, "logger[remote].limit", 0).(int), 300)

	//check non-existing map elements
	assert.String(t, "logger[local].Format (json)", data.GetOr(testData, "logger[local].Format", "xxx").(string), "xxx")
	assert.Int(t, "Logger[local].size (field)", data.GetOr(testData, "Logger[local].limit", 666).(int), 666)

	//check slice elements using map field name "Logger"
	assert.String(t, "Auditors[0].Format (field)", data.GetOr(testData, "Auditors[0].format", "").(string), "csv")
	assert.Int(t, "Auditors[0].size (field)", data.GetOr(testData, "Auditors[0].Limit", 0).(int), 10)

	//check slice elements using map JSON tag name "logger"
	assert.String(t, "auditors[1].Format (field)", data.GetOr(testData, "auditors[1].Format", "").(string), "json")
	assert.Int(t, "auditors[1].size (field)", data.GetOr(testData, "auditors[1].limit", 0).(int), 100)

	//check non-existing slice elements
	assert.String(t, "auditors[2].Format (field)", data.GetOr(testData, "auditors[2].Format", "yyy").(string), "yyy")
	assert.Int(t, "auditors[2].size (field)", data.GetOr(testData, "auditors[2].limit", 777).(int), 777)

	//check element in non-existing map
	assert.String(t, "people[joe].Name (field)", data.GetOr(testData, "people[joe].name", "Joe").(string), "Joe")
	assert.Int(t, "people[joe].size (field)", data.GetOr(testData, "people[joe].size", 999).(int), 999)
}

func TestGetJSON(t *testing.T) {
	jsonValue := `{
		"enabled":true,
		"name":"Test",
		"size":1000,
		"logger":{
			"file":{"format":"yaml", "limit":30},
			"remote":{"format":"asn1", "limit":300}
		},
		"auditors":[
			{"format":"csv", "limit":10},
			{"format":"json", "limit":100}
		]
	}
	`
	var testData interface{}
	if err := json.Unmarshal([]byte(jsonValue), &testData); err != nil {
		t.Fatalf("failed to parse JSON data: %+v", err)
	}

	assert.Bool(t, "enabled", true, data.GetOr(testData, "enabled", false).(bool))
	assert.String(t, "name", "Test", data.GetOr(testData, "name", "").(string))
	assert.Int(t, "size", 1000, data.GetOr(testData, "size", 12).(int))
	assert.Int(t, "funny (non-existing)", 123, data.GetOr(testData, "funny", 123).(int))

	assert.String(t, "logger[file].format", "yaml", data.GetOr(testData, "logger[file].format", "").(string))
	assert.Int(t, "logger[file].size", 30, data.GetOr(testData, "logger[file].limit", 0).(int))
	assert.String(t, "logger[remote].format", "asn1", data.GetOr(testData, "logger[remote].format", "").(string))
	assert.Int(t, "logger[remote].size", 300, data.GetOr(testData, "logger[remote].limit", 0).(int))

	//check non-existing map elements
	assert.String(t, "logger[local].format", "xxx", data.GetOr(testData, "logger[local].format", "xxx").(string))
	assert.Int(t, "logger[local].size", 666, data.GetOr(testData, "logger[local].limit", 666).(int))

	//check slice elements using map field name "Logger"
	assert.String(t, "auditors[0].format", "csv", data.GetOr(testData, "auditors[0].format", "").(string))
	assert.Int(t, "auditors[0].size", 10, data.GetOr(testData, "auditors[0].limit", 0).(int))

	//check slice elements using map JSON tag name "logger"
	assert.String(t, "auditors[1].format", "json", data.GetOr(testData, "auditors[1].format", "").(string))
	assert.Int(t, "auditors[1].size", 100, data.GetOr(testData, "auditors[1].limit", 0).(int))

	//check non-existing slice elements
	assert.String(t, "auditors[2].format", "yyy", data.GetOr(testData, "auditors[2].format", "yyy").(string))
	assert.Int(t, "auditors[2].size", 777, data.GetOr(testData, "auditors[2].limit", 777).(int))

	//check element in non-existing map
	assert.String(t, "people[joe].name", "Joe", data.GetOr(testData, "people[joe].name", "Joe").(string))
	assert.Int(t, "people[joe].size", 999, data.GetOr(testData, "people[joe].size", 999).(int))
}
