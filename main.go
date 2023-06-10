package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"
)

func main() {
	if err := run(); err != nil {
		_ = fmt.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("filename is required")
	}
	fileName := os.Args[1]
	tmpl, err := template.ParseFiles(fileName)
	if err != nil {
		return err
	}
	parameters := os.Args[2:]
	if len(parameters) == 0 {
		createCommand(tmpl)
		return nil
	}
	values, err := makeValues(parameters)
	if err != nil {
		return err
	}
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	output := filepath.Join(path, "gen_"+filepath.Base(fileName))
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, values)
	if err != nil {
		return err
	}
	return nil
}

func makeValues(parameters []string) (map[string]string, error) {
	ret := make(map[string]string)
	for _, param := range parameters {
		pair := strings.Split(param, "=")
		if len(pair) < 2 {
			return nil, errors.New("= must be included")
		}
		key := pair[0]
		value := pair[1]
		if _, ok := ret[key]; ok {
			return nil, errors.New(fmt.Sprintf("define duplicate '%s'", key))
		}
		ret[key] = value
	}
	return ret, nil
}

func createCommand(tmpl *template.Template) {
	fields := Fields{set: NewSet()}
	fields.extractField(tmpl.Tree.Root)

	var args []string
	for _, name := range fields.names {
		args = append(args, name+"=")
	}
	command := []string{os.Args[0], os.Args[1]}
	command = append(command, args...)

	fmt.Printf("%s\n", strings.Join(command, " "))
}

type Set struct {
	set    map[string]struct{}
	member struct{}
}

func (s *Set) add(str string) {
	s.set[str] = s.member
}

func (s *Set) find(str string) bool {
	_, ok := s.set[str]
	return ok
}

func NewSet() *Set {
	return &Set{set: make(map[string]struct{})}
}

type Fields struct {
	set   *Set
	names []string // to keep the order
}

func (f *Fields) extractField(node parse.Node) {
	switch n := node.(type) {
	case *parse.ListNode:
		for _, nn := range n.Nodes {
			f.extractField(nn)
		}
	case *parse.ActionNode:
		f.extractField(n.Pipe)
	case *parse.PipeNode:
		for _, nn := range n.Cmds {
			f.extractField(nn)
		}
	case *parse.CommandNode:
		for _, nn := range n.Args {
			f.extractField(nn)
		}
	case *parse.FieldNode:
		if len(n.Ident) > 1 {
			panic("multiple dots is not implemented")
		}
		name := strings.Join(n.Ident, ".")
		if f.set.find(name) {
			return
		}
		f.set.add(name)
		f.names = append(f.names, name)
	}
}
