package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/bashi/json-tools/parse"
	"github.com/fatih/color"
)

const (
	defaultIndentSize = 2
)

// formatClient is a ParserClient for Printer
type formatClient struct {
	w          io.Writer
	indent     string
	indentUnit string
	// colorize funcs
	memberColor  *color.Color
	stringColor  *color.Color
	numberColor  *color.Color
	literalColor *color.Color
}

func (c *formatClient) enterBlock() {
	c.indent = c.indent + c.indentUnit
}

func (c *formatClient) leaveBlock() {
	c.indent = c.indent[:len(c.indent)-len(c.indentUnit)]
}

func (c *formatClient) StartObject() {
	fmt.Fprintf(c.w, "{")
	c.enterBlock()
}

func (c *formatClient) EndObject() {
	c.leaveBlock()
	fmt.Fprintf(c.w, "\n%s}", c.indent)
}

func (c *formatClient) StartArray() {
	fmt.Fprintf(c.w, "[")
	c.enterBlock()
}

func (c *formatClient) EndArray() {
	c.leaveBlock()
	fmt.Fprintf(c.w, "\n%s]", c.indent)
}

func (c *formatClient) StartMember(s string) {
	c.memberColor.Printf("\n%s%s: ", c.indent, s)
}

func (c *formatClient) EndMember(next parse.HasNext) {
	if next {
		fmt.Fprintf(c.w, ",")
	}
}

func (c *formatClient) StartValue() {
	fmt.Fprintf(c.w, "\n%s", c.indent)
}

func (c *formatClient) EndValue(next parse.HasNext) {
	if next {
		fmt.Fprintf(c.w, ",")
	}
}

func (c *formatClient) StringValue(s string) {
	c.stringColor.Printf("%s", s)
}

func (c *formatClient) NumberValue(n string) {
	c.numberColor.Printf("%s", n)
}

func (c *formatClient) LiteralValue(l parse.Literal) {
	c.literalColor.Printf("%s", l.String())
}

type Formatter struct {
	r io.Reader
	c *formatClient
}

func (f *Formatter) Dump() error {
	parser := parse.NewParser(f.r, f.c)
	err := parser.Parse()
	fmt.Fprintln(f.c.w)
	return err
}

func (f *Formatter) SetIndentWidth(n int) {
	f.c.indentUnit = strings.Repeat(" ", n)
}

func (f *Formatter) EnableColor() {
	color.NoColor = false
}

func NewFormatter(r io.Reader, w io.Writer) *Formatter {
	color.Output = w
	color.NoColor = true
	client := &formatClient{
		w:            w,
		indent:       "",
		memberColor:  color.New(color.FgMagenta),
		stringColor:  color.New(color.FgRed),
		numberColor:  color.New(color.FgBlue),
		literalColor: color.New(color.FgCyan),
	}
	f := &Formatter{
		r: r,
		c: client,
	}
	f.SetIndentWidth(defaultIndentSize)
	return f
}
