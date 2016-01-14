package jsonlib

import (
	"fmt"
	"io"
	"text/scanner"
)

type Literal int

const (
	False = iota
	Null
	True
)

func (l Literal) String() string {
	switch l {
	case False:
		return "false"
	case Null:
		return "null"
	case True:
		return "true"
	default:
		return "unknown"
	}
}

type HasNext bool

type ParserClient interface {
	StartObject()
	EndObject()
	StartArray()
	EndArray()
	StartMember(string)
	EndMember(HasNext)
	StartValue()
	EndValue(HasNext)
	StringValue(string)
	NumberValue(string)
	LiteralValue(Literal)
}

type ParserClientBase struct {
}

func (p *ParserClientBase) StartObject()         {}
func (p *ParserClientBase) EndObject()           {}
func (p *ParserClientBase) StartArray()          {}
func (p *ParserClientBase) EndArray()            {}
func (p *ParserClientBase) StartMember(string)   {}
func (p *ParserClientBase) EndMember(HasNext)    {}
func (p *ParserClientBase) StartValue()          {}
func (p *ParserClientBase) EndValue(HasNext)     {}
func (p *ParserClientBase) StringValue(string)   {}
func (p *ParserClientBase) NumberValue(string)   {}
func (p *ParserClientBase) LiteralValue(Literal) {}

type Parser struct {
	s *scanner.Scanner
	c ParserClient
}

func NewParser(r io.Reader, c ParserClient) *Parser {
	s := new(scanner.Scanner).Init(r)
	s.Mode = (scanner.ScanIdents |
		scanner.ScanInts |
		scanner.ScanFloats |
		scanner.SkipComments |
		scanner.ScanStrings)
	return &Parser{
		s: s,
		c: c,
	}
}

type ParseError struct {
	Message string
	Pos     scanner.Position
}

func (e *ParseError) Error() string {
	return e.Pos.String() + ": " + e.Message
}

func (p *Parser) createError(format string, args ...interface{}) error {
	return &ParseError{
		Message: fmt.Sprintf(format, args...),
		Pos:     p.s.Pos(),
	}
}

func (p *Parser) Parse() error {
	tok := p.s.Scan()
	if tok == scanner.EOF {
		return nil
	}
	var err error
	switch tok {
	case '{':
		err = p.parseObject()
	case '[':
		err = p.parseArray()
	default:
		err = p.createError("expected object or array, but got %s",
			scanner.TokenString(tok))
	}
	return err
}

func (p *Parser) parseObject() error {
	p.c.StartObject()
	for {
		tok := p.s.Scan()
		if tok == '}' {
			p.c.EndObject()
			return nil
		}

		if tok != scanner.String {
			return p.createError("expected string, but got %s ",
				scanner.TokenString(tok))
		}
		p.c.StartMember(p.s.TokenText())

		tok = p.s.Scan()
		if tok != ':' {
			return p.createError("expected ':', but got %s",
				scanner.TokenString(tok))
		}
		if err := p.parseValue(); err != nil {
			return err
		}

		tok = p.s.Scan()
		if tok == '}' {
			p.c.EndMember(false)
			p.c.EndObject()
			return nil
		}
		if tok != ',' {
			return p.createError("in object, expected ',', but got %s",
				scanner.TokenString(tok))
		}
		p.c.EndMember(true)
	}
	// NOTREACHED
	return nil
}

func (p *Parser) parseArray() error {
	p.c.StartArray()
	for {
		tok := p.s.Peek()
		if tok == ']' {
			p.s.Scan()
			p.c.EndArray()
			return nil
		}
		p.c.StartValue()
		if err := p.parseValue(); err != nil {
			return err
		}

		tok = p.s.Scan()
		if tok == ']' {
			p.c.EndValue(false)
			p.c.EndArray()
			return nil
		}
		if tok != ',' {
			return p.createError("in array, expected ',', but got %s",
				scanner.TokenString(tok))
		}
		p.c.EndValue(true)
	}
	return nil
}

func (p *Parser) parseValue() error {
	tok := p.s.Scan()
	if tok == '{' {
		return p.parseObject()
	} else if tok == '[' {
		return p.parseArray()
	} else if tok == scanner.String {
		p.c.StringValue(p.s.TokenText())
		return nil
	} else if tok == '-' {
		tok = p.s.Scan()
		if tok == scanner.Int || tok == scanner.Float {
			p.c.NumberValue("-" + p.s.TokenText())
			return nil
		}
		return p.createError("expected number, but got %s",
			scanner.TokenString(tok))
	} else if tok == scanner.Int || tok == scanner.Float {
		p.c.NumberValue(p.s.TokenText())
		return nil
	} else if tok == scanner.Ident && isLiteral(p.s.TokenText()) {
		p.c.LiteralValue(toLiteral(p.s.TokenText()))
		return nil
	}
	return p.createError("expected object, array, string, number or literal, but got %s", scanner.TokenString(tok))
}

func isLiteral(s string) bool {
	return s == "false" || s == "null" || s == "true"
}

func toLiteral(s string) Literal {
	switch s {
	case "false":
		return False
	case "null":
		return Null
	case "true":
		return True
	}
	panic("Unsupported literal")
}
