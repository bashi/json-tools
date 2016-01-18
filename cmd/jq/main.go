package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bashi/json-tools/index"
	"github.com/k0kubun/pp"
)

var _ = pp.Println

func repl(index *index.Index) {
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
		results := index.Match(q)
		for _, r := range results {
			fmt.Println(r)
		}
	}
}

func main() {
	flag.Parse()
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	indexer := index.NewIndexer(file)
	index, err := indexer.CreateIndex()
	if err != nil {
		panic(err)
	}
	//pp.Println(index)
	repl(index)
}
