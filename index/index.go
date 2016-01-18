package index

import (
	"io"
	"strconv"

	"bytes"
	"fmt"
	"github.com/bashi/json-tools/parse"
	"regexp"
)

type IdentId int

type IndexEntry struct {
	Ident string
	Path  string
	Pos   parse.ParserPosition
}

func (i *IndexEntry) String() string {
	return fmt.Sprintf("%s: [%s]: %s", i.Pos.String(), i.Path, i.Ident)
}

type indexEntryInternal struct {
	Path []IdentId
	Pos  parse.ParserPosition
}

type Index struct {
	idents   map[string]IdentId
	identIds map[IdentId]string
	idx      map[IdentId][]*indexEntryInternal
}

func (i *Index) buildPathString(path []IdentId) string {
	l := len(path)
	if l <= 0 {
		return ""
	}

	var buf bytes.Buffer
	for _, id := range path[:l-1] {
		item := i.identIds[id]
		buf.WriteString(item)
		buf.WriteString(" -> ")
	}
	lastItem := i.identIds[path[l-1]]
	buf.WriteString(lastItem)
	return buf.String()
}

func (i *Index) newIndexEntry(id IdentId, e *indexEntryInternal) *IndexEntry {
	return &IndexEntry{
		Ident: i.identIds[id],
		Path:  i.buildPathString(e.Path),
		Pos:   e.Pos,
	}
}

func (i *Index) id(s string) IdentId {
	id, ok := i.idents[s]
	if !ok {
		panic(fmt.Errorf("No ID for %s", s))
	}
	return id
}

func (i *Index) Match(q string) []*IndexEntry {
	var results []*IndexEntry
	re, err := regexp.Compile(q)
	if err != nil {
		return results
	}
	for ident := range i.idents {
		if !re.MatchString(ident) {
			continue
		}
		id := i.id(ident)
		for _, e := range i.idx[id] {
			results = append(results, i.newIndexEntry(id, e))
		}
	}
	return results
}

type indexerClient struct {
	parse.ParserClientBase

	currentIdentId IdentId
	idents         map[string]IdentId
	path           []IdentId
	arrayIndex     int
	parser         *parse.Parser

	idx map[IdentId][]*indexEntryInternal
}

func (i *indexerClient) idFor(s string) IdentId {
	id, ok := i.idents[s]
	if !ok {
		id = i.currentIdentId
		i.idents[s] = i.currentIdentId
		i.currentIdentId++
	}
	return id
}

func (i *indexerClient) index(id IdentId, e *indexEntryInternal) {
	i.idx[id] = append(i.idx[id], e)
}

func (i *indexerClient) indexIdent(s string) {
	entry := &indexEntryInternal{
		Path: make([]IdentId, len(i.path)),
		Pos:  i.parser.CurrentPos(),
	}
	copy(entry.Path, i.path)
	id := i.idFor(s)
	i.index(id, entry)
}

func (i *indexerClient) PushPath(s string) {
	id := i.idFor(s)
	i.path = append(i.path, id)
}

func (i *indexerClient) PopPath() {
	i.path = i.path[:len(i.path)-1]
}

func (i *indexerClient) AddMember(s string) {
	i.indexIdent(s)
}

func (i *indexerClient) AddString(s string) {
	i.indexIdent(s)
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
	client *indexerClient
}

func (i *Indexer) CreateIndex() (*Index, error) {
	if err := i.parser.Parse(); err != nil {
		return nil, err
	}
	identIds := make(map[IdentId]string)
	for ident, id := range i.client.idents {
		identIds[id] = ident
	}
	return &Index{
		idents:   i.client.idents,
		identIds: identIds,
		idx:      i.client.idx,
	}, nil
}

func NewIndexer(r io.Reader) *Indexer {
	client := &indexerClient{
		currentIdentId: 0,
		idents:         make(map[string]IdentId),
		path:           make([]IdentId, 0),
		arrayIndex:     0,
		idx:            make(map[IdentId][]*indexEntryInternal),
	}
	parser := parse.NewParser(r, client)
	client.parser = parser
	return &Indexer{
		parser: parser,
		client: client,
	}
}
