package wall

import (
	"os"
	"path/filepath"
	"strings"
)

func NameNodeFromFile(file *FileNode) (NameNode, []*FileNode, error) {
	namedNodes := make(map[string]NameNode)
	parsedFiles := make([]*FileNode, 0)
	node, err := nameNodeFromFile(file, namedNodes, &parsedFiles)
	if err != nil {
		return nil, nil, err
	}
	return node, parsedFiles, nil
}

func nameNodeFromFile(file *FileNode, nameNodes map[string]NameNode, parsedFiles *[]*FileNode) (NameNode, error) {
	/* a.wl
	import B
	fun a() {
	}
		 ^
		 |
		 |
	A
	| \
	B  a
	*/
	moduleName := &ModuleNameNode{
		Naming:   moduleName(file.pos().Filename),
		Path:     file.pos().Filename,
		children: make([]NameNode, 0, len(file.Defs)),
	}
	path, err := filepath.Abs(moduleName.Path)
	if err != nil {
		return nil, err
	}
	nameNodes[path] = moduleName
	// handle everything except imports
	for _, def := range file.Defs {
		switch df := def.(type) {
		case *FunDef:
			newNode := &FunctionNameNode{
				Naming: string(df.Id.Content),
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

func nameNodeFromPath(path string, nameNodes map[string]NameNode, parsedFiles *[]*FileNode) (NameNode, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsedFile, err := ParseFile(path, source)
	if err != nil {
		return nil, err
	}
	*parsedFiles = append(*parsedFiles, parsedFile)
	return nameNodeFromFile(parsedFile, nameNodes, parsedFiles)
}

type NameNode interface {
	Name() string
	Children() []NameNode
}

type ModuleNameNode struct {
	Naming   string
	children []NameNode
	Path     string
}

type FunctionNameNode struct {
	Naming     string
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

func (m *ModuleNameNode) Children() []NameNode {
	return m.children
}
func (f *FunctionNameNode) Children() []NameNode {
	return f.children
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
