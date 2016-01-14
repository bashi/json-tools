package jsonlib

import (
	"fmt"
	"io"
	"strings"
)

const (
	defaultIndentSize = 2
)

// formatClient is a ParserClient for Printer
type formatClient struct {
	w          io.Writer
	indent     string
	indentUnit string
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
	fmt.Fprintf(c.w, "\n%s%s: ", c.indent, s)
}

func (c *formatClient) EndMember(next HasNext) {
	if next {
		fmt.Fprintf(c.w, ",")
	}
}

func (c *formatClient) StartValue() {
	fmt.Fprintf(c.w, "\n%s", c.indent)
}

func (c *formatClient) EndValue(next HasNext) {
	if next {
		fmt.Fprintf(c.w, ",")
	}
}

func (c *formatClient) StringValue(s string) {
	fmt.Fprintf(c.w, "%s", s)
}

func (c *formatClient) NumberValue(n string) {
	fmt.Fprintf(c.w, "%s", n)
}

func (c *formatClient) LiteralValue(l Literal) {
	fmt.Fprintf(c.w, "%s", l.String())
}

type Formatter struct {
	r io.Reader
	c *formatClient
}

func (f *Formatter) Dump() error {
	parser := NewParser(f.r, f.c)
	err := parser.Parse()
	fmt.Fprintln(f.c.w)
	return err
}

func (f *Formatter) SetIndentWidth(n int) {
	f.c.indentUnit = strings.Repeat(" ", n)
}

func NewFormatter(r io.Reader, w io.Writer) *Formatter {
	client := &formatClient{
		w:      w,
		indent: "",
	}
	f := &Formatter{
		r: r,
		c: client,
	}
	f.SetIndentWidth(defaultIndentSize)
	return f
}
