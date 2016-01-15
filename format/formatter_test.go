package format

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var _ = fmt.Print

func compare(t *testing.T, input string, expected string) {
	r := strings.NewReader(input)
	w := new(bytes.Buffer)
	f := NewFormatter(r, w)
	f.SetIndentWidth(2)
	err := f.Dump()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, w.String(), expected)
}

func TestFormat(t *testing.T) {
	compare(t, `[1,2,3]`, `[
  1,
  2,
  3
]
`)
	compare(t, `[1, {"foo": -2}, 2]`, `[
  1,
  {
    "foo": -2
  },
  2
]
`)
	compare(t, `{"true": true, "false": false, "null": null}`, `{
  "true": true,
  "false": false,
  "null": null
}
`)
	compare(t, `{"a": {}, "b": [], "c": {"foo": -3.14, "bar": "pi", "baz": ["", "hoge"]}}`, `{
  "a": {
  },
  "b": [
  ],
  "c": {
    "foo": -3.14,
    "bar": "pi",
    "baz": [
      "",
      "hoge"
    ]
  }
}
`)
	compare(t, `{
  "a": [
  ],
}
`, `{
  "a": [
  ],
}
`)
}
