package omap

import (
	"encoding/json"
	"fmt"
	"testing"
)

const SimpleJSON = `

{
	"n\"ame" : "foo",
	"version": "bar",
	"b": {},
 "large": [
{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},
{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},
{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"}

 ],
	"a"	: 5,
	"c": null,
	"z": false,
	"f": 9.9,
	"3": {  "m":7, "r":9 }
}


`
const ModifiedJSON = `{"n\"ame":"foo","version":"2.0.0","b":{},"large":[{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"},{"foo":"bar"}],"a":5,"c":null,"z":false,"f":9.9,"3":{"m":7,"r":9},"version2":"2.0.0","apple":"2.0.0"}`
const PackageJSON = `{"name":"enpm","version":"0.3.0","description":"A proof-of-concept for a new method of installing node packages.","main":"lib/enpm.js","scripts":{},"publishConfig":{"registry":"https://registry.npmjs.org"},"bin":{"enpm":"bin/enpm.js"},"author":{"name":"Nathan Caza"},"license":"MIT","dependencies":{"semver":"^2.2.1","lodash":"^2.4.1","request":"^2.34.0","mkdirp":"^0.4.0","tar":"^0.1.19","glob":"^3.2.9","commander":"^2.2.0","colors":"^0.6.2","bluebird":"^1.2.3"},"_id":"enpm@0.3.0","dist":{"shasum":"136950f3d26e9c60f06e97d049c98a40cc1c0313","tarball":"http://registry.npmjs.org/enpm/-/enpm-0.3.0.tgz"},"_from":".","_npmVersion":"1.4.3","_npmUser":{"name":"mastercactapus","email":"mastercactapus@gmail.com"},"maintainers":[{"name":"mastercactapus","email":"mastercactapus@gmail.com"}],"directories":{}}`

func ExampleOrderedMap_Set() {
	jsonData := `{
		"z": 1,
		"a": 2
	}`

	m := NewOrderedMap()
	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		fmt.Println("error:", err)
	}
	newVal := 3
	newData, err := json.Marshal(&newVal)
	if err != nil {
		fmt.Println("error:", err)
	}
	m.Set("z", newData)
	data, err := json.Marshal(&m)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(data))
	//Output: {"z":3,"a":2}
}

func ExampleOrderedMap_Get() {
	jsonData := `{
		"z": 1,
		"a": 2
	}`

	m := NewOrderedMap()
	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		fmt.Println("error:", err)
	}
	var val int
	err = json.Unmarshal(m.Get("z"), &val)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(val)
	//Output: 1
}

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

	m = NewOrderedMap()
	err = json.Unmarshal([]byte(PackageJSON), &m)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s", err.Error())
	}
	data, err = json.Marshal(&m)
	if err != nil {
		t.Fatalf("Could not marshal string: %s", err.Error())
	}
	if string(data) != PackageJSON {
		t.Errorf("Marshal error: expected '%s' but got '%s'", PackageJSON, string(data))
	}
}
