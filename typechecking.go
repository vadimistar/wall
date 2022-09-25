package wall

import (
	"fmt"
	"reflect"
)

func Check(f *FileNode) (*Module, error) {
	if err := CheckForDuplications(f); err != nil {
		return nil, err
	}
	m := NewModule()
	CheckImports(f, m)
	CheckTypesSignatures(f, m)
	if err := CheckFunctionsSignatures(f, m); err != nil {
		return nil, err
	}
	if err := CheckTypesContents(f, m); err != nil {
		return nil, err
	}
	if err := CheckBlocks(f, m); err != nil {
		return nil, err
	}
	return m, nil
}

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
			_ = m.DefImport(string(def.id()), module)
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
			_ = m.DefType(string(df.id()), NewStructType())
		}
	}
}

func CheckFunctionsSignatures(f *FileNode, m *Module) error {
	return checkFunctionsSignatures(f, m, make(map[*FileNode]struct{}))
}

func checkFunctionsSignatures(f *FileNode, m *Module, checkedNodes map[*FileNode]struct{}) error {
	if _, checked := checkedNodes[f]; checked {
		return nil
	}
	checkedNodes[f] = struct{}{}
	for _, def := range f.Defs {
		switch df := def.(type) {
		case *ParsedImportDef:
			checkTypesSignatures(df.ParsedNode, m.GlobalScope.Import(string(df.id())), checkedNodes)
		case *FunDef:
			var argsTypes []TypeId
			for _, param := range df.Params {
				typeId, err := checkType(param.Type, m)
				if err != nil {
					return err
				}
				argsTypes = append(argsTypes, typeId)
			}
			var returnType TypeId
			if df.ReturnType != nil {
				var err error
				returnType, err = checkType(df.ReturnType, m)
				if err != nil {
					return err
				}
			} else {
				returnType = UNIT_TYPE_ID
			}
			m.GlobalScope.DefFun(string(df.id()), argsTypes, returnType)
		}
	}
	return nil
}

func CheckTypesContents(f *FileNode, m *Module) error {
	return checkTypesContents(f, m, make(map[*FileNode]struct{}))
}

func checkTypesContents(f *FileNode, m *Module, checkedNodes map[*FileNode]struct{}) error {
	if _, checked := checkedNodes[f]; checked {
		return nil
	}
	checkedNodes[f] = struct{}{}
	for _, def := range f.Defs {
		switch df := def.(type) {
		case *ParsedImportDef:
			checkTypesSignatures(df.ParsedNode, m.GlobalScope.Import(string(df.id())), checkedNodes)
		case *StructDef:
			fields := make(map[string]TypeId)
			for _, field := range df.Fields {
				if _, exists := fields[string(field.Name.Content)]; exists {
					return NewError(field.Name.Pos, "field is redeclared: %s", field.Name.Content)
				}
				typeId, err := checkType(field.Type, m)
				if err != nil {
					return err
				}
				fields[string(field.Name.Content)] = typeId
			}
			_, id := m.GlobalScope.Type(string(df.id()))
			switch t := m.Types[id].(type) {
			case *StructType:
				t.Fields = fields
			default:
				panic("unreachable")
			}
		}
	}
	return nil
}

func CheckBlocks(f *FileNode, m *Module) error {
	return checkBlocks(f, m, make(map[*FileNode]struct{}))
}

func checkBlocks(f *FileNode, m *Module, checkedNodes map[*FileNode]struct{}) error {
	if _, checked := checkedNodes[f]; checked {
		return nil
	}
	checkedNodes[f] = struct{}{}
	for _, def := range f.Defs {
		switch def := def.(type) {
		case *ParsedImportDef:
			checkTypesSignatures(def.ParsedNode, m.GlobalScope.Import(string(def.id())), checkedNodes)
		case *FunDef:
			funType := m.Types[m.GlobalScope.Funs[string(def.Id.Content)]].(*FunctionType)
			paramsScope := NewScope(m.GlobalScope, m)
			for _, param := range def.Params {
				t, err := checkType(param.Type, m)
				if err != nil {
					return err
				}
				paramsScope.DefVar(string(param.Id.Content), t)
			}
			_, _, err := checkBlock(def.Body, paramsScope, &MustReturn{
				TypeId: funType.Returns,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkBlock(block *BlockStmt, scope *Scope, controlFlow ControlFlow) (typeid TypeId, returns TypeId, err error) {
	scope = NewScope(scope, scope.Module)
	if len(block.Stmts) == 0 {
		switch controlFlow := controlFlow.(type) {
		case *MustReturn:
			if controlFlow.TypeId != UNIT_TYPE_ID {
				return -1, -1, NewError(block.pos(), "expected %s, but %s is returned", scope.Module.typeIdAsStr(controlFlow.TypeId), scope.Module.typeIdAsStr(UNIT_TYPE_ID))
			}
			return UNIT_TYPE_ID, UNIT_TYPE_ID, nil
		}
		return UNIT_TYPE_ID, -1, nil
	}
	returns = -1
	for i, stmt := range block.Stmts {
		var cf ControlFlow
		if i >= len(block.Stmts)-1 {
			cf = controlFlow
		} else {
			cf = &MayReturn{
				TypeId: controlFlow.typeId(),
			}
		}
		t, currentReturns, err := CheckStmt(stmt, scope, cf)
		if err != nil {
			return -1, -1, err
		}
		if currentReturns >= 0 {
			if currentReturns != controlFlow.typeId() {
				return -1, -1, NewError(stmt.pos(), "unexpected return type: %s (%s is expected)", scope.Module.typeIdAsStr(currentReturns), scope.Module.typeIdAsStr(controlFlow.typeId()))
			}
			returns = currentReturns
		}
		if i >= len(block.Stmts)-1 {
			switch controlFlow := controlFlow.(type) {
			case *MustReturn:
				if returns < 0 {
					if controlFlow.TypeId != UNIT_TYPE_ID {
						return -1, -1, NewError(stmt.pos(), "expected return statement")
					}
					return t, UNIT_TYPE_ID, nil
				}
			}
			return t, returns, nil
		}
	}
	panic("unreachable")
}

// Checks a statement node
//
// The 'returns' return variable is a valid TypeId (>= 0), if the provided statement returns a value
func CheckStmt(stmt StmtNode, scope *Scope, controlFlow ControlFlow) (typeId TypeId, returns TypeId, err error) {
	switch stmt := stmt.(type) {
	case *BlockStmt:
		return checkBlock(stmt, scope, controlFlow)
	case *VarStmt:
		typ, err := CheckExpr(stmt.Value, scope)
		if err != nil {
			return -1, -1, err
		}
		scope.DefVar(string(stmt.Id.Content), typ)
	case *ExprStmt:
		_, err := CheckExpr(stmt.Expr, scope)
		if err != nil {
			return -1, -1, err
		}
		return UNIT_TYPE_ID, -1, nil
	case *ReturnStmt:
		var argType TypeId
		if stmt.Arg == nil {
			argType = UNIT_TYPE_ID
		} else {
			t, err := CheckExpr(stmt.Arg, scope)
			if err != nil {
				return -1, -1, err
			}
			argType = t
		}
		if argType != controlFlow.typeId() {
			return -1, -1, NewError(stmt.pos(), "unexpected return type: %s (%s is expected)", scope.Module.typeIdAsStr(argType), scope.Module.typeIdAsStr(controlFlow.typeId()))
		}
		return UNIT_TYPE_ID, argType, nil
	default:
		panic("unreachable")
	}
	return UNIT_TYPE_ID, -1, nil
}

func CheckExpr(expr ExprNode, scope *Scope) (TypeId, error) {
	switch expr := expr.(type) {
	case *UnaryExprNode:
		return checkUnaryExpr(expr, scope)
	case *BinaryExprNode:
		return checkBinaryExpr(expr, scope)
	case *GroupedExprNode:
		return CheckExpr(expr.Inner, scope)
	case *LiteralExprNode:
		return checkLiteralExpr(expr, scope)
	}
	panic("unreachable")
}

func checkUnaryExpr(expr *UnaryExprNode, scope *Scope) (TypeId, error) {
	inner, err := CheckExpr(expr.Operand, scope)
	if err != nil {
		return -1, err
	}
	switch expr.Operator.Kind {
	case PLUS:
		return inner, nil
	case MINUS:
		if canBeNegated(inner, scope.Module) {
			return inner, nil
		}
		return -1, NewError(expr.Operator.Pos, "trait %s is not implemented for %s", NEGATE_OPERATOR_TRAIT_NAME, scope.Module.typeIdAsStr(inner))
	}
	panic("unreachable")
}

func checkBinaryExpr(expr *BinaryExprNode, scope *Scope) (TypeId, error) {
	left, err := CheckExpr(expr.Left, scope)
	if err != nil {
		return -1, err
	}
	right, err := CheckExpr(expr.Right, scope)
	if err != nil {
		return -1, err
	}
	if left != right {
		return -1, NewError(expr.Op.Pos, "types of left and right operands are not the same (%s, %s)", scope.Module.typeIdAsStr(left), scope.Module.typeIdAsStr(right))
	}
	switch expr.Op.Kind {
	case PLUS:
		if !traitIsImplemented(ADD_OPERATOR_TRAIT_NAME, left, scope.Module) {
			return -1, NewError(expr.Op.Pos, "trait %s is not implemented for type %s", ADD_OPERATOR_TRAIT_NAME, scope.Module.typeIdAsStr(left))
		}
		return left, nil
	case MINUS:
		if !traitIsImplemented(SUBTRACT_OPERATOR_TRAIT_NAME, left, scope.Module) {
			return -1, NewError(expr.Op.Pos, "trait %s is not implemented for type %s", SUBTRACT_OPERATOR_TRAIT_NAME, scope.Module.typeIdAsStr(left))
		}
		return left, nil
	case STAR:
		if !traitIsImplemented(MULTIPLY_OPERATOR_TRAIT_NAME, left, scope.Module) {
			return -1, NewError(expr.Op.Pos, "trait %s is not implemented for type %s", MULTIPLY_OPERATOR_TRAIT_NAME, scope.Module.typeIdAsStr(left))
		}
		return left, nil
	case SLASH:
		if !traitIsImplemented(DIVIDE_OPERATOR_TRAIT_NAME, left, scope.Module) {
			return -1, NewError(expr.Op.Pos, "trait %s is not implemented for type %s", DIVIDE_OPERATOR_TRAIT_NAME, scope.Module.typeIdAsStr(left))
		}
		return left, nil
	}
	panic("unreachable")
}

func checkLiteralExpr(expr *LiteralExprNode, scope *Scope) (TypeId, error) {
	switch expr.Kind {
	case INTEGER:
		return INT_TYPE_ID, nil
	case FLOAT:
		return FLOAT_TYPE_ID, nil
	case IDENTIFIER:
		if typeId := scope.findVar(string(expr.Content)); typeId != nil {
			return *typeId, nil
		}
		return -1, NewError(expr.pos(), "undeclared: %s", expr.Token.Content)
	}
	panic("unimplemented")
}

func (m *Module) typeIdAsStr(id TypeId) string {
	switch typ := m.Types[id].(type) {
	case *PointerType:
		return "*" + m.typeIdAsStr(typ.To)
	case *StructType:
		if name := m.GlobalScope.findNameOfTypeId(id); name != nil {
			return *name
		}
		panic("invalid type id")
	case *FunctionType:
		params := ""
		for i, param := range typ.Args {
			params += m.typeIdAsStr(param)
			if i < len(typ.Args)-1 {
				params += ", "
			}
		}
		return fmt.Sprintf("fun (%s) %s", params, m.typeIdAsStr(typ.Returns))
	case *ExternType:
		importName := m.GlobalScope.findNameOfImportId(typ.Import)
		if importName == nil {
			panic("invalid import id")
		}
		importedType := m.Imports[typ.Import].typeIdAsStr(typ.Type)
		return fmt.Sprintf("%s.%s", *importName, importedType)
	case *BuildinType:
		switch id {
		case UNIT_TYPE_ID:
			return "()"
		case INT_TYPE_ID:
			return "int"
		case FLOAT_TYPE_ID:
			return "float"
		}
	}
	panic("unreachable")
}

const NEGATE_OPERATOR_TRAIT_NAME = "Negate"
const ADD_OPERATOR_TRAIT_NAME = "Add"
const SUBTRACT_OPERATOR_TRAIT_NAME = "Subtract"
const MULTIPLY_OPERATOR_TRAIT_NAME = "Multiply"
const DIVIDE_OPERATOR_TRAIT_NAME = "Divide"

func canBeNegated(inner TypeId, module *Module) bool {
	return traitIsImplemented(NEGATE_OPERATOR_TRAIT_NAME, inner, module)
}

func traitIsImplemented(name string, forType TypeId, module *Module) bool {
	switch name {
	case NEGATE_OPERATOR_TRAIT_NAME, ADD_OPERATOR_TRAIT_NAME, SUBTRACT_OPERATOR_TRAIT_NAME, MULTIPLY_OPERATOR_TRAIT_NAME, DIVIDE_OPERATOR_TRAIT_NAME:
		return forType == INT_TYPE_ID || forType == FLOAT_TYPE_ID
	}
	return false
}

func checkType(node TypeNode, module *Module) (TypeId, error) {
	switch nd := node.(type) {
	case *IdTypeNode:
		typ, id := module.GlobalScope.Type(string(nd.Content))
		if typ == nil {
			return -1, NewError(nd.pos(), "undeclared type: %s", nd.Content)
		}
		return id, nil
	}
	panic("unreachable")
}

type Type interface {
	typ()
}

type PointerType struct {
	To TypeId
}

type StructType struct {
	Fields map[string]TypeId
}

func NewStructType() *StructType {
	return &StructType{
		Fields: make(map[string]TypeId),
	}
}

type FunctionType struct {
	Args    []TypeId
	Returns TypeId
}

type ExternType struct {
	Import ImportId
	Type   TypeId
}

type BuildinType struct{}

const (
	UNIT_TYPE_ID TypeId = iota
	INT_TYPE_ID
	FLOAT_TYPE_ID
)

func (p *PointerType) typ()  {}
func (s *StructType) typ()   {}
func (f *FunctionType) typ() {}
func (e *ExternType) typ()   {}
func (b *BuildinType) typ()  {}

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
	m.DefType("()", &BuildinType{})
	m.DefType("int", &BuildinType{})
	m.DefType("float", &BuildinType{})
	return m
}

func (m *Module) DefImport(name string, module *Module) ImportId {
	id := ImportId(len(m.Imports))
	m.Imports = append(m.Imports, module)
	m.GlobalScope.Imports[name] = id
	return id
}

func (m *Module) DefType(name string, typ Type) TypeId {
	id := TypeId(len(m.Types))
	m.Types = append(m.Types, typ)
	m.GlobalScope.Types[name] = id
	return id
}

func (m *Module) TypeId(typ Type) TypeId {
	for id, t := range m.Types {
		if reflect.DeepEqual(t, typ) {
			return TypeId(id)
		}
	}
	id := TypeId(len(m.Types))
	m.Types = append(m.Types, typ)
	return id
}

type Scope struct {
	Parent   *Scope
	Children []*Scope
	Module   *Module
	Types    map[string]TypeId
	Funs     map[string]TypeId
	Vars     map[string]TypeId
	Imports  map[string]ImportId
}

type ControlFlow interface {
	controlFlow()
	typeId() TypeId
}

type MustReturn struct {
	TypeId
}
type MayReturn struct {
	TypeId
}

func (m *MustReturn) controlFlow() {}
func (m *MayReturn) controlFlow()  {}

func (m *MustReturn) typeId() TypeId {
	return m.TypeId
}
func (m *MayReturn) typeId() TypeId {
	return m.TypeId
}

func NewScope(parent *Scope, module *Module) *Scope {
	s := &Scope{
		Parent:   parent,
		Children: make([]*Scope, 0),
		Module:   module,
		Types:    map[string]TypeId{},
		Funs:     map[string]TypeId{},
		Vars:     map[string]TypeId{},
		Imports:  map[string]ImportId{},
	}
	if parent != nil {
		parent.Children = append(parent.Children, s)
	}
	return s
}

func (s *Scope) Import(name string) *Module {
	id, ok := s.Imports[name]
	if !ok {
		return nil
	}
	return s.Module.Imports[id]
}

func (s *Scope) Type(name string) (Type, TypeId) {
	id, ok := s.Types[name]
	if !ok {
		return nil, -1
	}
	return s.Module.Types[id], id
}

func (s *Scope) DefFun(name string, args []TypeId, returns TypeId) {
	typeId := s.Module.TypeId(&FunctionType{
		Args:    args,
		Returns: returns,
	})
	s.Funs[name] = typeId
}

func (s *Scope) DefVar(name string, typeId TypeId) {
	s.Vars[name] = typeId
}

func (s *Scope) findNameOfTypeId(id TypeId) *string {
	for name, typ := range s.Types {
		if typ == id {
			return &name
		}
	}
	for _, child := range s.Children {
		if n := child.findNameOfTypeId(id); n != nil {
			return n
		}
	}
	return nil
}

func (s *Scope) findNameOfImportId(id ImportId) *string {
	for name, typ := range s.Imports {
		if typ == id {
			return &name
		}
	}
	for _, child := range s.Children {
		if n := child.findNameOfImportId(id); n != nil {
			return n
		}
	}
	return nil
}

func (s *Scope) findVar(name string) *TypeId {
	if typeId, ok := s.Vars[name]; ok {
		return &typeId
	}
	if s.Parent != nil {
		return s.Parent.findVar(name)
	}
	return nil
}
