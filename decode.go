package jsontools

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
)

type symbolTable map[uint]string

type symtabMaker struct {
	nextId   uint
	symtab   symbolTable
	inverted map[string]uint
}

func newSymtabMaker() *symtabMaker {
	return &symtabMaker{
		nextId:   0,
		symtab:   make(symbolTable),
		inverted: make(map[string]uint),
	}
}

func (m *symtabMaker) getId(s string) uint {
	if id, ok := m.inverted[s]; ok {
		return id
	}
	id := m.nextId
	m.nextId += 1
	m.inverted[s] = id
	m.symtab[id] = s
	return id
}

type jsonValue interface {
	ToString() string
}

type objectValue struct {
	props map[uint]jsonValue
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
	id uint
}

func (v *stringValue) ToString() string {
	return "[String]"
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
	stack         []jsonValue
	memberStack   []string
	symtabMaker   *symtabMaker
	numObjects    int64
	numArrays     int64
	numPrimitives int64
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
		props: make(map[uint]jsonValue),
	})
	c.numObjects += 1
	c.memberStack = append(c.memberStack, "")
}

func (c *decoderClient) EndObject() {
	c.memberStack = c.memberStack[:len(c.memberStack)-1]
}

func (c *decoderClient) StartArray() {
	c.push(&arrayValue{})
	c.numArrays += 1
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
	symid := c.symtabMaker.getId(name)
	obj.props[symid] = v
}

func (c *decoderClient) StartValue() {
}

func (c *decoderClient) EndValue(next HasNext) {
	v := c.pop()
	arr := c.currentArray()
	arr.elems = append(arr.elems, v)
}

func (c *decoderClient) StringValue(s string) {
	value := s[1 : len(s)-1]
	id := c.symtabMaker.getId(value)
	c.push(&stringValue{id})
	c.numPrimitives += 1
}

func (c *decoderClient) NumberValue(s string) {
	n, _ := strconv.ParseFloat(s, 64)
	c.push(&numberValue{n})
	c.numPrimitives += 1
}

func (c *decoderClient) LiteralValue(l Literal) {
	c.push(&literalValue{l})
	c.numPrimitives += 1
}

type decodeResult struct {
	toplevel      jsonValue
	symtab        symbolTable
	numObjects    int64
	numArrays     int64
	numPrimitives int64
}

func Decode(r io.Reader) (*decodeResult, error) {
	c := &decoderClient{
		symtabMaker: newSymtabMaker(),
	}
	parser := NewParser(r, c)
	err := parser.Parse()
	if err != nil {
		return nil, err
	}
	if len(c.stack) != 1 {
		return nil, fmt.Errorf("Internal logic error: %d", len(c.stack))
	}
	result := &decodeResult{
		toplevel:      c.pop(),
		symtab:        c.symtabMaker.symtab,
		numObjects:    c.numObjects,
		numArrays:     c.numArrays,
		numPrimitives: c.numPrimitives,
	}
	c = nil
	runtime.GC()
	return result, nil
}
