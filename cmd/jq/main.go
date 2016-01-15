package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

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

type indexer struct {
	parser *parse.Parser
	index  *Index
}

func (i *indexer) createIndex() (*Index, error) {
	if err := i.parser.Parse(); err != nil {
		return nil, err
	}
	return i.index, nil
}

func newIndexer(r io.Reader) *indexer {
	index := newIndex()
	client := &indexerClient{
		path:    make([]string, 0),
		pathStr: "",
		index:   index,
	}
	parser := parse.NewParser(r, client)
	client.parser = parser
	return &indexer{
		parser: parser,
		index:  index,
	}
}

func repl(index *Index) {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			return
		}
		q := strings.Trim(line, "\n")
		results := index.Lookup(q)
		for _, path := range results {
			fmt.Println(path)
		}
	}
}

func main() {
	flag.Parse()
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	indexer := newIndexer(file)
	index, err := indexer.createIndex()
	if err != nil {
		panic(err)
	}
	repl(index)
}
