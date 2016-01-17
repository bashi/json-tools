package index

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/bashi/json-tools/parse"
)

type IndexEntry struct {
	Ident string
	Path  string
	Pos   parse.ParserPosition
}

func (e IndexEntry) String() string {
	return fmt.Sprintf("%s: %s %s", e.Pos.String(), e.Path, e.Ident)
}

// TODO: Replace this naive impl with a sophisticated one which supports
// regexp queries.
type indexImpl struct {
	idx map[string][]*IndexEntry
}

func (i *indexImpl) Index(e *IndexEntry) {
	i.idx[e.Ident] = append(i.idx[e.Ident], e)
}

func (i *indexImpl) Lookup(q string) []string {
	var results []string
	for _, e := range i.idx[q] {
		results = append(results, e.String())
	}
	return results
}

type Index struct {
	impl *indexImpl
}

func (i *Index) Lookup(q string) string {
	return strings.Join(i.impl.Lookup(q), "\n")
}

type indexerClient struct {
	parse.ParserClientBase

	path       []string
	pathStr    string
	arrayIndex int
	parser     *parse.Parser

	index *indexImpl
}

func (i *indexerClient) updatePath() {
	i.pathStr = strings.Join(i.path, " -> ")
}

func (i *indexerClient) PushPath(s string) {
	i.path = append(i.path, s)
	i.updatePath()
}

func (i *indexerClient) PopPath() {
	i.path = i.path[:len(i.path)-1]
	i.updatePath()
}

func (i *indexerClient) AddMember(s string) {
	entry := &IndexEntry{
		Ident: s,
		Path:  i.pathStr,
		Pos:   i.parser.CurrentPos(),
	}
	i.index.Index(entry)
}

func (i *indexerClient) AddString(s string) {
	entry := &IndexEntry{
		Ident: s,
		Path:  i.pathStr,
		Pos:   i.parser.CurrentPos(),
	}
	i.index.Index(entry)
}

// ParserClient implementations

func (i *indexerClient) StartMember(s string) {
	s = s[1 : len(s)-1]
	i.AddMember(s)
	i.PushPath(s)
}

func (i *indexerClient) EndMember(parse.HasNext) {
	i.PopPath()
}

func (i *indexerClient) StartArray() {
	i.arrayIndex = 0
}

func (i *indexerClient) StartValue() {
	i.PushPath(strconv.Itoa(i.arrayIndex))
	i.arrayIndex += 1
}

func (i *indexerClient) EndValue(parse.HasNext) {
	i.PopPath()
}

func (i *indexerClient) StringValue(s string) {
	s = s[1 : len(s)-1]
	i.AddString(s)
}

type Indexer struct {
	parser *parse.Parser
	index  *indexImpl
}

func (i *Indexer) CreateIndex() (*Index, error) {
	if err := i.parser.Parse(); err != nil {
		return nil, err
	}
	return &Index{
		impl: i.index,
	}, nil
}

func NewIndexer(r io.Reader) *Indexer {
	index := &indexImpl{
		idx: make(map[string][]*IndexEntry),
	}
	client := &indexerClient{
		path:       make([]string, 0),
		pathStr:    "",
		arrayIndex: 0,
		index:      index,
	}
	parser := parse.NewParser(r, client)
	client.parser = parser
	return &Indexer{
		parser: parser,
		index:  index,
	}
}
