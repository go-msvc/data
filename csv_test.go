package data_test

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/go-msvc/assert"
	"github.com/go-msvc/data"
)

func TestStruct2CSV(t *testing.T) {
	type Station struct {
		ContainerId      interface{} `json:"containerId"`
		GroupDescription interface{} `json:"groupDescription"`
		Name             string      `json:"name"`
		PadTime          int         `json:"padTime"`
		ScannableId      string      `json:"scannableId"`
		Type             string      `json:"type"`
		Version          interface{} `json:"version"`
		WarehouseId      string      `json:"warehouseId"`
	}
	type ConfigData struct {
		Enabled  bool      `json:"enabled"`
		Stations []Station `json:"stations"`
	}

	config := ConfigData{
		Enabled: false,
		Stations: []Station{
			{
				ContainerId:      nil,
				GroupDescription: "123",
				Name:             "Station-123",
				PadTime:          321,
				ScannableId:      "44335-111",
				Type:             "TheBest",
				Version:          "5.4.3",
				WarehouseId:      "Local-WH-5",
			},
			{
				ContainerId:      "c-456",
				GroupDescription: "456",
				Name:             "Station-456",
				PadTime:          654,
				ScannableId:      "776655-222",
				Type:             "TheWorst",
				Version:          "7.8.9",
				WarehouseId:      "Local-WH-2",
			},
		},
	}

	// config := ConfigData{}
	// if err := json.Unmarshal([]byte(jsonConfig), &config); err != nil {
	// 	t.Fatalf("failed to decode test data: %+v", err)
	// }
	values, err := data.CSV(config)
	if err != nil {
		t.Fatalf("cannot convert json data to CSV: %+v", err)
	}
	t.Logf("%d Values: %v", len(values), values)
	if len(values) != 17 {
		t.Fatalf("got %d != 17 values", len(values))
	}

	b := bytes.NewBuffer(nil)
	w := csv.NewWriter(b)
	if err := w.Write(values); err != nil {
		t.Fatalf("failed: %+v", err)
	}
	w.Flush()
	csvString := string(b.Bytes())
	t.Logf("CSV string: %s", csvString)
	expectedCSV := "false,,123,Station-123,321,44335-111,TheBest,5.4.3,Local-WH-5,c-456,456,Station-456,654,776655-222,TheWorst,7.8.9,Local-WH-2\n"
	assert.String(t, "csv string", expectedCSV, csvString)
}
