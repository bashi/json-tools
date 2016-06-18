package main

import (
	"fmt"
	"io"
	"os"

	"github.com/bashi/json-tools"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Must specify JSON file\n")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	defer f.Close()
	if err != nil {
		panic(err)
	}
	i, err := jsontools.NewInspector(f)
	if err != nil {
		panic(err)
	}
	if err := i.Repl(); err != nil && err != io.EOF {
		panic(err)
	}
}
