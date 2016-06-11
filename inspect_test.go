package jsontools

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = fmt.Println

func TestDecode(t *testing.T) {
	r := strings.NewReader(`{"a": "foo", "b": [1, 2, 3], "c": {"x": "moge", "y": false, "z": 3.14}}`)
	expected := &objectValue{
		props: map[string]jsonValue{
			"a": &stringValue{"foo"},
			"b": &arrayValue{
				elems: []jsonValue{
					&numberValue{1},
					&numberValue{2},
					&numberValue{3},
				},
			},
			"c": &objectValue{
				props: map[string]jsonValue{
					"x": &stringValue{"moge"},
					"y": &literalValue{False},
					"z": &numberValue{3.14},
				},
			},
		},
	}
	value, err := Decode(r)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, value)
}
