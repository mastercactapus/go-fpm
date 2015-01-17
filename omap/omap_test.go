package omap

import (
	"encoding/json"
	"testing"
)

const SimpleJSON = `

{
	"n\"ame" : "foo",
	"version": "bar",
	"b": {},
	"a"	: 5,
	"c": null,
	"z": false,
	"f": 9.9,
	"3": {  "m":7, "r":9 }
}


`
const ModifiedJSON = `{"n\"ame":"foo","version":"2.0.0","b":{},"a":5,"c":null,"z":false,"f":9.9,"3":{"m":7,"r":9},"version2":"2.0.0","apple":"2.0.0"}`

func TestUnmarshalMarshalJSON(t *testing.T) {
	var m = NewOrderedMap()
	err := json.Unmarshal([]byte(SimpleJSON), &m)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s", err.Error())
	}

	var version string
	err = json.Unmarshal(m.Get("version"), &version)
	if err != nil {
		t.Fatalf("Failed to unmarshal version: %s", err.Error())
	}
	if version != "bar" {
		t.Errorf("Version was decoded improperly, expected '%s' but got '%s'", "bar", version)
	}
	version = "2.0.0"
	data, err := json.Marshal(&version)
	if err != nil {
		t.Fatalf("Could not marshal string: %s", err.Error())
	}
	m.Set("version2", json.RawMessage(data))
	m.Set("version", json.RawMessage(data))
	m.Set("apple", json.RawMessage(data))

	data, err = json.Marshal(&m)
	if err != nil {
		t.Fatalf("Could not marshal the OrderedMap: %s", err.Error())
	}
	if string(data) != ModifiedJSON {
		t.Errorf("Marshal error: expected '%s' but got '%s'", ModifiedJSON, string(data))
	}
}
