package wall

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

func CodegenCompilationUnit(c *CheckedFile) string {
	WallPrefixesToGlobalNames(c)
	var result strings.Builder
	fmt.Fprintf(&result, "/* source filename: %s */\n", c.Filename)
	result.WriteString("/* type declarations */\n")
	result.WriteString(CodegenTypeDeclarations(c))
	result.WriteString("/* function typedefs */\n")
	result.WriteString(CodegenFuncTypedefs(c))
	result.WriteString("/* function declarations */\n")
	result.WriteString(CodegenFuncDeclarations(c))
	result.WriteString("/* type definitions */\n")
	result.WriteString(CodegenTypeDefinitions(c))
	result.WriteString("/* function definitions */\n")
	result.WriteString(CodegenFuncDefinitions(c))
	return result.String()
}

func CodegenFuncTypedefs(c *CheckedFile) string {
	return codegenFuncTypedefs(c, make(map[*CheckedFile]struct{}))
}

func codegenFuncTypedefs(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) string {
	if _, ok := checkedFiles[c]; ok {
		return ""
	}
	checkedFiles[c] = struct{}{}
	var builder strings.Builder
	for i, typ := range c.Types {
		if typ, ok := typ.(*FunctionType); ok {
			fmt.Fprintf(&builder, "typedef %s (*%s)(", CodegenType(typ.Returns, c.GlobalScope), cFuncTypeId(i, c.Filename))
			if len(typ.Params) == 0 {
				builder.WriteString(CodegenType(UNIT_TYPE_ID, c.GlobalScope))
			}
			for i, param := range typ.Params {
				builder.WriteString(CodegenType(param, c.GlobalScope))
				if i < len(typ.Params)-1 {
					builder.WriteString(", ")
				}
			}
			builder.WriteString(");\n")
		}
	}
	for _, imp := range c.Imports {
		builder.WriteString(codegenFuncTypedefs(imp.File, checkedFiles))
	}
	return builder.String()
}

func cFuncTypeId(id int, filename string) string {
	return cId(fmt.Sprintf("%s_FUNC_TYPE_%d", moduleNameFromFilename(filename), id))
}

func WallPrefixesToGlobalNames(c *CheckedFile) {
	moduleNamesToGlobalNames(c, make(map[*CheckedFile]struct{}))
	wallPrefixesToGlobalNames(c, make(map[*CheckedFile]struct{}))
}

func CodegenTypeDeclarations(c *CheckedFile) string {
	return codegenTypeDeclarations(c, make(map[*CheckedFile]struct{}))
}

func CodegenFuncDeclarations(c *CheckedFile) string {
	return codegenFuncDeclarations(c, make(map[*CheckedFile]struct{}))
}

func CodegenTypeDefinitions(c *CheckedFile) string {
	return codegenTypeDefinitions(c, make(map[*CheckedFile]struct{}))
}

func CodegenFuncDefinitions(c *CheckedFile) string {
	return codegenFuncDefinitions(c, make(map[*CheckedFile]struct{}))
}

func CodegenExpr(expr CheckedExpr, s *Scope) string {
	switch expr := expr.(type) {
	case *CheckedUnaryExpr:
		return codegenUnaryExpr(expr, s)
	case *CheckedBinaryExpr:
		return codegenBinaryExpr(expr, s)
	case *CheckedGroupedExpr:
		return codegenGroupedExpr(expr, s)
	case *CheckedLiteralExpr:
		return codegenLiteralExpr(expr, s)
	case *CheckedIdExpr:
		return codegenIdExpr(expr, s)
	case *CheckedCallExpr:
		return codegenCallExpr(expr, s)
	case *CheckedStructInitExpr:
		return codegenStructInitExpr(expr, s)
	}
	panic("unreachable")
}

func codegenStructInitExpr(expr *CheckedStructInitExpr, s *Scope) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "(%s) {\n", string(expr.Id.Content))
	for i, field := range expr.Fields {
		fmt.Fprintf(&builder, ".%s = %s", field.Name.Content, CodegenExpr(field.Value, s))
		if i+1 < len(expr.Fields) {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("}")
	return builder.String()
}

func codegenIdExpr(expr *CheckedIdExpr, s *Scope) string {
	return string(expr.Id.Content)
}

func codegenUnaryExpr(expr *CheckedUnaryExpr, s *Scope) string {
	if isBuildin(expr.Operand.TypeId()) {
		switch expr.Operator {
		case CHECKED_NEGATE:
			return fmt.Sprintf("-%s", CodegenExpr(expr.Operand, s))
		}
	}
	panic("unreachable")
}

func codegenBinaryExpr(expr *CheckedBinaryExpr, s *Scope) string {
	if isBuildin(expr.Left.TypeId()) {
		switch expr.Op {
		case CHECKED_ADD:
			return fmt.Sprintf("%s+%s", CodegenExpr(expr.Left, s), CodegenExpr(expr.Right, s))
		case CHECKED_SUBTRACT:
			return fmt.Sprintf("%s-%s", CodegenExpr(expr.Left, s), CodegenExpr(expr.Right, s))
		case CHECKED_MULTIPLY:
			return fmt.Sprintf("%s*%s", CodegenExpr(expr.Left, s), CodegenExpr(expr.Right, s))
		case CHECKED_DIVIDE:
			return fmt.Sprintf("%s/%s", CodegenExpr(expr.Left, s), CodegenExpr(expr.Right, s))
		}
	}
	panic("unreachable")
}

func codegenGroupedExpr(expr *CheckedGroupedExpr, s *Scope) string {
	return fmt.Sprintf("(%s)", CodegenExpr(expr.Inner, s))
}

func codegenLiteralExpr(expr *CheckedLiteralExpr, s *Scope) string {
	if expr.Literal.Kind == STRING {
		s := string(expr.Literal.Content)
		s = strings.ReplaceAll(s, "\n", "\n\"")
		return fmt.Sprintf("\"%s\"", s)
	}
	if expr.Literal.Kind == TRUE {
		return "1"
	}
	if expr.Literal.Kind == FALSE {
		return "0"
	}
	return string(expr.Literal.Content)
}

func codegenCallExpr(expr *CheckedCallExpr, s *Scope) string {
	callee := CodegenExpr(expr.Callee, s)
	if callee == "inlineC" {
		inlineC := CodegenExpr(expr.Args[0], s)
		inlineC = strings.ReplaceAll(inlineC, "\"", "")
		return inlineC
	}
	var builder strings.Builder
	builder.WriteString(callee)
	builder.WriteString("(")
	for i, arg := range expr.Args {
		builder.WriteString(CodegenExpr(arg, s))
		if i < len(expr.Args)-1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(")")
	return builder.String()
}

func isBuildin(id TypeId) bool {
	return id == UNIT_TYPE_ID || id == INT_TYPE_ID || id == FLOAT_TYPE_ID
}

func CodegenStmt(stmt CheckedStmt, s *Scope) string {
	switch stmt := stmt.(type) {
	case *CheckedVar:
		return codegenVarStmt(stmt, s)
	case *CheckedExprStmt:
		return codegenExprStmt(stmt, s)
	case *CheckedBlock:
		return codegenBlock(stmt, s)
	case *CheckedReturn:
		return codegenReturn(stmt, s)
	case *CheckedIf:
		return codegenIf(stmt, s)
	}
	panic("unimplemented")
}

func codegenVarStmt(stmt *CheckedVar, s *Scope) string {
	if stmt.Value == nil {
		t := CodegenType(stmt.Type, s)
		return fmt.Sprintf("%s %s = (%s) {0};", t, string(stmt.Name.Content), t)
	}
	val := CodegenExpr(stmt.Value, s)
	t := CodegenType(stmt.Type, s)
	return fmt.Sprintf("%s %s = %s;", t, string(stmt.Name.Content), val)
}

func codegenExprStmt(stmt *CheckedExprStmt, s *Scope) string {
	return CodegenExpr(stmt.Expr, s) + ";"
}

func codegenBlock(b *CheckedBlock, s *Scope) string {
	var builder strings.Builder
	builder.WriteString("{\n")
	for _, stmt := range b.Stmts {
		builder.WriteString(CodegenStmt(stmt, s))
		builder.WriteString("\n")
	}
	builder.WriteString("}\n")
	return builder.String()
}

func codegenReturn(r *CheckedReturn, s *Scope) string {
	if r.Value == nil {
		return "return;"
	}
	arg := CodegenExpr(r.Value, s)
	return fmt.Sprintf("return %s;", arg)
}

func codegenIf(i *CheckedIf, s *Scope) string {
	if i.ElseBody != nil {
		return fmt.Sprintf("if (%s) %s else %s", CodegenExpr(i.Cond, s), codegenBlock(i.Body, s), codegenBlock(i.ElseBody, s))
	} else {
		return fmt.Sprintf("if (%s) %s", CodegenExpr(i.Cond, s), codegenBlock(i.Body, s))
	}
}

func codegenFuncDefinitions(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) string {
	if _, ok := checkedFiles[c]; ok {
		return ""
	}
	checkedFiles[c] = struct{}{}
	var builder strings.Builder
	for _, def := range c.Funs {
		builder.WriteString(codegenFunDef(def, c.GlobalScope))
	}
	for _, imp := range c.Imports {
		builder.WriteString(codegenFuncDefinitions(imp.File, checkedFiles))
	}
	return builder.String()
}

func codegenFunDef(def *CheckedFunDef, s *Scope) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s %s(", CodegenType(def.ReturnType, s), def.Name.Content)
	if len(def.Params) == 0 {
		builder.WriteString(CodegenType(UNIT_TYPE_ID, s))
	}
	for i, param := range def.Params {
		fmt.Fprintf(&builder, "%s %s", CodegenType(param.Type, s), param.Name.Content)
		if i < len(def.Params)-1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(") ")
	builder.WriteString(codegenBlock(def.Body, s))
	return builder.String()
}

func wallPrefixesToGlobalNames(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) {
	if _, ok := checkedFiles[c]; ok {
		return
	}
	checkedFiles[c] = struct{}{}
	for _, def := range c.Structs {
		if !c.GlobalScope.findAndRenameType(string(def.Name.Content), attachWallPrefix(def.Name.Content)) {
			panic(fmt.Sprintf("type not found: %s", def.Name.Content))
		}
		def.Name.Content = []byte(attachWallPrefix(def.Name.Content))
	}
	for _, def := range c.Funs {
		if len(checkedFiles) == 1 /* this is a root module */ && bytes.Equal(def.Name.Content, []byte("main")) {
			continue
		}
		if !c.GlobalScope.findAndRenameFun(string(def.Name.Content), attachWallPrefix(def.Name.Content)) {
			panic(fmt.Sprintf("fun not found: %s", def.Name.Content))
		}
		def.Name.Content = []byte(attachWallPrefix(def.Name.Content))
	}
	for _, imp := range c.Imports {
		wallPrefixesToGlobalNames(imp.File, checkedFiles)
	}
}

func moduleNamesToGlobalNames(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) {
	if _, ok := checkedFiles[c]; ok {
		return
	}
	checkedFiles[c] = struct{}{}
	for _, def := range c.Structs {
		if !c.GlobalScope.findAndRenameType(string(def.Name.Content), attachModuleName(def.Name.Content, def.Name.Filename)) {
			panic("type not found")
		}
		def.Name.Content = []byte(attachModuleName(def.Name.Content, def.Name.Filename))
	}
	for _, def := range c.Funs {
		if len(checkedFiles) == 1 /* this is a root module */ && bytes.Equal(def.Name.Content, []byte("main")) {
			continue
		}
		if !c.GlobalScope.findAndRenameFun(string(def.Name.Content), attachModuleName(def.Name.Content, def.Name.Filename)) {
			panic("fun not found")
		}
		def.Name.Content = []byte(attachModuleName(def.Name.Content, def.Name.Filename))
	}
	for _, imp := range c.Imports {
		moduleNamesToGlobalNames(imp.File, checkedFiles)
	}
}

func attachModuleName(name []byte, filename string) string {
	return moduleNameFromFilename(filename) + "_" + string(name)
}

func moduleNameFromFilename(filename string) string {
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "SUPER")
	filename = strings.ReplaceAll(filename, ":", "")
	return filename
}

func attachWallPrefix(name []byte) string {
	return "WALL_" + string(name)
}

func codegenTypeDeclarations(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) string {
	if _, ok := checkedFiles[c]; ok {
		return ""
	}
	checkedFiles[c] = struct{}{}
	var builder strings.Builder
	for _, def := range c.Structs {
		id := string(def.Name.Content)
		fmt.Fprintf(&builder, "typedef struct %s %s;\n", id, id)
	}
	for _, imp := range c.Imports {
		builder.WriteString(codegenTypeDeclarations(imp.File, checkedFiles))
	}
	return builder.String()
}

func codegenFuncDeclarations(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) string {
	if _, ok := checkedFiles[c]; ok {
		return ""
	}
	checkedFiles[c] = struct{}{}
	var builder strings.Builder
	for _, def := range c.Funs {
		fmt.Fprintf(&builder, "%s %s(", CodegenType(def.ReturnType, c.GlobalScope), string(def.Name.Content))
		if len(def.Params) == 0 {
			builder.WriteString(CodegenType(UNIT_TYPE_ID, c.GlobalScope))
		}
		for i, param := range def.Params {
			fmt.Fprintf(&builder, "%s", CodegenType(param.Type, c.GlobalScope))
			if i < len(def.Params)-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString(");\n")
	}
	for _, imp := range c.Imports {
		builder.WriteString(codegenFuncDeclarations(imp.File, checkedFiles))
	}
	return builder.String()
}

func codegenTypeDefinitions(c *CheckedFile, checkedFiles map[*CheckedFile]struct{}) string {
	if _, ok := checkedFiles[c]; ok {
		return ""
	}
	checkedFiles[c] = struct{}{}
	var builder strings.Builder
	for _, imp := range c.Imports {
		builder.WriteString(codegenTypeDefinitions(imp.File, checkedFiles))
	}
	for _, def := range c.Structs {
		builder.WriteString(CodegenStructDef(string(def.Name.Content), def.Fields, c.GlobalScope))
	}
	return builder.String()
}

func CodegenStructDef(id string, fields []CheckedStructField, s *Scope) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "struct %s {\n", id)
	for _, field := range fields {
		fmt.Fprintf(&builder, "%s %s;\n", CodegenType(field.Type, s), field.Name.Content)
	}
	builder.WriteString("};\n")
	return builder.String()
}

func CodegenType(id TypeId, s *Scope) string {
	t := s.File.Types[id]
	switch t := t.(type) {
	case *BuildinType:
		switch id {
		case UNIT_TYPE_ID:
			return "void"
		case INT_TYPE_ID, BOOL_TYPE_ID:
			return "int"
		case FLOAT_TYPE_ID:
			return "float"
		case CHAR_TYPE_ID:
			return "char"
		default:
			panic("unreachable")
		}
	case *IdType, *StructType:
		return s.TypeToString(id)
	case *PointerType:
		return CodegenType(t.Type, s) + "*"
	case *FunctionType:
		return cFuncTypeId(int(id), s.File.Filename)
	}
	panic("unreachable")
}

func cId(id string) string {
	return fmt.Sprintf("WALL_%s", id)
}
