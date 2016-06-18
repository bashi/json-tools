package jsontools

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/peterh/liner"
)

type Path string

type stackItem struct {
	value jsonValue
	name  string
	path  string
}

type Inspector struct {
	toplevel jsonValue
	stack    []*stackItem
}

func NewInspector(r io.Reader) (*Inspector, error) {
	obj, err := Decode(r)
	if err != nil {
		return nil, err
	}
	stack := []*stackItem{
		{obj, "", ""},
	}
	return &Inspector{
		toplevel: obj,
		stack:    stack,
	}, nil
}

func (i *Inspector) current() *stackItem {
	return i.stack[len(i.stack)-1]
}

func (i *Inspector) pushMember(name string, value jsonValue) {
	path := i.current().path + "." + name
	i.stack = append(i.stack, &stackItem{
		value: value,
		name:  name,
		path:  path,
	})
}

func (i *Inspector) pushValue(index int, value jsonValue) {
	name := fmt.Sprintf("[%d]", index)
	path := i.current().path + name
	i.stack = append(i.stack, &stackItem{
		value: value,
		name:  name,
		path:  path,
	})
}

func (i *Inspector) pop() {
	l := len(i.stack)
	if l == 1 {
		return
	}
	i.stack = i.stack[:l-1]
}

func (i *Inspector) moveTo(path string) {
	if path == ".." {
		i.pop()
	}

	switch cur := i.current().value.(type) {
	case *objectValue:
		if value, ok := cur.props[path]; ok {
			i.pushMember(path, value)
		}
	case *arrayValue:
		index, err := strconv.Atoi(path)
		if err != nil || index < 0 || index >= len(cur.elems) {
			return
		}
		i.pushValue(index, cur.elems[index])
	}
}

func list(v jsonValue, limit int) {
	switch value := v.(type) {
	case *objectValue:
		for k, v := range value.props {
			fmt.Printf("%s: %s\n", k, v.ToString())
			limit -= 1
			if limit <= 0 {
				return
			}
		}
	case *arrayValue:
		for index, v := range value.elems {
			fmt.Printf("%d: %s\n", index, v.ToString())
			limit -= 1
			if limit <= 0 {
				return
			}
		}
	}
}

func show(v jsonValue) {
	pp.Println(v)
}

func (i *Inspector) doCommand(line string) error {
	if strings.HasPrefix(line, "ls") {
		list(i.current().value, 100)
	}
	if strings.HasPrefix(line, "cd") {
		i.moveTo(strings.TrimSpace(line[2:]))
	}
	if line == "show" {
		show(i.current().value)
	}
	return nil
}

func (i *Inspector) Repl() error {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	for {
		l, err := line.Prompt(i.current().path + "> ")
		if err != nil {
			// Map SIGINT to EOF
			if err == liner.ErrPromptAborted {
				return io.EOF
			}
			return err
		}
		if err := i.doCommand(l); err != nil {
			return err
		}
	}
	return nil
}
