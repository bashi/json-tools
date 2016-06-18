package jsontools

import (
	"fmt"
	"io"
	"strconv"
)

type jsonValue interface {
	ToString() string
}

type objectValue struct {
	props map[string]jsonValue
}

func (v *objectValue) ToString() string {
	return "[Object]"
}

type arrayValue struct {
	elems []jsonValue
}

func (v *arrayValue) ToString() string {
	return "[Array]"
}

type stringValue struct {
	value string
}

func (v *stringValue) ToString() string {
	return v.value
}

type numberValue struct {
	value float64
}

func (v *numberValue) ToString() string {
	return fmt.Sprintf("%f", v.value)
}

type literalValue struct {
	value Literal
}

func (v *literalValue) ToString() string {
	return v.value.String()
}

type decoderClient struct {
	stack       []jsonValue
	memberStack []string
}

func (c *decoderClient) push(v jsonValue) {
	c.stack = append(c.stack, v)
}

func (c *decoderClient) pop() jsonValue {
	curlen := len(c.stack)
	ret := c.stack[curlen-1]
	c.stack = c.stack[:curlen-1]
	return ret
}

func (c *decoderClient) currentObject() *objectValue {
	top := c.stack[len(c.stack)-1]
	return top.(*objectValue)
}

func (c *decoderClient) currentArray() *arrayValue {
	top := c.stack[len(c.stack)-1]
	return top.(*arrayValue)
}

func (c *decoderClient) StartObject() {
	c.push(&objectValue{
		props: make(map[string]jsonValue),
	})
	c.memberStack = append(c.memberStack, "")
}

func (c *decoderClient) EndObject() {
	c.memberStack = c.memberStack[:len(c.memberStack)-1]
}

func (c *decoderClient) StartArray() {
	c.push(&arrayValue{})
}

func (c *decoderClient) EndArray() {
}

func (c *decoderClient) StartMember(s string) {
	c.memberStack[len(c.memberStack)-1] = s[1 : len(s)-1]
}

func (c *decoderClient) EndMember(next HasNext) {
	v := c.pop()
	obj := c.currentObject()
	name := c.memberStack[len(c.memberStack)-1]
	obj.props[name] = v
}

func (c *decoderClient) StartValue() {
}

func (c *decoderClient) EndValue(next HasNext) {
	v := c.pop()
	arr := c.currentArray()
	arr.elems = append(arr.elems, v)
}

func (c *decoderClient) StringValue(s string) {
	c.push(&stringValue{s[1 : len(s)-1]})
}

func (c *decoderClient) NumberValue(s string) {
	n, _ := strconv.ParseFloat(s, 64)
	c.push(&numberValue{n})
}

func (c *decoderClient) LiteralValue(l Literal) {
	c.push(&literalValue{l})
}

func Decode(r io.Reader) (jsonValue, error) {
	c := &decoderClient{}
	parser := NewParser(r, c)
	err := parser.Parse()
	if err != nil {
		return nil, err
	}
	if len(c.stack) != 1 {
		// This shouldn't happen.
		return nil, fmt.Errorf("Internal logic error: %d", len(c.stack))
	}
	return c.pop(), nil
}
