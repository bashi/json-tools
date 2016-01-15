package index

import (
	"fmt"

	"github.com/bashi/json-tools/parse"
)

type IndexEntry struct {
	Path string
	Pos  parse.ParserPosition
}

func (e IndexEntry) String() string {
	return fmt.Sprintf("%s: %s", e.Pos.String(), e.Path)
}

type Index struct {
	memberIndex map[string][]IndexEntry
	stringIndex map[string][]IndexEntry
}

func (i *Index) Lookup(q string) []IndexEntry {
	var results []IndexEntry
	results = append(results, i.memberIndex[q]...)
	results = append(results, i.stringIndex[q]...)
	return results
}

func newIndex() *Index {
	return &Index{
		memberIndex: make(map[string][]IndexEntry),
		stringIndex: make(map[string][]IndexEntry),
	}
}
