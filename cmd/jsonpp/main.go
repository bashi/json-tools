package main

import (
	"flag"
	"os"

	"github.com/bashi/json-tools/jsonlib"
)

var nocolor = flag.Bool("nocolor", false, "No color")
var indent = flag.Int("indent", 2, "Indent width")

func main() {
	flag.Parse()

	formatter := jsonlib.NewFormatter(os.Stdin, os.Stdout)
	formatter.SetIndentWidth(*indent)
	if !*nocolor {
		formatter.EnableColor()
	}

	if err := formatter.Dump(); err != nil {
		panic(err)
	}
}
