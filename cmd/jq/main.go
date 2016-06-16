package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bashi/go-repl"
	"github.com/bashi/json-tools"
	"github.com/k0kubun/pp"
)

var _ = pp.Println

func main() {
	flag.Parse()
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	indexer := jsontools.NewIndexer(file)
	index, err := indexer.CreateIndex()
	if err != nil {
		panic(err)
	}
	//pp.Println(index)
	err = repl.Run(func(line string) error {
		results := index.Match(line)
		for _, r := range results {
			fmt.Println(r)
		}
		return nil
	})
	if err != nil && err != io.EOF {
		panic(err)
	}
}
