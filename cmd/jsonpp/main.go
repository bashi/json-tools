package main

import (
	"flag"
	"io"
	"os"

	"github.com/bashi/json-tools"
)

var nocolor = flag.Bool("nocolor", false, "No color")
var indent = flag.Int("indent", 2, "Indent width")

func print(r io.Reader) {
	formatter := jsontools.NewFormatter(r, os.Stdout)
	formatter.SetIndentWidth(*indent)
	if !*nocolor {
		formatter.EnableColor()
	}
	if err := formatter.Dump(); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		print(os.Stdin)
	} else if flag.NArg() == 1 {
		r, err := os.Open(flag.Arg(0))
		if err != nil {
			panic(err)
		}
		defer r.Close()
		print(r)
	} else {
		panic("Invalid number of arguments.")
	}
}
