package jsontools

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fatih/color"
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

func (i *Inspector) moveTo(path []string) {
	if len(path) == 0 {
		return
	}

	name := path[0]
	rest := path[1:]
	if name == "." {
		i.moveTo(rest)
		return
	}
	if name == ".." {
		i.pop()
		i.moveTo(rest)
		return
	}

	switch cur := i.current().value.(type) {
	case *objectValue:
		if value, ok := cur.props[name]; ok {
			i.pushMember(name, value)
		}
	case *arrayValue:
		index, err := strconv.Atoi(name)
		if err != nil || index < 0 || index >= len(cur.elems) {
			return
		}
		i.pushValue(index, cur.elems[index])
	}
	i.moveTo(rest)
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

var literalColor = color.New(color.FgCyan)
var valueColor = color.New(color.FgBlue)
var memberColor = color.New(color.FgMagenta)

func printValue(v jsonValue, depth int, rows int, indent string) {
	switch value := v.(type) {
	case *literalValue:
		literalColor.Printf("%s", value.ToString())
	case *stringValue, *numberValue:
		valueColor.Printf("%s", value.ToString())
	case *objectValue:
		if depth <= 0 || rows <= 0 {
			fmt.Printf("%s", value.ToString())
		} else {
			fmt.Printf("{")
			innerIndent := indent + "  "
			i := rows
			prefix := "\n"
			for k, m := range value.props {
				fmt.Printf("%s%s", prefix, innerIndent)
				memberColor.Printf("%s", k)
				fmt.Printf(": ")
				printValue(m, depth-1, rows/2, innerIndent)
				i -= 1
				if i <= 0 {
					fmt.Printf("%s%s...", prefix, innerIndent)
					break
				}
				prefix = ",\n"
			}
			fmt.Printf("\n%s}", indent)
		}
	case *arrayValue:
		if depth <= 0 || rows <= 0 {
			fmt.Printf("%s", value.ToString())
		} else {
			fmt.Printf("[")
			innerIndent := indent + "  "
			prefix := "\n"
			for i, e := range value.elems {
				fmt.Printf("%s%s", prefix, innerIndent)
				printValue(e, depth-1, rows/2, innerIndent)
				if rows-i <= 0 {
					fmt.Printf("%s%s...", prefix, innerIndent)
					break
				}
				prefix = ",\n"
			}
			fmt.Printf("\n%s]", indent)
		}
	default:
		fmt.Printf("%s", value.ToString())
	}
}

func show(v jsonValue) {
	printValue(v, 4, 64, "")
	fmt.Println()
}

func (i *Inspector) doCommand(line string) error {
	if strings.HasPrefix(line, "ls") {
		list(i.current().value, 100)
	} else if strings.HasPrefix(line, "cd") {
		pathStr := strings.TrimSpace(line[2:])
		path := strings.Split(pathStr, "/")
		i.moveTo(path)
	} else if line == "show" {
		show(i.current().value)
	} else {
		fmt.Printf("Unrecognized: %s\n", line)
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
