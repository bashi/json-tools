package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bashi/json-tools"
)

type jsonClient struct {
	jsontools.ParserClientBase
	w io.Writer
}

func (c *jsonClient) StartObject() {
	fmt.Fprintf(c.w, "{")
}

func (c *jsonClient) EndObject() {
	fmt.Fprintf(c.w, "}")
}

func (c *jsonClient) StartArray() {
	fmt.Fprintf(c.w, "[")
}

func (c *jsonClient) EndArray() {
	fmt.Fprintf(c.w, "]")
}

func (c *jsonClient) StartMember(s string) {
	fmt.Fprintf(c.w, "%s:", s)
}

func (c *jsonClient) EndMember(next jsontools.HasNext) {
	if next {
		fmt.Fprintf(c.w, ",")
	}
}

func (c *jsonClient) EndValue(next jsontools.HasNext) {
	if next {
		fmt.Fprintf(c.w, ",")
	}
}

func (c *jsonClient) StringValue(s string) {
	fmt.Fprintf(c.w, "%s", s)
}

func (c *jsonClient) NumberValue(n string) {
	fmt.Fprintf(c.w, "%s", n)
}

func (c *jsonClient) LiteralValue(l jsontools.Literal) {
	fmt.Fprintf(c.w, "%s", l.String())
}

func compact(r io.Reader, w io.Writer) error {
	client := &jsonClient{
		w: w,
	}
	parser := jsontools.NewParser(r, client)
	return parser.Parse()
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		os.Exit(1)
	}
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer r.Close()
	compact(r, os.Stdout)
}
