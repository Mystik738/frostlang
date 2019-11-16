package frostlang_test

import (
	"encoding/hex"
	"frostlang"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestConvertLangToJSON(t *testing.T) {
	input_data, _ := hex.DecodeString(strings.ReplaceAll("30 30 30 30 30 30 30 30 08 00 74 61 67 2F 73 75 62 31 04 00 74 00 65 00 78 00 74 00 08 00 74 61 67 2F 73 75 62 32 04 00 74 00 65 00 73 00 74 00", " ", ""))
	expected_result := `{
  "tag":{
    "sub1":"text",
    "sub2":"test"
  }
}`

	dir, _ := os.Getwd()
	file, _ := os.Create("test.lang")

	file.Write(input_data)
	file.Close()
	frostlang.ConvertLangToJSON(dir, true)
	content, _ := ioutil.ReadFile("test.json")

	if string(content) != expected_result {
		t.Errorf("Content unexpected")
	}

	os.Remove("test.lang")
	os.Remove("test.json")
}
