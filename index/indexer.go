package index

import (
	"io"
	"strings"

	"github.com/bashi/json-tools/parse"
)

type indexerClient struct {
	parse.ParserClientBase

	path    []string
	pathStr string
	index   *Index
	parser  *parse.Parser
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
	entry := IndexEntry{Path: i.pathStr, Pos: i.parser.CurrentPos()}
	i.index.memberIndex[s] = append(i.index.memberIndex[s], entry)
}

func (i *indexerClient) AddString(s string) {
	entry := IndexEntry{Path: i.pathStr, Pos: i.parser.CurrentPos()}
	i.index.stringIndex[s] = append(i.index.stringIndex[s], entry)
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

func (i *indexerClient) StringValue(s string) {
	s = s[1 : len(s)-1]
	i.AddString(s)
}

type Indexer struct {
	parser *parse.Parser
	index  *Index
}

func (i *Indexer) CreateIndex() (*Index, error) {
	if err := i.parser.Parse(); err != nil {
		return nil, err
	}
	return i.index, nil
}

func NewIndexer(r io.Reader) *Indexer {
	index := newIndex()
	client := &indexerClient{
		path:    make([]string, 0),
		pathStr: "",
		index:   index,
	}
	parser := parse.NewParser(r, client)
	client.parser = parser
	return &Indexer{
		parser: parser,
		index:  index,
	}
}
