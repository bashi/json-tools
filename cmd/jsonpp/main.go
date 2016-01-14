package main

import (
	"os"

	"github.com/bashi/json-tools/jsonlib"
)

func main() {
	formatter := jsonlib.NewFormatter(os.Stdin, os.Stdout)
	if err := formatter.Dump(); err != nil {
		panic(err)
	}
}
