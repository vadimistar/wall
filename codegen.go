package wall

import (
	"fmt"

	"tinygo.org/x/go-llvm"
)

func Codegen(f *FileNode) llvm.Module {
	module := llvm.NewModule(f.pos().Filename)
	types := llvmTypes()
	values := make(map[string]codegenValue)
	CodegenFile(f, module, types, values)
	return module
}

type codegenValue struct {
	llvm    llvm.Value
	onStack bool
}

func CodegenFile(f *FileNode, m llvm.Module, types map[string]llvm.Type, values map[string]codegenValue) {
	for _, def := range f.Defs {
		CodegenDef(def, m, types, values)
	}
}

func CodegenDef(def DefNode, module llvm.Module, types map[string]llvm.Type, values map[string]codegenValue) {
	switch def := def.(type) {
	case *FunDef:
		CodegenFunDef(def, module, types, values)
	case *ParsedImportDef:
		CodegenFile(def.ParsedNode, module, types, values)
	case *StructDef:
		CodegenStructDef(def, module, types, values)
	}
}

func CodegenStructDef(s *StructDef, module llvm.Module, types map[string]llvm.Type, values map[string]codegenValue) {
	var llvmTypes []llvm.Type
	for _, field := range s.Fields {
		llvmTypes = append(llvmTypes, CodegenType(field.Type, types))
	}
	structType := llvm.StructType(llvmTypes, false)
	types[string(s.Name.Content)] = structType
}

func CodegenFunDef(f *FunDef, module llvm.Module, types map[string]llvm.Type, values map[string]codegenValue) llvm.Value {
	var paramTypes []llvm.Type
	var paramNames []string
	for _, param := range f.Params {
		paramTypes = append(paramTypes, CodegenType(param.Type, types))
		paramNames = append(paramNames, string(param.Id.Content))
	}
	returnType := llvm.VoidType()
	if f.ReturnType != nil {
		returnType = CodegenType(f.ReturnType, types)
	}
	functionType := llvm.FunctionType(returnType, paramTypes, false)
	fun := llvm.AddFunction(module, string(f.Id.Content), functionType)
	bb := llvm.AddBasicBlock(fun, ".entry")
	builder := llvm.NewBuilder()
	builder.SetInsertPointAtEnd(bb)
	for i, name := range paramNames {
		values[name] = codegenValue{
			llvm:    fun.Param(i),
			onStack: false,
		}
	}
	if len(f.Body.Stmts) == 0 {
		builder.CreateRetVoid()
		return fun
	}
	CodegenBlock(f.Body, builder, types, values)
	return fun
}

func CodegenBlock(block *BlockStmt, builder llvm.Builder, types map[string]llvm.Type, values map[string]codegenValue) {
	for _, stmt := range block.Stmts {
		CodegenStmt(stmt, builder, types, values)
	}
}

func CodegenStmt(stmt StmtNode, builder llvm.Builder, types map[string]llvm.Type, values map[string]codegenValue) {
	switch stmt := stmt.(type) {
	case *VarStmt:
		value := CodegenExpr(stmt.Value, builder, types, values)
		alloca := builder.CreateAlloca(value.Type(), string(stmt.Id.Content))
		builder.CreateStore(value, alloca)
		values[string(stmt.Id.Content)] = codegenValue{
			llvm:    alloca,
			onStack: true,
		}
	case *ExprStmt:
		CodegenExpr(stmt.Expr, builder, types, values)
	case *BlockStmt:
		CodegenBlock(stmt, builder, types, values)
	case *ReturnStmt:
		if stmt.Arg != nil {
			arg := CodegenExpr(stmt.Arg, builder, types, values)
			builder.CreateRet(arg)
		} else {
			builder.CreateRetVoid()
		}
	default:
		panic("unreachable")
	}
}

func CodegenExpr(expr ExprNode, builder llvm.Builder, types map[string]llvm.Type, values map[string]codegenValue) llvm.Value {
	switch expr := expr.(type) {
	case *UnaryExprNode:
		switch expr.Operator.Kind {
		case PLUS:
			return CodegenExpr(expr.Operand, builder, types, values)
		case MINUS:
			operand := CodegenExpr(expr.Operand, builder, types, values)
			return builder.CreateNeg(operand, tempName())
		}
	case *BinaryExprNode:
		left := CodegenExpr(expr.Left, builder, types, values)
		right := CodegenExpr(expr.Left, builder, types, values)
		switch expr.Op.Kind {
		case PLUS:
			return builder.CreateAdd(left, right, tempName())
		case MINUS:
			return builder.CreateSub(left, right, tempName())
		case STAR:
			return builder.CreateMul(left, right, tempName())
		case SLASH:
			return builder.CreateUDiv(left, right, tempName())
		}
	case *GroupedExprNode:
		return CodegenExpr(expr.Inner, builder, types, values)
	case *LiteralExprNode:
		switch expr.Token.Kind {
		case INTEGER:
			return llvm.ConstIntFromString(llvm.Int64Type(), string(expr.Token.Content), 10)
		case FLOAT:
			return llvm.ConstFloatFromString(llvm.DoubleType(), string(expr.Token.Content))
		case IDENTIFIER:
			value := values[string(expr.Token.Content)]
			if value.onStack {
				loaded := builder.CreateLoad(value.llvm.AllocatedType(), value.llvm, tempName())
				return loaded
			}
			return value.llvm
		}
	}
	panic("unreachable")
}

func CodegenType(t TypeNode, types map[string]llvm.Type) llvm.Type {
	switch t := t.(type) {
	case *IdTypeNode:
		return types[string(t.Content)]
	default:
		panic("unreachable")
	}
}

func llvmTypes() map[string]llvm.Type {
	return map[string]llvm.Type{
		"()":    llvm.VoidType(),
		"int":   llvm.Int64Type(),
		"float": llvm.DoubleType(),
	}
}

var tempsCount = 0

func tempName() string {
	tempsCount++
	return fmt.Sprintf(".tmp%d", tempsCount)
}
