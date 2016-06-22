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
	json  *decodeResult
	stack []*stackItem
}

func NewInspector(r io.Reader) (*Inspector, error) {
	json, err := Decode(r)
	if err != nil {
		return nil, err
	}
	stack := []*stackItem{
		{json.toplevel, "", ""},
	}
	return &Inspector{
		json:  json,
		stack: stack,
	}, nil
}

func (i *Inspector) findMember(obj *objectValue, name string) jsonValue {
	for symid, v := range obj.props {
		propName := i.json.symtab[symid]
		if propName == name {
			return v
		}
	}
	return nil
}

func (i *Inspector) idToStr(id uint) string {
	return i.json.symtab[id]
}

func (i *Inspector) valueToString(v jsonValue) string {
	switch value := v.(type) {
	case *stringValue:
		return i.idToStr(value.id)
	default:
		return v.ToString()
	}
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
		if value := i.findMember(cur, name); value != nil {
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

func (i *Inspector) list(v jsonValue) {
	switch value := v.(type) {
	case *objectValue:
		for k, v := range value.props {
			fmt.Printf("%s: %s\n", i.idToStr(k), i.valueToString(v))
		}
	case *arrayValue:
		for index, v := range value.elems {
			fmt.Printf("%d: %s\n", index, i.valueToString(v))
		}
	}
}

var literalColor = color.New(color.FgCyan)
var valueColor = color.New(color.FgBlue)
var memberColor = color.New(color.FgMagenta)

func (i *Inspector) printValue(
	v jsonValue, depth int, rows int, indent string) {
	switch value := v.(type) {
	case *literalValue:
		literalColor.Printf("%s", i.valueToString(value))
	case *stringValue, *numberValue:
		valueColor.Printf("%s", i.valueToString(value))
	case *objectValue:
		if depth <= 0 || rows <= 0 {
			fmt.Printf("%s", i.valueToString(value))
		} else {
			fmt.Printf("{")
			innerIndent := indent + "  "
			count := rows
			prefix := "\n"
			for k, m := range value.props {
				fmt.Printf("%s%s", prefix, innerIndent)
				memberColor.Printf("%s", i.idToStr(k))
				fmt.Printf(": ")
				i.printValue(m, depth-1, rows/2, innerIndent)
				count -= 1
				if count <= 0 {
					fmt.Printf("%s%s...", prefix, innerIndent)
					break
				}
				prefix = ",\n"
			}
			fmt.Printf("\n%s}", indent)
		}
	case *arrayValue:
		if depth <= 0 || rows <= 0 {
			fmt.Printf("%s", i.valueToString(value))
		} else {
			fmt.Printf("[")
			innerIndent := indent + "  "
			prefix := "\n"
			for count, e := range value.elems {
				fmt.Printf("%s%s", prefix, innerIndent)
				i.printValue(e, depth-1, rows/2, innerIndent)
				if rows-count <= 0 {
					fmt.Printf("%s%s...", prefix, innerIndent)
					break
				}
				prefix = ",\n"
			}
			fmt.Printf("\n%s]", indent)
		}
	default:
		fmt.Printf("%s", i.valueToString(value))
	}
}

func (i *Inspector) show(v jsonValue) {
	i.printValue(v, 4, 64, "")
	fmt.Println()
}

var metaColor = color.New(color.FgGreen)

func (i *Inspector) showMetadata() {
	if len(i.stack) == 1 {
		fmt.Printf("# objects = %d, arrays = %d, primitives = %d\n",
			i.json.numObjects, i.json.numArrays, i.json.numPrimitives)
	}
	switch value := i.current().value.(type) {
	case *objectValue:
		metaColor.Printf("[Object] size = %d\n", len(value.props))
	case *arrayValue:
		metaColor.Printf("[Array] size = %d\n", len(value.elems))
	default:
		metaColor.Printf("%s\n", i.valueToString(value))
	}
}

func (i *Inspector) doCommand(line string) error {
	if strings.HasPrefix(line, "ls") {
		i.list(i.current().value)
	} else if strings.HasPrefix(line, "cd") {
		pathStr := strings.TrimSpace(line[2:])
		path := strings.Split(pathStr, "/")
		i.moveTo(path)
	} else if line == "show" {
		i.show(i.current().value)
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
		i.showMetadata()
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
