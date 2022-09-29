package wall

import (
	"bytes"
	"reflect"
)

func CheckCompilationUnit(f *ParsedFile) (*CheckedFile, error) {
	checkedFile := NewCheckedFile(f.pos().Filename)
	if err := CheckImports(f, checkedFile); err != nil {
		return nil, err
	}
	if err := CheckTypeSignatures(f, checkedFile); err != nil {
		return nil, err
	}
	if err := CheckFunctionSignatures(f, checkedFile); err != nil {
		return nil, err
	}
	if err := CheckTypeContents(f, checkedFile); err != nil {
		return nil, err
	}
	if err := CheckBlocks(f, checkedFile); err != nil {
		return nil, err
	}
	return checkedFile, nil
}

func CheckImports(f *ParsedFile, c *CheckedFile) error {
	return checkImports(f, c, make(map[*ParsedFile]*CheckedFile))
}

func checkImports(f *ParsedFile, c *CheckedFile, checkedFiles map[*ParsedFile]*CheckedFile) error {
	if _, checked := checkedFiles[f]; checked {
		return nil
	}
	checkedFiles[f] = c
	for _, def := range f.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			if def.File == nil {
				panic("unresolved import")
			}
			var checkedFile *CheckedFile
			if file, checked := checkedFiles[def.File]; checked {
				checkedFile = file
			} else {
				checkedFile = NewCheckedFile(def.File.pos().Filename)
			}
			checkedImport := &CheckedImport{
				Name: def.Name,
				File: checkedFile,
			}
			c.GlobalScope.Import(checkedImport)
			checkImports(def.File, checkedImport.File, checkedFiles)
		}
	}
	return nil
}

func CheckTypeSignatures(p *ParsedFile, c *CheckedFile) error {
	return checkTypeSignatures(p, c, make(map[*ParsedFile]struct{}))
}

func checkTypeSignatures(p *ParsedFile, c *CheckedFile, checkedFiles map[*ParsedFile]struct{}) error {
	if isChecked(p, checkedFiles) {
		return nil
	}
	for _, def := range p.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			if err := handleImport(def, c, checkedFiles, checkTypeSignatures); err != nil {
				return err
			}
		case *ParsedStructDef:
			if err := c.GlobalScope.DefineType(string(def.id()), def.pos(), NewStructType()); err != nil {
				return err
			}
			c.Structs = append(c.Structs, &CheckedStructDef{
				Name:   def.Name,
				Fields: make([]CheckedStructField, 0, len(def.Fields)),
			})
		}
	}
	return nil
}

func isChecked(p *ParsedFile, checkedFiles map[*ParsedFile]struct{}) bool {
	if _, checked := checkedFiles[p]; checked {
		return true
	}
	checkedFiles[p] = struct{}{}
	return false
}

type checkFunc func(*ParsedFile, *CheckedFile, map[*ParsedFile]struct{}) error

func handleImport(p *ParsedImport, c *CheckedFile, checkedFiles map[*ParsedFile]struct{}, checkF checkFunc) error {
	return checkF(p.File, c.Imports[c.GlobalScope.Imports[string(p.id())]].File, checkedFiles)
}

func CheckFunctionSignatures(p *ParsedFile, c *CheckedFile) error {
	return checkFunctionSignatures(p, c, make(map[*ParsedFile]struct{}))
}

func checkFunctionSignatures(p *ParsedFile, c *CheckedFile, checkedFiles map[*ParsedFile]struct{}) error {
	if isChecked(p, checkedFiles) {
		return nil
	}
	for _, def := range p.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			if err := handleImport(def, c, checkedFiles, checkFunctionSignatures); err != nil {
				return err
			}
		case *ParsedFunDef:
			checkedParams := make([]CheckedFunParam, 0, len(def.Params))
			paramTypes := make([]TypeId, 0, len(def.Params))
			for _, param := range def.Params {
				paramType, err := checkType(param.Type, c.GlobalScope)
				if err != nil {
					return err
				}
				paramTypes = append(paramTypes, paramType)
				checkedParams = append(checkedParams, CheckedFunParam{
					Name: param.Id,
					Type: paramType,
				})
			}
			returnType := UNIT_TYPE_ID
			if def.ReturnType != nil {
				var err error
				returnType, err = checkType(def.ReturnType, c.GlobalScope)
				if err != nil {
					return err
				}
			}
			if err := c.GlobalScope.DefineFunction(string(def.id()), def.Id.Pos, &FunctionType{
				Params:  paramTypes,
				Returns: returnType,
			}); err != nil {
				return err
			}
			c.Funs = append(c.Funs, &CheckedFunDef{
				Name:       def.Id,
				Params:     checkedParams,
				ReturnType: returnType,
				Body:       &CheckedBlock{},
			})
		}
	}
	return nil
}

func CheckTypeContents(p *ParsedFile, c *CheckedFile) error {
	return checkTypeContents(p, c, make(map[*ParsedFile]struct{}))
}

func checkTypeContents(p *ParsedFile, c *CheckedFile, checkedFiles map[*ParsedFile]struct{}) error {
	if isChecked(p, checkedFiles) {
		return nil
	}
	for _, def := range p.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			if err := handleImport(def, c, checkedFiles, checkTypeContents); err != nil {
				return err
			}
		case *ParsedStructDef:
			for _, s := range c.Structs {
				if bytes.Equal(s.Name.Content, def.Name.Content) {
					if err := checkStructContents(def, s, c.GlobalScope); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func checkStructContents(def *ParsedStructDef, c *CheckedStructDef, s *Scope) error {
	fields := make(map[string]TypeId)
	for _, parsedField := range def.Fields {
		if _, exists := fields[string(parsedField.Name.Content)]; exists {
			return NewError(parsedField.Name.Pos, "field is redeclared: %s", parsedField.Name.Content)
		}
		checkedType, err := checkType(parsedField.Type, s)
		if err != nil {
			return err
		}
		fields[string(parsedField.Name.Content)] = checkedType
		c.Fields = append(c.Fields, CheckedStructField{
			Name: parsedField.Name,
			Type: checkedType,
		})
	}
	structType := s.File.Types[s.findTypeByName(string(def.Name.Content))].(*StructType)
	structType.Fields = fields
	return nil
}

func CheckBlocks(p *ParsedFile, c *CheckedFile) error {
	return checkBlocks(p, c, make(map[*ParsedFile]struct{}))
}

func checkBlocks(p *ParsedFile, c *CheckedFile, checkedFiles map[*ParsedFile]struct{}) error {
	if isChecked(p, checkedFiles) {
		return nil
	}
	for _, def := range p.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			if err := handleImport(def, c, checkedFiles, checkBlocks); err != nil {
				return err
			}
		case *ParsedFunDef:
			for _, f := range c.Funs {
				if bytes.Equal(f.Name.Content, def.Id.Content) {
					if err := checkFunBlock(def, f, c.GlobalScope); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func checkFunBlock(p *ParsedFunDef, c *CheckedFunDef, s *Scope) error {
	paramsScope := NewScope(s)
	for _, param := range p.Params {
		t, err := checkType(param.Type, s)
		if err != nil {
			return err
		}
		if err := paramsScope.DefineVar(string(param.Id.Content), param.Id.Pos, t); err != nil {
			return err
		}
	}
	returns := UNIT_TYPE_ID
	if p.ReturnType != nil {
		var err error
		returns, err = checkType(p.ReturnType, s)
		if err != nil {
			return err
		}
	}
	block, err := checkBlock(p.Body, s, &MustReturn{
		Type: returns,
	})
	if err != nil {
		return err
	}
	c.Body = block
	return nil
}

func checkBlock(p *ParsedBlock, s *Scope, controlFlow ControlFlow) (*CheckedBlock, error) {
	s = NewScope(s)
	checkedBlock := &CheckedBlock{
		Stmts: make([]CheckedStmt, 0, len(p.Stmts)),
	}
	if len(p.Stmts) == 0 {
		switch controlFlow := controlFlow.(type) {
		case *MustReturn:
			if controlFlow.Type != UNIT_TYPE_ID {
				return nil, NewError(p.pos(), "%s is returned from empty block, but %s is expected", s.typeToString(UNIT_TYPE_ID), s.typeToString(controlFlow.Type))
			}
			return checkedBlock, nil
		}
	}
	for i, stmt := range p.Stmts {
		var cf ControlFlow
		if i >= len(p.Stmts)-1 && controlFlow.typeId() != UNIT_TYPE_ID {
			cf = &MustReturn{Type: controlFlow.typeId()}
		} else {
			cf = &MayReturn{Type: controlFlow.typeId()}
		}
		checkedStmt, err := CheckStmt(stmt, s, cf)
		if err != nil {
			return nil, err
		}
		checkedBlock.Stmts = append(checkedBlock.Stmts, checkedStmt)
	}
	return checkedBlock, nil
}

func CheckStmt(stmt ParsedStmt, scope *Scope, controlFlow ControlFlow) (CheckedStmt, error) {
	switch stmt := stmt.(type) {
	case *ParsedReturn, *ParsedBlock:
		{
		}
	default:
		if _, mustReturn := controlFlow.(*MustReturn); mustReturn {
			return nil, NewError(stmt.pos(), "expected return statement")
		}
	}
	switch stmt := stmt.(type) {
	case *ParsedVar:
		return checkVar(stmt, scope, controlFlow)
	case *ParsedExprStmt:
		return checkExprStmt(stmt, scope, controlFlow)
	case *ParsedBlock:
		return checkBlock(stmt, scope, controlFlow)
	case *ParsedReturn:
		return checkReturn(stmt, scope, controlFlow)
	}
	panic("unimplemented")
}

func checkReturn(p *ParsedReturn, s *Scope, controlFlow ControlFlow) (*CheckedReturn, error) {
	if p.Arg == nil {
		if controlFlow.typeId() != UNIT_TYPE_ID {
			return nil, NewError(p.pos(), "expected return with an argument of type %s", s.typeToString(controlFlow.typeId()))
		}
		return &CheckedReturn{
			Value: nil,
		}, nil
	}
	arg, err := CheckExpr(p.Arg, s)
	if err != nil {
		return nil, err
	}
	if controlFlow.typeId() != arg.TypeId() {
		return nil, NewError(p.Arg.pos(), "expected %s, but got %s", s.typeToString(controlFlow.typeId()), s.typeToString(arg.TypeId()))
	}
	return &CheckedReturn{
		Value: arg,
	}, nil
}

func checkExprStmt(p *ParsedExprStmt, s *Scope, controlFlow ControlFlow) (*CheckedExprStmt, error) {
	expr, err := CheckExpr(p.Expr, s)
	if err != nil {
		return nil, err
	}
	return &CheckedExprStmt{
		Expr: expr,
	}, nil
}

func checkVar(p *ParsedVar, s *Scope, controlFlow ControlFlow) (*CheckedVar, error) {
	val, err := CheckExpr(p.Value, s)
	if err != nil {
		return nil, err
	}
	if err := s.DefineVar(string(p.Id.Content), p.pos(), val.TypeId()); err != nil {
		return nil, err
	}
	return &CheckedVar{
		Name:  p.Id,
		Type:  val.TypeId(),
		Value: val,
	}, nil
}

func CheckExpr(p ParsedExpr, s *Scope) (CheckedExpr, error) {
	switch p := p.(type) {
	case *ParsedUnaryExpr:
		return checkUnaryExpr(p, s)
	case *ParsedBinaryExpr:
		return checkBinaryExpr(p, s)
	case *ParsedGroupedExpr:
		return checkGroupedExpr(p, s)
	case *ParsedLiteralExpr:
		return checkLiteralExpr(p, s)
	case *ParsedCallExpr:
		return checkCallExpr(p, s)
	}
	panic("unreachable")
}

func checkCallExpr(p *ParsedCallExpr, s *Scope) (*CheckedCallExpr, error) {
	callee, err := CheckExpr(p.Callee, s)
	if err != nil {
		return nil, err
	}
	if funType, ok := s.File.Types[callee.TypeId()].(*FunctionType); ok {
		args := make([]CheckedExpr, 0, len(p.Args))
		for _, arg := range p.Args {
			checkedArg, err := CheckExpr(arg, s)
			if err != nil {
				return nil, err
			}
			args = append(args, checkedArg)
		}
		argsTypes := make([]TypeId, 0, len(p.Args))
		for _, arg := range args {
			argsTypes = append(argsTypes, arg.TypeId())
		}
		if !reflect.DeepEqual(funType.Params, argsTypes) {
			return nil, NewError(p.pos(), "expected args %s, but got %s", s.typesToStrings(funType.Params), s.typesToStrings(argsTypes))
		}
		return &CheckedCallExpr{
			Callee: callee,
			Args:   args,
			Type:   funType.Returns,
		}, nil
	}
	return nil, NewError(p.pos(), "callee is not of a function: %s", s.typeToString(callee.TypeId()))
}

func (s *Scope) typesToStrings(types []TypeId) (res []string) {
	for _, t := range types {
		res = append(res, s.typeToString(t))
	}
	return
}

func checkLiteralExpr(p *ParsedLiteralExpr, s *Scope) (*CheckedLiteralExpr, error) {
	switch p.Kind {
	case INTEGER:
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    INT_TYPE_ID,
		}, nil
	case FLOAT:
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    FLOAT_TYPE_ID,
		}, nil
	case IDENTIFIER:
		name := s.findName(string(p.Content))
		if name == NOT_FOUND {
			return nil, NewError(p.pos(), "undeclared: %s", p.Content)
		}
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    name,
		}, nil
	}
	panic("unreachable")
}

func checkGroupedExpr(p *ParsedGroupedExpr, s *Scope) (*CheckedGroupedExpr, error) {
	inner, err := CheckExpr(p.Inner, s)
	if err != nil {
		return nil, err
	}
	return &CheckedGroupedExpr{
		Left:  p.Left,
		Inner: inner,
		Right: p.Right,
	}, nil
}

func checkBinaryExpr(p *ParsedBinaryExpr, s *Scope) (*CheckedBinaryExpr, error) {
	left, err := CheckExpr(p.Left, s)
	if err != nil {
		return nil, err
	}
	right, err := CheckExpr(p.Right, s)
	if err != nil {
		return nil, err
	}
	operator, err := checkBinaryOperator(p.Op, left.TypeId(), right.TypeId(), s)
	if err != nil {
		return nil, err
	}
	return &CheckedBinaryExpr{
		Left:  left,
		Op:    operator,
		Right: right,
	}, nil
}

func checkBinaryOperator(operator Token, left, right TypeId, s *Scope) (CheckedBinaryOperator, error) {
	if left != right {
		return INVALID_BINARY_OPERATOR, NewError(operator.Pos, "operator %s is not defined for types %s and %s", operator.Kind, s.typeToString(left), s.typeToString(right))
	}
	switch operator.Kind {
	case PLUS:
		if !traitIsImplemented(ADD_TRAIT, left) {
			return INVALID_BINARY_OPERATOR, NewError(operator.Pos, "operator + is not defined for types %s and %s (try to implement %s trait)", s.typeToString(left), s.typeToString(right), ADD_TRAIT)
		}
		return CHECKED_ADD, nil
	case MINUS:
		if !traitIsImplemented(SUBTRACT_TRAIT, left) {
			return INVALID_BINARY_OPERATOR, NewError(operator.Pos, "operator - is not defined for types %s and %s (try to implement %s trait)", s.typeToString(left), s.typeToString(right), SUBTRACT_TRAIT)
		}
		return CHECKED_SUBTRACT, nil
	case STAR:
		if !traitIsImplemented(MULTIPLY_TRAIT, left) {
			return INVALID_BINARY_OPERATOR, NewError(operator.Pos, "operator * is not defined for types %s and %s (try to implement %s trait)", s.typeToString(left), s.typeToString(right), MULTIPLY_TRAIT)
		}
		return CHECKED_MULTIPLY, nil
	case SLASH:
		if !traitIsImplemented(DIVIDE_TRAIT, left) {
			return INVALID_BINARY_OPERATOR, NewError(operator.Pos, "operator / is not defined for types %s and %s (try to implement %s trait)", s.typeToString(left), s.typeToString(right), DIVIDE_TRAIT)
		}
		return CHECKED_DIVIDE, nil
	}
	panic("unreachable")
}

func checkUnaryExpr(p *ParsedUnaryExpr, s *Scope) (*CheckedUnaryExpr, error) {
	operand, err := CheckExpr(p.Operand, s)
	if err != nil {
		return nil, err
	}
	operator, err := checkUnaryOperator(p.Operator, operand.TypeId(), s)
	if err != nil {
		return nil, err
	}
	return &CheckedUnaryExpr{
		Pos:      p.pos(),
		Operator: operator,
		Operand:  operand,
	}, nil
}

func checkUnaryOperator(operator Token, operand TypeId, s *Scope) (CheckedUnaryOperator, error) {
	switch operator.Kind {
	case MINUS:
		if !traitIsImplemented(NEGATE_TRAIT, operand) {
			return INVALID_UNARY_OPERATOR, NewError(operator.Pos, "operator - is not defined for type %s (try to implement %s trait)", s.typeToString(operand), NEGATE_TRAIT)
		}
		return CHECKED_NEGATE, nil
	}
	panic("unreachable")
}

func traitIsImplemented(trait string, typeId TypeId) bool {
	switch trait {
	case NEGATE_TRAIT, ADD_TRAIT, SUBTRACT_TRAIT, MULTIPLY_TRAIT, DIVIDE_TRAIT:
		return typeId == INT_TYPE_ID || typeId == FLOAT_TYPE_ID
	}
	return false
}

const NEGATE_TRAIT = "Negate"
const ADD_TRAIT = "Add"
const SUBTRACT_TRAIT = "Subtract"
const MULTIPLY_TRAIT = "Multiply"
const DIVIDE_TRAIT = "Divide"

type ControlFlow interface {
	controlFlow()
	typeId() TypeId
}

type MustReturn struct {
	Type TypeId
}

type MayReturn struct {
	Type TypeId
}

func (m *MustReturn) controlFlow() {}
func (m *MayReturn) controlFlow()  {}

func (m *MustReturn) typeId() TypeId { return m.Type }
func (m *MayReturn) typeId() TypeId  { return m.Type }

func checkType(t ParsedType, s *Scope) (TypeId, error) {
	switch t := t.(type) {
	case *ParsedIdType:
		switch string(t.Content) {
		case "()":
			return UNIT_TYPE_ID, nil
		case "int":
			return INT_TYPE_ID, nil
		case "float":
			return FLOAT_TYPE_ID, nil
		case "char":
			return CHAR_TYPE_ID, nil
		default:
			typeId := s.findTypeByName(string(t.Content))
			if typeId == NOT_FOUND {
				return NOT_FOUND, NewError(t.pos(), "undeclared: %s", t.Content)
			}
			return typeId, nil
		}
	}
	panic("unreachable")
}

type Scope struct {
	Parent   *Scope
	Children []*Scope
	File     *CheckedFile
	Types    map[string]TypeId
	Funs     map[string]TypeId
	Vars     map[string]TypeId
	Imports  map[string]ImportId
}

func NewScope(parent *Scope) *Scope {
	s := &Scope{
		Parent:   parent,
		Children: make([]*Scope, 0),
		Types:    make(map[string]TypeId),
		Funs:     make(map[string]TypeId),
		Vars:     make(map[string]TypeId),
		Imports:  make(map[string]ImportId),
	}
	if parent != nil {
		s.File = parent.File
		parent.Children = append(parent.Children, s)
	}
	return s
}

func (s *Scope) Import(checkedImport *CheckedImport) error {
	if s.findImport(string(checkedImport.Name.Content)) != IMPORT_NOT_FOUND {
		return NewError(checkedImport.Name.Pos, "%s is already imported", checkedImport.Name.Content)
	}
	s.File.Imports = append(s.File.Imports, checkedImport)
	s.Imports[checkedImport.Name.String()] = ImportId(len(s.File.Imports) - 1)
	return nil
}

func (s *Scope) DefineType(name string, pos Pos, typ Type) error {
	if s.findTypeByName(name) != NOT_FOUND {
		return NewError(pos, "type %s is already defined", name)
	}
	s.File.Types = append(s.File.Types, typ)
	s.Types[name] = TypeId(len(s.File.Types) - 1)
	return nil
}

func (s *Scope) DefineFunction(name string, pos Pos, typ *FunctionType) error {
	if s.findName(name) != NOT_FOUND {
		return NewError(pos, "%s is already declared", name)
	}
	s.File.Types = append(s.File.Types, typ)
	s.Funs[name] = TypeId(len(s.File.Types) - 1)
	return nil
}

func (s *Scope) DefineVar(name string, pos Pos, typ TypeId) error {
	if s.findName(name) != NOT_FOUND {
		return NewError(pos, "%s is already declared", name)
	}
	s.Vars[name] = typ
	return nil
}

func (s *Scope) findName(name string) TypeId {
	if t := s.findFunction(name); t != NOT_FOUND {
		return t
	}
	if t := s.findVar(name); t != NOT_FOUND {
		return t
	}
	return NOT_FOUND
}

func (s *Scope) findFunction(name string) TypeId {
	if f, ok := s.Funs[name]; ok {
		return f
	}
	if s.Parent != nil {
		return s.Parent.findFunction(name)
	}
	return NOT_FOUND
}

func (s *Scope) findVar(name string) TypeId {
	if v, ok := s.Vars[name]; ok {
		return v
	}
	if s.Parent != nil {
		return s.Parent.findVar(name)
	}
	return NOT_FOUND
}

func (s *Scope) findImport(name string) ImportId {
	if imp, ok := s.Imports[name]; ok {
		return imp
	}
	if s.Parent != nil {
		return s.Parent.findImport(name)
	}
	return IMPORT_NOT_FOUND
}

func (s *Scope) findTypeByName(name string) TypeId {
	if t, ok := s.Types[name]; ok {
		return t
	}
	if s.Parent != nil {
		return s.Parent.findTypeByName(name)
	}
	return NOT_FOUND
}

func (s *Scope) typeToString(typeId TypeId) string {
	for name, t := range s.Types {
		if t == typeId {
			return name
		}
	}
	if s.Parent != nil {
		return s.typeToString(typeId)
	}
	panic("unreachable")
}

func (s *Scope) findAndRenameType(name string, newName string) bool {
	if t, ok := s.Types[name]; ok {
		delete(s.Types, name)
		s.Types[newName] = t
		return true
	}
	for _, child := range s.Children {
		if child.findAndRenameType(name, newName) {
			return true
		}
	}
	return false
}

func (s *Scope) findAndRenameFun(name string, newName string) bool {
	if t, ok := s.Funs[name]; ok {
		delete(s.Funs, name)
		s.Funs[newName] = t
		return true
	}
	for _, child := range s.Children {
		if child.findAndRenameFun(name, newName) {
			return true
		}
	}
	return false
}

type ImportId int

const IMPORT_NOT_FOUND ImportId = -1

type CheckedFile struct {
	Filename    string
	Imports     []*CheckedImport
	Funs        []*CheckedFunDef
	Structs     []*CheckedStructDef
	Types       []Type
	GlobalScope *Scope
}

func NewCheckedFile(filename string) *CheckedFile {
	c := &CheckedFile{
		Filename:    filename,
		Imports:     make([]*CheckedImport, 0),
		Funs:        make([]*CheckedFunDef, 0),
		Structs:     make([]*CheckedStructDef, 0),
		Types:       make([]Type, 0),
		GlobalScope: NewScope(nil),
	}
	c.GlobalScope.File = c
	c.GlobalScope.DefineType("()", Pos{}, &BuildinType{})
	c.GlobalScope.DefineType("int", Pos{}, &BuildinType{})
	c.GlobalScope.DefineType("float", Pos{}, &BuildinType{})
	c.GlobalScope.DefineType("char", Pos{}, &BuildinType{})
	constChar := c.TypeId(&PointerType{
		Type: CHAR_TYPE_ID,
	})
	c.GlobalScope.DefineFunction("inlineC", Pos{}, &FunctionType{
		Params:  []TypeId{constChar},
		Returns: UNIT_TYPE_ID,
	})
	return c
}

func (c *CheckedFile) TypeId(t Type) TypeId {
	for i, typ := range c.Types {
		if t == typ {
			return TypeId(i)
		}
	}
	return NOT_FOUND
}

type CheckedImport struct {
	Name Token
	File *CheckedFile
}

type CheckedFunDef struct {
	Name       Token
	Params     []CheckedFunParam
	ReturnType TypeId
	Body       *CheckedBlock
}

type CheckedFunParam struct {
	Name Token
	Type TypeId
}

type CheckedStructDef struct {
	Name   Token
	Fields []CheckedStructField
}

type CheckedStructField struct {
	Name Token
	Type TypeId
}

type CheckedStmt interface {
	checkedStmt()
}

type CheckedVar struct {
	Name  Token
	Type  TypeId
	Value CheckedExpr
}

type CheckedExprStmt struct {
	Expr CheckedExpr
}

type CheckedBlock struct {
	Stmts []CheckedStmt
}

type CheckedReturn struct {
	Value CheckedExpr
}

func (c *CheckedVar) checkedStmt()      {}
func (c *CheckedExprStmt) checkedStmt() {}
func (c *CheckedBlock) checkedStmt()    {}
func (c *CheckedReturn) checkedStmt()   {}

type CheckedExpr interface {
	checkedExpr()
	TypeId() TypeId
}

type CheckedUnaryExpr struct {
	Pos
	Operator CheckedUnaryOperator
	Operand  CheckedExpr
}

type CheckedUnaryOperator int

const INVALID_UNARY_OPERATOR CheckedUnaryOperator = -1

const (
	CHECKED_UNARY_PLUS CheckedUnaryOperator = iota
	CHECKED_NEGATE
)

type CheckedBinaryExpr struct {
	Left  CheckedExpr
	Op    CheckedBinaryOperator
	Right CheckedExpr
}

type CheckedBinaryOperator int

const INVALID_BINARY_OPERATOR CheckedBinaryOperator = -1

const (
	CHECKED_ADD CheckedBinaryOperator = iota
	CHECKED_SUBTRACT
	CHECKED_MULTIPLY
	CHECKED_DIVIDE
)

type CheckedGroupedExpr struct {
	Left  Token
	Inner CheckedExpr
	Right Token
}

type CheckedLiteralExpr struct {
	Literal Token
	Type    TypeId
}

type CheckedCallExpr struct {
	Callee CheckedExpr
	Args   []CheckedExpr
	Type   TypeId
}

func (c *CheckedUnaryExpr) checkedExpr()   {}
func (c *CheckedBinaryExpr) checkedExpr()  {}
func (c *CheckedGroupedExpr) checkedExpr() {}
func (c *CheckedLiteralExpr) checkedExpr() {}
func (c *CheckedCallExpr) checkedExpr()    {}

func (c *CheckedUnaryExpr) TypeId() TypeId {
	return c.Operand.TypeId()
}
func (c *CheckedBinaryExpr) TypeId() TypeId {
	return c.Left.TypeId()
}
func (c *CheckedGroupedExpr) TypeId() TypeId {
	return c.Inner.TypeId()
}
func (c *CheckedLiteralExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedCallExpr) TypeId() TypeId {
	return c.Type
}

type TypeId int

const NOT_FOUND TypeId = -1
const NEVER_TYPE_ID TypeId = -2

const (
	UNIT_TYPE_ID TypeId = iota
	INT_TYPE_ID
	FLOAT_TYPE_ID
	CHAR_TYPE_ID
)

type Type interface {
	typ()
}

type BuildinType struct{}

type IdType struct {
	Id string
}

type PointerType struct {
	Type TypeId
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
	Params  []TypeId
	Returns TypeId
}

func (b *BuildinType) typ()  {}
func (i *IdType) typ()       {}
func (p *PointerType) typ()  {}
func (s *StructType) typ()   {}
func (f *FunctionType) typ() {}
