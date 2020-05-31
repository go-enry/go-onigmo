package rubex

import (
	"reflect"
	"testing"
)

func FooTestExpand(t *testing.T) {
	content := []byte(`
	# comment line
	option1: value1
	option2: value2

	# another comment line
	option3: value3
`)

	pattern := MustCompile(`(?m)(?P<key>\w+):\s+(?P<value>\w+)$`)
	template := []byte("$key=$value\n")
	result := []byte{}

	for _, submatches := range pattern.FindAllSubmatchIndex(content, -1) {
		result = pattern.Expand(result, template, content, submatches)
	}

	expected := []byte("option1=value1\noption2=value2\noption3=value3\n")
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %q; want %q", expected, result)
	}
}

func TestExpandString(t *testing.T) {
	content := `
	# comment line
	option1: value1
	option2: value2

	# another comment line
	option3: value3
`

	pattern := MustCompile(`(?m)(?P<key>\w+):\s+(?P<value>\w+)$`)
	template := "$key=$value\n"
	result := []byte{}

	for _, submatches := range pattern.FindAllStringSubmatchIndex(content, -1) {
		result = pattern.ExpandString(result, template, content, submatches)
	}

	expected := []byte("option1=value1\noption2=value2\noption3=value3\n")
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %q; want %q", expected, result)
	}
}
