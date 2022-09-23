package wall

func CheckForDuplications(f *FileNode) error {
	notypes := make(map[string]struct{}, 0)
	types := make(map[string]struct{}, 0)
	for _, def := range f.Defs {
		id := string(def.id())
		switch df := def.(type) {
		case *FunDef, *ImportDef, *ParsedImportDef:
			if _, ok := notypes[id]; ok {
				return NewError(df.pos(), "name is already declared: %s", id)
			}
			notypes[id] = struct{}{}
		case *StructDef:
			if _, ok := types[id]; ok {
				return NewError(df.pos(), "type is already declared: %s", id)
			}
			types[id] = struct{}{}
		default:
			panic("unreachable")
		}
	}
	return nil
}

func CheckImports(f *FileNode, m *Module) {
	checkedImports := make(map[*FileNode]*Module)
	checkedImports[f] = m
	importCheckingLoop(f, m, checkedImports)
}

func checkImports(f *FileNode, m *Module, checkedModules map[*FileNode]*Module) {
	if _, checked := checkedModules[f]; checked {
		return
	}
	checkedModules[f] = m
	importCheckingLoop(f, m, checkedModules)
}

func importCheckingLoop(f *FileNode, m *Module, checkedModules map[*FileNode]*Module) {
	for _, def := range f.Defs {
		switch importDef := def.(type) {
		case *ImportDef:
			panic("unparsed import def")
		case *ParsedImportDef:
			var module *Module
			if checkedModule, isChecked := checkedModules[importDef.ParsedNode]; isChecked {
				module = checkedModule
			} else {
				module = NewModule()
			}
			_ = m.AddImport(string(def.id()), module)
			checkImports(importDef.ParsedNode, module, checkedModules)
		}
	}
}

func CheckTypesSignatures(f *FileNode, m *Module) {
	checkTypesSignatures(f, m, make(map[*FileNode]struct{}))
}

func checkTypesSignatures(f *FileNode, m *Module, checkedNodes map[*FileNode]struct{}) {
	if _, checked := checkedNodes[f]; checked {
		return
	}
	checkedNodes[f] = struct{}{}
	for _, def := range f.Defs {
		switch df := def.(type) {
		case *ParsedImportDef:
			checkTypesSignatures(df.ParsedNode, m.GlobalScope.Import(string(df.id())), checkedNodes)
		case *StructDef:
			_ = m.AddType(string(df.id()), NewStructType())
		}
	}
}

type Type interface {
	typ()
}

type PointerType struct {
	To Type
}

type StructType struct {
	Fields map[string]Type
}

func NewStructType() *StructType {
	return &StructType{
		Fields: make(map[string]Type),
	}
}

type FunctionType struct {
	Args    []Type
	Returns Type
}

type ExternType struct {
	Import ImportId
	Type   TypeId
}

func (p *PointerType) typ()  {}
func (s *StructType) typ()   {}
func (f *FunctionType) typ() {}
func (e *ExternType) typ()   {}

type TypeId int
type ImportId int

type Module struct {
	Types       []Type
	Imports     []*Module
	GlobalScope *Scope
}

func NewModule() *Module {
	m := &Module{
		Types:       []Type{},
		Imports:     []*Module{},
		GlobalScope: NewScope(nil, nil),
	}
	m.GlobalScope.Module = m
	return m
}

func (m *Module) AddImport(name string, module *Module) ImportId {
	id := ImportId(len(m.Imports))
	m.Imports = append(m.Imports, module)
	m.GlobalScope.Imports[name] = id
	return id
}

func (m *Module) AddType(name string, typ Type) TypeId {
	id := TypeId(len(m.Types))
	m.Types = append(m.Types, typ)
	m.GlobalScope.Types[name] = id
	return id
}

type Scope struct {
	Parent  *Scope
	Module  *Module
	Types   map[string]TypeId
	Funs    map[string]TypeId
	Vars    map[string]TypeId
	Imports map[string]ImportId
}

func NewScope(parent *Scope, module *Module) *Scope {
	return &Scope{
		Parent:  parent,
		Module:  module,
		Types:   map[string]TypeId{},
		Funs:    map[string]TypeId{},
		Vars:    map[string]TypeId{},
		Imports: map[string]ImportId{},
	}
}

func (s *Scope) Import(name string) *Module {
	id, ok := s.Imports[name]
	if !ok {
		return nil
	}
	return s.Module.Imports[id]
}

func (s *Scope) Type(name string) Type {
	id, ok := s.Types[name]
	if !ok {
		return nil
	}
	return s.Module.Types[id]
}
