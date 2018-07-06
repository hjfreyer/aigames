package main

import (
	"encoding/json"
	"os"
	"testing"
)

func Test(t *testing.T) {
	var ins TestInputs

	if err := ReadJson("testdata/inputs.json", &ins); err != nil {
		t.Fatal(err)
	}

	actualOuts, err := RunTests(&ins)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("testdata/outputs.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	if err := e.Encode(actualOuts); err != nil {
		t.Fatal(err)
	}

	var goldenOuts TestOutputs
	if err := ReadJson("testdata/golden.json", &goldenOuts); err != nil {
		t.Fatal(err)
	}

	if err := DiffOuts(actualOuts, &goldenOuts); err != nil {
		t.Fatal(err)
	}
}
