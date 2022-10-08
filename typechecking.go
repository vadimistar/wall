package wall

import (
	"fmt"
	"reflect"
	"strings"
)

func CheckCompilationUnit(f *ParsedFile) (*CheckedFile, error) {
	checkedUnit := NewCheckedCompilationUnit(f.pos().Filename)
	if err := CheckImports(f, checkedUnit); err != nil {
		return nil, err
	}
	if err := CheckTypeSignatures(f, checkedUnit); err != nil {
		return nil, err
	}
	if err := CheckFunctionSignatures(f, checkedUnit); err != nil {
		return nil, err
	}
	if err := CheckTypeContents(f, checkedUnit); err != nil {
		return nil, err
	}
	if err := CheckBlocks(f, checkedUnit); err != nil {
		return nil, err
	}
	return checkedUnit, nil
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
				checkedFile = NewCheckedFile(def.File.pos().Filename, c.Types)
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
			chechedStructDef := &CheckedStructDef{
				Name:   def.Name,
				Fields: make([]CheckedStructField, 0, len(def.Fields)),
			}
			if err := c.GlobalScope.DefineType(&chechedStructDef.Name, NewStructType()); err != nil {
				return err
			}
			c.Structs = append(c.Structs, chechedStructDef)
		case *ParsedTypealiasDef:
			checked := &CheckedTypealiasDef{
				Name: &def.Name,
				Type: NEVER_TYPE_ID,
			}
			if err := c.GlobalScope.DefineTypealias(checked.Name); err != nil {
				return err
			}
			c.Typealiases = append(c.Typealiases, checked)
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
	return checkFunctionSignatures(p, c, make(map[*ParsedFile]struct{}), c)
}

func checkFunctionSignatures(p *ParsedFile, c *CheckedFile, checkedFiles map[*ParsedFile]struct{}, mainFile *CheckedFile) error {
	if isChecked(p, checkedFiles) {
		return nil
	}
	for _, def := range p.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			if err := checkFunctionSignatures(def.File, c.Imports[c.GlobalScope.Imports[string(def.id())]].File, checkedFiles, mainFile); err != nil {
				return err
			}
		case *ParsedFunDef:
			checkedParams := make([]CheckedFunParam, 0, len(def.Params))
			paramTypes := make([]TypeId, 0, len(def.Params))
			for i, param := range def.Params {
				paramType, err := checkType(param.Type, c.GlobalScope)
				if err != nil {
					return err
				}
				paramTypes = append(paramTypes, paramType)
				checkedParams = append(checkedParams, CheckedFunParam{
					Name: &def.Params[i].Id,
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
			if def.Typename != nil {
				typename := c.GlobalScope.findType(def.Typename.Content)
				if typename == nil {
					return NewError(def.Typename.Pos, "type is not declared: %s", typename.Token.Content)
				}
				checkedMethod := &CheckedMethodDef{
					Typename:   typename.Token,
					Name:       &def.Id,
					Params:     checkedParams,
					ReturnType: returnType,
					Body:       &CheckedBlock{},
				}
				if err := c.GlobalScope.DefineMethod(checkedMethod.Typename, checkedMethod.Name, &FunctionType{
					Params:  paramTypes,
					Returns: returnType,
				}); err != nil {
					return err
				}
				c.Methods = append(c.Methods, checkedMethod)
				break
			}
			if def.Id.Content == "main" && c == mainFile {
				if err := validateMain(def.pos(), paramTypes, returnType, c); err != nil {
					return err
				}
			}
			checkedFunDef := &CheckedFunDef{
				Name:       &def.Id,
				Params:     checkedParams,
				ReturnType: returnType,
				Body:       &CheckedBlock{},
			}
			if err := c.GlobalScope.DefineFunction(&def.Id, &FunctionType{
				Params:  paramTypes,
				Returns: returnType,
			}); err != nil {
				return err
			}
			c.Funs = append(c.Funs, checkedFunDef)
		case *ParsedExternFunDef:
			checkedParams := make([]CheckedFunParam, 0, len(def.Params))
			paramTypes := make([]TypeId, 0, len(def.Params))
			for i, param := range def.Params {
				paramType, err := checkType(param.Type, c.GlobalScope)
				if err != nil {
					return err
				}
				paramTypes = append(paramTypes, paramType)
				checkedParams = append(checkedParams, CheckedFunParam{
					Name: &def.Params[i].Id,
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
			checkedFunDef := &CheckedExternFunDef{
				Name:       &def.Name,
				Params:     checkedParams,
				ReturnType: returnType,
			}
			if err := c.GlobalScope.DefineFunction(&def.Name, &FunctionType{
				Params:  paramTypes,
				Returns: returnType,
			}); err != nil {
				return err
			}
			c.ExternFuns = append(c.ExternFuns, checkedFunDef)
		}
	}
	return nil
}

func validateMain(pos Pos, paramTypes []TypeId, returnType TypeId, c *CheckedFile) error {
	constChar := c.TypeId(&PointerType{
		Type: CHAR_TYPE_ID,
	})
	if !reflect.DeepEqual(paramTypes, []TypeId{}) && !reflect.DeepEqual(paramTypes, []TypeId{INT32_TYPE_ID, constChar}) {
		return NewError(pos, "invalid params in main function: %s (expected [] or [int32, **char])", c.GlobalScope.typesToStrings(paramTypes))
	}
	if !reflect.DeepEqual(returnType, INT32_TYPE_ID) {
		return NewError(pos, "invalid return type in main function: %s (expected int32)", c.GlobalScope.TypeToString(returnType))
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
				if s.Name.Content == def.Name.Content {
					if err := checkStructContents(def, s, c.GlobalScope); err != nil {
						return err
					}
				}
			}
		case *ParsedTypealiasDef:
			for _, t := range c.Typealiases {
				if t.Name.Content == def.Name.Content {
					checkedT, err := checkType(def.Type, c.GlobalScope)
					if err != nil {
						return err
					}
					t.Type = checkedT
					c.GlobalScope.Types[string(t.Name.Content)].TypeId = checkedT
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
	structType := (*s.File.Types)[s.findType(string(def.Name.Content)).TypeId].(*StructType)
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
				if f.Name.Content == def.Id.Content {
					if err := checkFunBlock(def, f, c.GlobalScope); err != nil {
						return err
					}
				}
			}
			for _, m := range c.Methods {
				if m.Name.Content == def.Id.Content {
					if err := checkMethodBlock(def, m, c.GlobalScope, c.GlobalScope.findType(m.Typename.Content).TypeId); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func checkMethodBlock(p *ParsedFunDef, c *CheckedMethodDef, s *Scope, thisType TypeId) error {
	s = NewScope(s)
	s.MethodType = thisType
	for _, param := range c.Params {
		if err := s.DefineVar(param.Name, param.Type, false); err != nil {
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

func checkFunBlock(p *ParsedFunDef, c *CheckedFunDef, s *Scope) error {
	s = NewScope(s)
	for _, param := range c.Params {
		if err := s.DefineVar(param.Name, param.Type, false); err != nil {
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
				return nil, NewError(p.pos(), "%s is returned from empty block, but %s is expected", s.TypeToString(UNIT_TYPE_ID), s.TypeToString(controlFlow.Type))
			}
			return checkedBlock, nil
		}
	}
	_, mustReturn := controlFlow.(*MustReturn)
	_, loop := controlFlow.(*MayReturnFromLoop)
	for i, stmt := range p.Stmts {
		var cf ControlFlow
		if mustReturn && i+1 >= len(p.Stmts) && controlFlow.typeId() != UNIT_TYPE_ID {
			cf = &MustReturn{Type: controlFlow.typeId()}
		} else if loop {
			cf = &MayReturnFromLoop{Type: controlFlow.typeId()}
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
	case *ParsedReturn, *ParsedBlock, *ParsedIf:
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
	case *ParsedIf:
		return checkIf(stmt, scope, controlFlow)
	case *ParsedWhile:
		return checkWhile(stmt, scope, controlFlow)
	case *ParsedBreak:
		return checkBreak(stmt, scope, controlFlow)
	case *ParsedContinue:
		return checkContinue(stmt, scope, controlFlow)
	}
	panic("unimplemented")
}

func checkContinue(p *ParsedContinue, s *Scope, controlFlow ControlFlow) (*CheckedContinue, error) {
	if _, mayReturnFromLoop := controlFlow.(*MayReturnFromLoop); mayReturnFromLoop {
		return &CheckedContinue{
			Continue: p.Continue,
		}, nil
	}
	return nil, NewError(p.pos(), "can't use continue not in a loop")
}

func checkBreak(p *ParsedBreak, s *Scope, controlFlow ControlFlow) (*CheckedBreak, error) {
	if _, mayReturnFromLoop := controlFlow.(*MayReturnFromLoop); mayReturnFromLoop {
		return &CheckedBreak{
			Break: p.Break,
		}, nil
	}
	return nil, NewError(p.pos(), "can't use break not in a loop")
}

func checkWhile(p *ParsedWhile, s *Scope, controlFlow ControlFlow) (*CheckedWhile, error) {
	if _, mustReturn := controlFlow.(*MustReturn); mustReturn {
		return nil, NewError(p.pos(), "expected return statement")
	}
	cond, err := CheckExpr(p.Condition, s)
	if err != nil {
		return nil, err
	}
	if cond.TypeId() != BOOL_TYPE_ID {
		return nil, NewError(p.Condition.pos(), "a condition must be a boolean expression, but it's %s", s.TypeToString(cond.TypeId()))
	}
	body, err := checkBlock(p.Body, s, &MayReturnFromLoop{
		Type: controlFlow.typeId(),
	})
	if err != nil {
		return nil, err
	}
	return &CheckedWhile{
		Cond: cond,
		Body: body,
	}, nil
}

func checkIf(p *ParsedIf, s *Scope, controlFlow ControlFlow) (*CheckedIf, error) {
	if _, mustReturn := controlFlow.(*MustReturn); mustReturn {
		if p.ElseBody == nil {
			return nil, NewError(p.pos(), "if statement without else block may not return (add else block with return statement)")
		}
	}
	cond, err := CheckExpr(p.Condition, s)
	if err != nil {
		return nil, err
	}
	body, err := checkBlock(p.Body, s, controlFlow)
	if err != nil {
		return nil, err
	}
	var elseBody *CheckedBlock
	if p.ElseBody != nil {
		elseBody, err = checkBlock(p.ElseBody, s, controlFlow)
		if err != nil {
			return nil, err
		}
	}
	return &CheckedIf{
		Cond:     cond,
		Body:     body,
		ElseBody: elseBody,
	}, nil
}

func checkReturn(p *ParsedReturn, s *Scope, controlFlow ControlFlow) (*CheckedReturn, error) {
	if p.Arg == nil {
		if controlFlow.typeId() != UNIT_TYPE_ID {
			return nil, NewError(p.pos(), "expected return with an argument of type %s", s.TypeToString(controlFlow.typeId()))
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
		return nil, NewError(p.Arg.pos(), "expected %s, but got %s", s.TypeToString(controlFlow.typeId()), s.TypeToString(arg.TypeId()))
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
	if val.TypeId() == UNIT_TYPE_ID {
		return nil, NewError(p.pos(), "can't declare a variable of type %s", s.TypeToString(UNIT_TYPE_ID))
	}
	checked := &CheckedVar{
		Mut:   p.Mut,
		Name:  &p.Id,
		Value: val,
	}
	if err := s.DefineVar(checked.Name, val.TypeId(), p.Mut != nil); err != nil {
		return nil, err
	}
	return checked, nil
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
	case *ParsedIdExpr:
		return checkIdExpr(p, s)
	case *ParsedCallExpr:
		return checkCallExpr(p, s)
	case *ParsedStructInitExpr:
		return checkStructInitExpr(p, s)
	case *ParsedObjectAccessExpr:
		return checkObjectAccessExpr(p, s)
	case *ParsedModuleAccessExpr:
		return checkModuleAccessExpr(p, s)
	case *ParsedAsExpr:
		return checkAsExpr(p, s)
	}
	panic("unreachable")
}

func checkAsExpr(p *ParsedAsExpr, s *Scope) (*CheckedAsExpr, error) {
	val, err := CheckExpr(p.Value, s)
	if err != nil {
		return nil, err
	}
	if !isScalar(val.TypeId(), s) {
		return nil, NewError(p.pos(), "expected a scalar type value, but got %s", s.TypeToString(val.TypeId()))
	}
	typ, err := checkType(p.Type, s)
	if err != nil {
		return nil, err
	}
	if !isScalar(typ, s) {
		return nil, NewError(p.pos(), "expected a scalar type, but got %s", s.TypeToString(typ))
	}
	return &CheckedAsExpr{
		Value: val,
		Type:  typ,
	}, nil
}

func checkModuleAccessExpr(p *ParsedModuleAccessExpr, s *Scope) (*CheckedModuleAccessExpr, error) {
	importId := s.findImport(string(p.Module.Content))
	if importId == IMPORT_NOT_FOUND {
		return nil, NewError(p.Module.Pos, "unresolved import: %s", p.Module.Content)
	}
	importScope := s.File.Imports[importId].File.GlobalScope
	member, err := CheckExpr(p.Member, importScope)
	if err != nil {
		return nil, err
	}
	return &CheckedModuleAccessExpr{
		Module: p.Module,
		Member: member,
		Type:   member.TypeId(),
	}, nil
}

func checkObjectAccessExpr(p *ParsedObjectAccessExpr, s *Scope) (CheckedExpr, error) {
	var typ TypeId
	var object CheckedExpr
	if p.Object == nil {
		typ = s.MethodType
	} else {
		var err error
		object, err = CheckExpr(p.Object, s)
		if err != nil {
			return nil, err
		}
		typ = object.TypeId()
	}
	structType, isStructType := (*s.File.Types)[typ].(*StructType)
	if !isStructType {
		return nil, NewError(p.pos(), "can't use . operator: expected struct type, but got %s", s.TypeToString(typ))
	}
	method := s.findMethod(s.TypeToString(typ), p.Member.Content, make(map[string]struct{}))
	if method != nil {
		return &CheckedMethodExpr{
			Object: object,
			Method: method.Id,
			Type:   method.TypeId,
		}, nil
	}
	fieldType, fieldExists := structType.Fields[string(p.Member.Content)]
	if !fieldExists {
		return nil, NewError(p.Member.Pos, "unknown field: %s", p.Member.Content)
	}
	return &CheckedMemberAccessExpr{
		Object: object,
		Member: p.Member,
		Type:   fieldType,
	}, nil
}

func checkStructInitExpr(p *ParsedStructInitExpr, s *Scope) (*CheckedStructInitExpr, error) {
	checkedFields := make([]CheckedStructInitField, 0, len(p.Fields))
	notInitialized := make(map[string]struct{}, len(p.Fields))
	structTypeId, err := checkType(p.Name, s)
	if err != nil {
		return nil, err
	}
	structType, ok := (*s.File.Types)[structTypeId].(*StructType)
	if !ok {
		return nil, NewError(p.pos(), "an invalid type in the struct initializer")
	}
	for name := range structType.Fields {
		notInitialized[name] = struct{}{}
	}
	for _, field := range p.Fields {
		if t, ok := structType.Fields[string(field.Name.Content)]; ok {
			val, err := CheckExpr(field.Value, s)
			if err != nil {
				return nil, err
			}
			if t != val.TypeId() {
				return nil, NewError(field.Name.Pos, "expected %s, but got %s", s.TypeToString(t), s.TypeToString(val.TypeId()))
			}
			delete(notInitialized, string(field.Name.Content))
			checkedFields = append(checkedFields, CheckedStructInitField{
				Name:  field.Name,
				Value: val,
			})
		} else {
			return nil, NewError(field.Name.Pos, "unknown field: %s", field.Name.Content)
		}
	}
	if len(notInitialized) > 0 {
		keys := make([]string, 0, len(notInitialized))
		for k := range notInitialized {
			keys = append(keys, k)
		}
		return nil, NewError(p.pos(), "uninitialized fields: %s", keys)
	}
	return &CheckedStructInitExpr{
		Fields: checkedFields,
		Pos:    p.pos(),
		Type:   structTypeId,
	}, nil
}

func checkIdExpr(p *ParsedIdExpr, s *Scope) (*CheckedIdExpr, error) {
	name := s.findName(string(p.Content))
	if name == nil {
		return nil, NewError(p.pos(), "undeclared: %s", p.Content)
	}
	return &CheckedIdExpr{
		Id:   name.Token,
		Type: name.TypeId,
		Pos:  p.Pos,
	}, nil
}

func checkCallExpr(p *ParsedCallExpr, s *Scope) (*CheckedCallExpr, error) {
	callee, err := CheckExpr(p.Callee, s)
	if err != nil {
		return nil, err
	}
	if funType, ok := (*s.File.Types)[callee.TypeId()].(*FunctionType); ok {
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
	return nil, NewError(p.pos(), "callee is not of a function: %s", s.TypeToString(callee.TypeId()))
}

func (s *Scope) typesToStrings(types []TypeId) (res []string) {
	for _, t := range types {
		res = append(res, s.TypeToString(t))
	}
	return
}

func checkLiteralExpr(p *ParsedLiteralExpr, s *Scope) (*CheckedLiteralExpr, error) {
	switch p.Kind {
	case INTEGER:
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    INT32_TYPE_ID,
		}, nil
	case FLOAT:
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    FLOAT64_TYPE_ID,
		}, nil
	case STRING:
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    s.File.TypeId(&PointerType{Type: CHAR_TYPE_ID}),
		}, nil
	case TRUE, FALSE:
		return &CheckedLiteralExpr{
			Literal: p.Token,
			Type:    BOOL_TYPE_ID,
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
	operator, returnType, err := checkBinaryOperator(p.Op, left, right.TypeId(), s)
	if err != nil {
		return nil, err
	}
	return &CheckedBinaryExpr{
		Left:  left,
		Op:    operator,
		Right: right,
		Type:  returnType,
	}, nil
}

func checkBinaryOperator(operator Token, left CheckedExpr, right TypeId, s *Scope) (CheckedBinaryOperator, TypeId, error) {
	if left.TypeId() != right {
		return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator %s is not defined for types %s and %s", operator.Kind, s.TypeToString(left.TypeId()), s.TypeToString(right))
	}
	switch operator.Kind {
	case EQ:
		if isTemporaryValue(left) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "can't assign to a temporary value: %s", s.TypeToString(left.TypeId()))
		}
		if isMutable(left, s) {
			return CHECKED_ASSIGN, left.TypeId(), nil
		}
		return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "left side of an expression is not mutable")
	case PLUS:
		if !traitIsImplemented(ADD_TRAIT, left.TypeId(), s) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator + is not defined for types %s and %s (try to implement %s trait)", s.TypeToString(left.TypeId()), s.TypeToString(right), ADD_TRAIT)
		}
		return CHECKED_ADD, left.TypeId(), nil
	case MINUS:
		if !traitIsImplemented(SUBTRACT_TRAIT, left.TypeId(), s) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator - is not defined for types %s and %s (try to implement %s trait)", s.TypeToString(left.TypeId()), s.TypeToString(right), SUBTRACT_TRAIT)
		}
		return CHECKED_SUBTRACT, left.TypeId(), nil
	case STAR:
		if !traitIsImplemented(MULTIPLY_TRAIT, left.TypeId(), s) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator * is not defined for types %s and %s (try to implement %s trait)", s.TypeToString(left.TypeId()), s.TypeToString(right), MULTIPLY_TRAIT)
		}
		return CHECKED_MULTIPLY, left.TypeId(), nil
	case SLASH:
		if !traitIsImplemented(DIVIDE_TRAIT, left.TypeId(), s) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator / is not defined for types %s and %s (try to implement %s trait)", s.TypeToString(left.TypeId()), s.TypeToString(right), DIVIDE_TRAIT)
		}
		return CHECKED_DIVIDE, left.TypeId(), nil
	case EQEQ, BANGEQ:
		if !traitIsImplemented(EQUALS_TRAIT, left.TypeId(), s) && !traitIsImplemented(ORDERING_TRAIT, left.TypeId(), s) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator %s is not defined for types %s and %s (try to implement %s trait)", operator.Kind, s.TypeToString(left.TypeId()), s.TypeToString(right), EQUALS_TRAIT)
		}
		if operator.Kind == EQEQ {
			return CHECKED_EQUALS, BOOL_TYPE_ID, nil
		}
		if operator.Kind == BANGEQ {
			return CHECKED_NOTEQUALS, BOOL_TYPE_ID, nil
		}
	case LT, LTEQ, GT, GTEQ:
		if !traitIsImplemented(ORDERING_TRAIT, left.TypeId(), s) {
			return INVALID_BINARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator %s is not defined for types %s and %s (try to implement %s trait)", operator.Kind, s.TypeToString(left.TypeId()), s.TypeToString(right), ORDERING_TRAIT)
		}
		if operator.Kind == LT {
			return CHECKED_LESSTHAN, BOOL_TYPE_ID, nil
		}
		if operator.Kind == LTEQ {
			return CHECKED_LESSOREQUAL, BOOL_TYPE_ID, nil
		}
		if operator.Kind == GT {
			return CHECKED_GREATERTHAN, BOOL_TYPE_ID, nil
		}
		if operator.Kind == GTEQ {
			return CHECKED_GREATEROREQUAL, BOOL_TYPE_ID, nil
		}
	}
	panic("unreachable")
}

func isMutable(left CheckedExpr, s *Scope) bool {
	switch left := left.(type) {
	case *CheckedIdExpr:
		name := s.findVar(left.Id.Content)
		if name != nil {
			return name.Mutable
		}
	case *CheckedGroupedExpr:
		return isMutable(left.Inner, s)
	case *CheckedMemberAccessExpr:
		return isMutable(left.Object, s)
	case *CheckedModuleAccessExpr:
		return isMutable(left.Member, s.File.Imports[s.findImport(left.Module.Content)].File.GlobalScope)
	}
	return false
}

func checkUnaryExpr(p *ParsedUnaryExpr, s *Scope) (*CheckedUnaryExpr, error) {
	operand, err := CheckExpr(p.Operand, s)
	if err != nil {
		return nil, err
	}
	operator, resultType, err := checkUnaryOperator(p.Operator, operand, s)
	if err != nil {
		return nil, err
	}
	return &CheckedUnaryExpr{
		Pos:      p.pos(),
		Operator: operator,
		Operand:  operand,
		Type:     resultType,
	}, nil
}

func checkUnaryOperator(operator Token, operand CheckedExpr, s *Scope) (CheckedUnaryOperator, TypeId, error) {
	switch operator.Kind {
	case MINUS:
		if !traitIsImplemented(NEGATE_TRAIT, operand.TypeId(), s) {
			return INVALID_UNARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "operator - is not defined for type %s (try to implement %s trait)", s.TypeToString(operand.TypeId()), NEGATE_TRAIT)
		}
		return CHECKED_NEGATE, operand.TypeId(), nil
	case AMP:
		if isTemporaryValue(operand) {
			return INVALID_UNARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "can't take an address of a temporary value: %s", s.TypeToString(operand.TypeId()))
		}
		return CHECKED_ADDRESS, s.File.TypeId(&PointerType{
			Type: operand.TypeId(),
		}), nil
	case STAR:
		if pointerType, isPointer := (*s.File.Types)[operand.TypeId()].(*PointerType); isPointer {
			return CHECKED_DEREF, pointerType.Type, nil
		}
		return INVALID_UNARY_OPERATOR, NOT_FOUND, NewError(operator.Pos, "can't use * operator on %s (a pointer type is expected)", s.TypeToString(operand.TypeId()))
	}
	panic("unreachable")
}

func isTemporaryValue(operand CheckedExpr) bool {
	switch operand := operand.(type) {
	case *CheckedUnaryExpr, *CheckedBinaryExpr, *CheckedLiteralExpr, *CheckedCallExpr, *CheckedStructInitExpr:
		return true
	case *CheckedGroupedExpr:
		return isTemporaryValue(operand.Inner)
	case *CheckedIdExpr, *CheckedMemberAccessExpr:
		return false
	}
	return false
}

func traitIsImplemented(trait string, typeId TypeId, s *Scope) bool {
	switch trait {
	case NEGATE_TRAIT, ADD_TRAIT, SUBTRACT_TRAIT, MULTIPLY_TRAIT, DIVIDE_TRAIT, ORDERING_TRAIT:
		return isArithmetic(typeId)
	case EQUALS_TRAIT:
		return isScalar(typeId, s)
	}
	return false
}

func isArithmetic(typeId TypeId) bool {
	return typeId == INT_TYPE_ID || typeId == INT8_TYPE_ID ||
		typeId == INT16_TYPE_ID || typeId == INT32_TYPE_ID || typeId ==
		INT64_TYPE_ID || typeId == UINT_TYPE_ID || typeId == UINT8_TYPE_ID ||
		typeId == UINT16_TYPE_ID || typeId == UINT32_TYPE_ID || typeId ==
		UINT64_TYPE_ID || typeId == FLOAT32_TYPE_ID || typeId == FLOAT64_TYPE_ID
}

func isScalar(typeId TypeId, s *Scope) bool {
	if _, isPointee := (*s.File.Types)[typeId].(*PointerType); isPointee {
		return true
	}
	return isArithmetic(typeId) || typeId == CHAR_TYPE_ID || typeId == BOOL_TYPE_ID
}

const NEGATE_TRAIT = "Negate"
const ADD_TRAIT = "Add"
const SUBTRACT_TRAIT = "Subtract"
const MULTIPLY_TRAIT = "Multiply"
const DIVIDE_TRAIT = "Divide"
const EQUALS_TRAIT = "Equals"
const ORDERING_TRAIT = "Ordering"

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

type MayReturnFromLoop struct {
	Type TypeId
}

func (m *MustReturn) controlFlow()        {}
func (m *MayReturn) controlFlow()         {}
func (m *MayReturnFromLoop) controlFlow() {}

func (m *MustReturn) typeId() TypeId        { return m.Type }
func (m *MayReturn) typeId() TypeId         { return m.Type }
func (m *MayReturnFromLoop) typeId() TypeId { return m.Type }

func checkType(t ParsedType, s *Scope) (TypeId, error) {
	switch t := t.(type) {
	case *ParsedIdType:
		switch string(t.Content) {
		case "()":
			return UNIT_TYPE_ID, nil
		case "int":
			return INT_TYPE_ID, nil
		case "int8":
			return INT8_TYPE_ID, nil
		case "int16":
			return INT16_TYPE_ID, nil
		case "int32":
			return INT32_TYPE_ID, nil
		case "int64":
			return INT64_TYPE_ID, nil
		case "uint":
			return UINT_TYPE_ID, nil
		case "uint8":
			return UINT8_TYPE_ID, nil
		case "uint16":
			return UINT16_TYPE_ID, nil
		case "uint32":
			return UINT32_TYPE_ID, nil
		case "uint64":
			return UINT64_TYPE_ID, nil
		case "float32":
			return FLOAT32_TYPE_ID, nil
		case "float64":
			return FLOAT64_TYPE_ID, nil
		case "char":
			return CHAR_TYPE_ID, nil
		case "bool":
			return BOOL_TYPE_ID, nil
		default:
			typ := s.findType(string(t.Content))
			if typ == nil {
				return NOT_FOUND, NewError(t.pos(), "undeclared: %s", t.Content)
			}
			return typ.TypeId, nil
		}
	case *ParsedPointerType:
		to, err := checkType(t.To, s)
		if err != nil {
			return NOT_FOUND, err
		}
		return s.File.TypeId(&PointerType{
			Type: to,
		}), nil
	case *ParsedModuleAccessType:
		importId := s.findImport(string(t.Module.Content))
		if importId == IMPORT_NOT_FOUND {
			return NOT_FOUND, NewError(t.Module.Pos, "unresolved import: %s", t.Module.Content)
		}
		importScope := s.File.Imports[importId].File.GlobalScope
		member, err := checkType(t.Member, importScope)
		if err != nil {
			return NOT_FOUND, err
		}
		return member, nil
	}
	panic("unreachable")
}

type Name struct {
	Token *Token
	TypeId
	Mutable bool
}

type TypeName struct {
	Token *Token
	TypeId
}

type MethodName struct {
	Typename *Token
	Id       *Token
	TypeId
}

type Scope struct {
	Parent     *Scope
	Children   []*Scope
	File       *CheckedFile
	Types      map[string]*TypeName
	Funs       map[string]*Name
	Methods    map[string]*MethodName
	Vars       map[string]*Name
	Imports    map[string]ImportId
	MethodType TypeId
}

func NewScope(parent *Scope) *Scope {
	s := &Scope{
		Parent:   parent,
		Children: make([]*Scope, 0),
		Types:    make(map[string]*TypeName),
		Funs:     make(map[string]*Name),
		Methods:  make(map[string]*MethodName),
		Vars:     make(map[string]*Name),
		Imports:  make(map[string]ImportId),
	}
	if parent != nil {
		s.File = parent.File
		s.MethodType = parent.MethodType
		parent.Children = append(parent.Children, s)
	}
	return s
}

func (s *Scope) Import(checkedImport *CheckedImport) error {
	if s.findImport(string(checkedImport.Name.Content)) != IMPORT_NOT_FOUND {
		return NewError(checkedImport.Name.Pos, "%s is already imported", checkedImport.Name.Content)
	}
	s.File.Imports = append(s.File.Imports, checkedImport)
	s.Imports[string(checkedImport.Name.Content)] = ImportId(len(s.File.Imports) - 1)
	return nil
}

func (s *Scope) DefineType(token *Token, typ Type) error {
	if s.findType(string(token.Content)) != nil {
		return NewError(token.Pos, "type %s is already defined", token.Content)
	}
	*s.File.Types = append((*s.File.Types), typ)
	s.Types[string(token.Content)] = &TypeName{
		Token:  token,
		TypeId: TypeId(len((*s.File.Types)) - 1),
	}
	return nil
}

func (s *Scope) DefineTypealias(token *Token) error {
	if s.findType(string(token.Content)) != nil {
		return NewError(token.Pos, "type %s is already defined", token.Content)
	}
	s.Types[string(token.Content)] = &TypeName{
		Token:  token,
		TypeId: 0,
	}
	return nil
}

func (s *Scope) DefineFunction(token *Token, typ *FunctionType) error {
	if s.findName(string(token.Content)) != nil {
		return NewError(token.Pos, "%s is already declared", token.Content)
	}
	(*s.File.Types) = append((*s.File.Types), typ)
	s.Funs[string(token.Content)] = &Name{
		Token:  token,
		TypeId: TypeId(len((*s.File.Types)) - 1),
	}
	return nil
}

func (s *Scope) DefineMethod(typename *Token, id *Token, typ *FunctionType) error {
	if s.findMethod(typename.Content, id.Content, make(map[string]struct{})) != nil {
		return NewError(id.Pos, "method %s for type %s is already declared", id.Content, typename.Content)
	}
	s.Methods[typename.Content+"."+id.Content] = &MethodName{
		Typename: typename,
		Id:       id,
		TypeId:   s.File.TypeId(typ),
	}
	return nil
}

func (s *Scope) DefineVar(token *Token, typ TypeId, mutable bool) error {
	if s.findName(string(token.Content)) != nil {
		return NewError(token.Pos, "%s is already declared", token.Content)
	}
	s.Vars[string(token.Content)] = &Name{
		Token:   token,
		TypeId:  typ,
		Mutable: mutable,
	}
	return nil
}

func (s *Scope) findMethod(typename string, name string, checkedFiles map[string]struct{}) *MethodName {
	if _, checked := checkedFiles[s.File.Filename]; checked {
		return nil
	}
	checkedFiles[s.File.Filename] = struct{}{}
	for s.Parent != nil {
		s = s.Parent
	}
	if f, ok := s.Methods[typename+"."+name]; ok {
		return f
	}
	for _, imp := range s.File.Imports {
		if m := imp.File.GlobalScope.findMethod(typename, name, checkedFiles); m != nil {
			return m
		}
	}
	return nil
}

func (s *Scope) findName(name string) *Name {
	if t := s.findFunction(name); t != nil {
		return t
	}
	if t := s.findVar(name); t != nil {
		return t
	}
	return nil
}

func (s *Scope) findFunction(name string) *Name {
	if f, ok := s.Funs[name]; ok {
		return f
	}
	if s.Parent != nil {
		return s.Parent.findFunction(name)
	}
	return nil
}

func (s *Scope) findVar(name string) *Name {
	if v, ok := s.Vars[name]; ok {
		return v
	}
	if s.Parent != nil {
		return s.Parent.findVar(name)
	}
	return nil
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

func (s *Scope) findType(name string) *TypeName {
	if t, ok := s.Types[name]; ok {
		return t
	}
	if s.Parent != nil {
		return s.Parent.findType(name)
	}
	return nil
}

func (s *Scope) TypeToString(typeId TypeId) string {
	switch t := (*s.File.Types)[typeId].(type) {
	case *BuildinType, *IdType, *StructType:
		res := findIdTypeInFile(typeId, s.File, make(map[*CheckedFile]struct{}))
		return res
	case *PointerType:
		return "*" + s.TypeToString(t.Type)
	case *FunctionType:
		return fmt.Sprintf("fun (%s) %s", strings.Join(s.typesToStrings(t.Params), ", "), s.TypeToString(t.Returns))
	}
	panic("unreachable")
}

func findIdTypeInFile(typeId TypeId, f *CheckedFile, checked map[*CheckedFile]struct{}) string {
	if _, checked := checked[f]; checked {
		return ""
	}
	checked[f] = struct{}{}
	if found := findIdTypeInScope(typeId, f.GlobalScope); found != "" {
		return found
	}
	for _, imp := range f.Imports {
		if found := findIdTypeInFile(typeId, imp.File, checked); found != "" {
			return found
		}
	}
	panic("unreachable")
}

func findIdTypeInScope(typeId TypeId, s *Scope) string {
	for name, t := range s.Types {
		if t.TypeId == typeId {
			return name
		}
	}
	for _, child := range s.Children {
		if t := findIdTypeInScope(typeId, child); t != "" {
			return t
		}
	}
	return ""
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

func (s *Scope) findAndRenameMethod(name string, newName string) bool {
	if t, ok := s.Methods[name]; ok {
		delete(s.Methods, name)
		s.Methods[newName] = t
		return true
	}
	for _, child := range s.Children {
		if child.findAndRenameMethod(name, newName) {
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
	Methods     []*CheckedMethodDef
	ExternFuns  []*CheckedExternFunDef
	Structs     []*CheckedStructDef
	Typealiases []*CheckedTypealiasDef
	Types       *[]Type
	GlobalScope *Scope
}

func NewCheckedCompilationUnit(filename string) *CheckedFile {
	types := &[]Type{}
	c := &CheckedFile{
		Filename:    filename,
		Imports:     make([]*CheckedImport, 0),
		Funs:        make([]*CheckedFunDef, 0),
		Methods:     make([]*CheckedMethodDef, 0),
		Structs:     make([]*CheckedStructDef, 0),
		Types:       types,
		GlobalScope: NewScope(nil),
	}
	c.GlobalScope.File = c
	panicIf(len(*c.Types) != int(UNIT_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "()"}, &BuildinType{})
	panicIf(len(*c.Types) != int(INT_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "int"}, &BuildinType{})
	panicIf(len(*c.Types) != int(INT8_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "int8"}, &BuildinType{})
	panicIf(len(*c.Types) != int(INT16_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "int16"}, &BuildinType{})
	panicIf(len(*c.Types) != int(INT32_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "int32"}, &BuildinType{})
	panicIf(len(*c.Types) != int(INT64_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "int64"}, &BuildinType{})
	panicIf(len(*c.Types) != int(UINT_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "uint"}, &BuildinType{})
	panicIf(len(*c.Types) != int(UINT8_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "uint8"}, &BuildinType{})
	panicIf(len(*c.Types) != int(UINT16_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "uint16"}, &BuildinType{})
	panicIf(len(*c.Types) != int(UINT32_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "uint32"}, &BuildinType{})
	panicIf(len(*c.Types) != int(UINT64_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "uint64"}, &BuildinType{})
	panicIf(len(*c.Types) != int(FLOAT32_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "float32"}, &BuildinType{})
	panicIf(len(*c.Types) != int(FLOAT64_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "float64"}, &BuildinType{})
	panicIf(len(*c.Types) != int(CHAR_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "char"}, &BuildinType{})
	panicIf(len(*c.Types) != int(BOOL_TYPE_ID))
	c.GlobalScope.DefineType(&Token{Content: "bool"}, &BuildinType{})
	defineInlineC(c)
	return c
}

func NewCheckedFile(filename string, types *[]Type) *CheckedFile {
	c := &CheckedFile{
		Filename:    filename,
		Imports:     make([]*CheckedImport, 0),
		Funs:        make([]*CheckedFunDef, 0),
		Structs:     make([]*CheckedStructDef, 0),
		Types:       types,
		GlobalScope: NewScope(nil),
	}
	c.GlobalScope.File = c
	defineInlineC(c)
	return c
}

func defineInlineC(c *CheckedFile) {
	constChar := c.TypeId(&PointerType{
		Type: CHAR_TYPE_ID,
	})
	c.GlobalScope.DefineFunction(&Token{Content: "inlineC"}, &FunctionType{
		Params:  []TypeId{constChar},
		Returns: UNIT_TYPE_ID,
	})
}

func panicIf(b bool) {
	if b {
		panic("unreachable")
	}
}

func (c *CheckedFile) TypeId(t Type) TypeId {
	for i, typ := range *c.Types {
		if reflect.DeepEqual(t, typ) {
			return TypeId(i)
		}
	}
	*c.Types = append(*c.Types, t)
	return TypeId(len(*c.Types) - 1)
}

type CheckedImport struct {
	Name Token
	File *CheckedFile
}

type CheckedFunDef struct {
	Name       *Token
	Params     []CheckedFunParam
	ReturnType TypeId
	Body       *CheckedBlock
}

type CheckedMethodDef struct {
	Typename   *Token
	Name       *Token
	Params     []CheckedFunParam
	ReturnType TypeId
	Body       *CheckedBlock
}

type CheckedFunParam struct {
	Name *Token
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

type CheckedExternFunDef struct {
	Name       *Token
	Params     []CheckedFunParam
	ReturnType TypeId
}

type CheckedTypealiasDef struct {
	Name *Token
	Type TypeId
}

type CheckedStmt interface {
	checkedStmt()
}

type CheckedVar struct {
	Mut   *Token
	Name  *Token
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

type CheckedIf struct {
	Cond     CheckedExpr
	Body     *CheckedBlock
	ElseBody *CheckedBlock
}

type CheckedWhile struct {
	Cond CheckedExpr
	Body *CheckedBlock
}

type CheckedBreak struct {
	Break Token
}

type CheckedContinue struct {
	Continue Token
}

func (c *CheckedVar) checkedStmt()      {}
func (c *CheckedExprStmt) checkedStmt() {}
func (c *CheckedBlock) checkedStmt()    {}
func (c *CheckedReturn) checkedStmt()   {}
func (c *CheckedIf) checkedStmt()       {}
func (c *CheckedWhile) checkedStmt()    {}
func (c *CheckedBreak) checkedStmt()    {}
func (c *CheckedContinue) checkedStmt() {}

type CheckedExpr interface {
	checkedExpr()
	TypeId() TypeId
}

type CheckedUnaryExpr struct {
	Pos
	Operator CheckedUnaryOperator
	Operand  CheckedExpr
	Type     TypeId
}

type CheckedUnaryOperator int

const INVALID_UNARY_OPERATOR CheckedUnaryOperator = -1

const (
	CHECKED_UNARY_PLUS CheckedUnaryOperator = iota
	CHECKED_NEGATE
	CHECKED_ADDRESS
	CHECKED_DEREF
)

type CheckedBinaryExpr struct {
	Left  CheckedExpr
	Op    CheckedBinaryOperator
	Right CheckedExpr
	Type  TypeId
}

type CheckedBinaryOperator int

const INVALID_BINARY_OPERATOR CheckedBinaryOperator = -1

const (
	CHECKED_ADD CheckedBinaryOperator = iota
	CHECKED_SUBTRACT
	CHECKED_MULTIPLY
	CHECKED_DIVIDE
	CHECKED_EQUALS
	CHECKED_NOTEQUALS
	CHECKED_LESSTHAN
	CHECKED_LESSOREQUAL
	CHECKED_GREATERTHAN
	CHECKED_GREATEROREQUAL
	CHECKED_ASSIGN
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

type CheckedIdExpr struct {
	Id   *Token
	Type TypeId
	Pos
}

type CheckedCallExpr struct {
	Callee CheckedExpr
	Args   []CheckedExpr
	Type   TypeId
}

type CheckedStructInitExpr struct {
	Type   TypeId
	Fields []CheckedStructInitField
	Pos
}

type CheckedStructInitField struct {
	Name  Token
	Value CheckedExpr
}

type CheckedMemberAccessExpr struct {
	Object CheckedExpr
	Member Token
	Type   TypeId
}

type CheckedModuleAccessExpr struct {
	Module Token
	Member CheckedExpr
	Type   TypeId
}

type CheckedAsExpr struct {
	Value CheckedExpr
	Type  TypeId
}

type CheckedMethodExpr struct {
	Object CheckedExpr
	Method *Token
	Type   TypeId
}

func (c *CheckedUnaryExpr) checkedExpr()        {}
func (c *CheckedBinaryExpr) checkedExpr()       {}
func (c *CheckedGroupedExpr) checkedExpr()      {}
func (c *CheckedLiteralExpr) checkedExpr()      {}
func (c *CheckedIdExpr) checkedExpr()           {}
func (c *CheckedCallExpr) checkedExpr()         {}
func (c *CheckedStructInitExpr) checkedExpr()   {}
func (c *CheckedMemberAccessExpr) checkedExpr() {}
func (c *CheckedModuleAccessExpr) checkedExpr() {}
func (c *CheckedAsExpr) checkedExpr()           {}
func (c *CheckedMethodExpr) checkedExpr()       {}

func (c *CheckedUnaryExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedBinaryExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedGroupedExpr) TypeId() TypeId {
	return c.Inner.TypeId()
}
func (c *CheckedLiteralExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedIdExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedCallExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedStructInitExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedMemberAccessExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedModuleAccessExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedAsExpr) TypeId() TypeId {
	return c.Type
}
func (c *CheckedMethodExpr) TypeId() TypeId {
	return c.Type
}

type TypeId int

const NOT_FOUND TypeId = -1
const NEVER_TYPE_ID TypeId = -2

const (
	UNIT_TYPE_ID TypeId = iota
	INT_TYPE_ID
	INT8_TYPE_ID
	INT16_TYPE_ID
	INT32_TYPE_ID
	INT64_TYPE_ID
	UINT_TYPE_ID
	UINT8_TYPE_ID
	UINT16_TYPE_ID
	UINT32_TYPE_ID
	UINT64_TYPE_ID
	FLOAT32_TYPE_ID
	FLOAT64_TYPE_ID
	CHAR_TYPE_ID
	BOOL_TYPE_ID
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
