package wall

import (
	"os"
	"path/filepath"
	"strings"
)

func NameNodeFromFile(file *FileNode) (NameNode, map[string]*FileNode, error) {
	namedNodes := make(map[string]NameNode)
	parsedFiles := make(map[string]*FileNode)
	node, err := nameNodeFromFile(file, namedNodes, parsedFiles)
	if err != nil {
		return nil, nil, err
	}
	return node, parsedFiles, nil
}

func FindDuplicates(nameNode NameNode) error {
	return findDuplicates(nameNode, nameNode)
}

func findDuplicates(node, start NameNode) error {
	if node == start {
		return nil
	}
	names := make(map[string]struct{})
	for _, child := range node.Children() {
		if _, ok := names[child.Name()]; ok {
			return NewError(child.Pos(), "already declared: %s", child.Name())
		}
		names[child.Name()] = struct{}{}
	}
	for _, child := range node.Children() {
		if err := findDuplicates(child, start); err != nil {
			return err
		}
	}
	return nil
}

func ResolveTypeSignatures(nameNode NameNode, parsedFiles map[string]*FileNode) {
	resolveTypeSignatures(nameNode, parsedFiles, nameNode)
}

func resolveTypeSignatures(nameNode NameNode, parsedFiles map[string]*FileNode, startNode NameNode) {
	if nameNode == startNode {
		return
	}
	switch node := nameNode.(type) {
	case *ModuleNameNode:
		{
		}
	case *FunctionNameNode:
		{
			_ = node
		}
	default:
		panic("unimplemented")
	}
	for _, child := range nameNode.Children() {
		resolveTypeSignatures(child, parsedFiles, startNode)
	}
}

type NameNode interface {
	Name() string
	Pos() Pos
	Children() []NameNode
}

type ModuleNameNode struct {
	Naming   string
	pos      Pos
	children []NameNode
	Path     string
}

type FunctionNameNode struct {
	Naming     string
	pos        Pos
	children   []NameNode
	Args       []string
	ReturnType string
}

func (m *ModuleNameNode) Name() string {
	return m.Naming
}
func (f *FunctionNameNode) Name() string {
	return f.Naming
}

func (m *ModuleNameNode) Pos() Pos {
	return m.pos
}
func (f *FunctionNameNode) Pos() Pos {
	return f.pos
}

func (m *ModuleNameNode) Children() []NameNode {
	return m.children
}
func (f *FunctionNameNode) Children() []NameNode {
	return f.children
}

func nameNodeFromFile(file *FileNode, nameNodes map[string]NameNode, parsedFiles map[string]*FileNode) (NameNode, error) {
	path, err := filepath.Abs(file.pos().Filename)
	if err != nil {
		return nil, err
	}
	moduleName := &ModuleNameNode{
		Naming:   moduleName(file.pos().Filename),
		pos:      file.pos(),
		Path:     path,
		children: make([]NameNode, 0, len(file.Defs)),
	}
	nameNodes[path] = moduleName
	parsedFiles[path] = file
	// handle everything except imports
	for _, def := range file.Defs {
		switch df := def.(type) {
		case *FunDef:
			newNode := &FunctionNameNode{
				Naming: string(df.Id.Content),
				pos:    df.pos(),
			}
			moduleName.children = append(moduleName.children, newNode)
		case *ImportDef:
			break
		default:
			panic("unreachable")
		}
	}
	// handle only imports
	for _, def := range file.Defs {
		switch df := def.(type) {
		case *ImportDef:
			name := string(df.Name.Content)
			path := pathOfModule(name, file.pos().Filename)
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			if node, ok := nameNodes[absPath]; ok {
				moduleName.children = append(moduleName.children, node)
				continue
			}
			nameNode, err := nameNodeFromPath(absPath, nameNodes, parsedFiles)
			if err != nil {
				return nil, err
			}
			moduleName.children = append(moduleName.children, nameNode)
		}
	}
	return moduleName, nil
}

func nameNodeFromPath(path string, nameNodes map[string]NameNode, parsedFiles map[string]*FileNode) (NameNode, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsedFile, err := ParseFile(path, source)
	if err != nil {
		return nil, err
	}
	parsedFiles[path] = parsedFile
	return nameNodeFromFile(parsedFile, nameNodes, parsedFiles)
}

func moduleName(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(path)
	return strings.TrimSuffix(name, ext)
}

const WALL_EXTENSION string = ".wl"

func pathOfModule(name, path string) string {
	dir := filepath.Dir(path)
	module := name + WALL_EXTENSION
	return filepath.Join(dir, module)
}
