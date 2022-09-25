package wall_test

import (
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
)

func TestCheckForDuplications(t *testing.T) {
	notypes := []wall.DefNode{
		&wall.FunDef{
			Id: wall.Token{Content: []byte("main")},
		},
		&wall.ImportDef{
			Name: wall.Token{Content: []byte("main")},
		},
		&wall.ParsedImportDef{
			ImportDef: wall.ImportDef{
				Name: wall.Token{Content: []byte("main")},
			},
		},
	}
	types := []wall.DefNode{
		&wall.StructDef{
			Name: wall.Token{Content: []byte("main")},
		},
	}
	for _, ntp := range notypes {
		for _, tp := range types {
			f := &wall.FileNode{
				Defs: []wall.DefNode{
					ntp, tp,
				},
			}
			assert.NoError(t, wall.CheckForDuplications(f))
		}
	}
	for _, ntp1 := range notypes {
		for _, ntp2 := range notypes {
			f := &wall.FileNode{
				Defs: []wall.DefNode{
					ntp1, ntp2,
				},
			}
			assert.Error(t, wall.CheckForDuplications(f))
		}
	}
	for _, tp1 := range types {
		for _, tp2 := range types {
			f := &wall.FileNode{
				Defs: []wall.DefNode{
					tp1, tp2,
				},
			}
			assert.Error(t, wall.CheckForDuplications(f))
		}
	}
}

func TestCheckImports(t *testing.T) {
	fileA := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.ParsedImportDef{
				ImportDef: wall.ImportDef{
					Name: wall.Token{Content: []byte("B")},
				},
				ParsedNode: nil,
			},
		},
	}
	fileB := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.ParsedImportDef{
				ImportDef: wall.ImportDef{
					Name: wall.Token{Content: []byte("A")},
				},
				ParsedNode: fileA,
			},
		},
	}
	fileA.Defs[0].(*wall.ParsedImportDef).ParsedNode = fileB
	moduleA := wall.NewModule()
	wall.CheckImports(fileA, moduleA)
	moduleB := moduleA.GlobalScope.Import("B")
	if assert.NotNil(t, moduleB) {
		moduleA2 := moduleB.GlobalScope.Import("A")
		assert.Equal(t, moduleA, moduleA2)
	}
}

func TestCheckTypesSignatures(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.StructDef{
				Name: wall.Token{Content: []byte("a")},
			},
		},
	}
	mod := wall.NewModule()
	wall.CheckTypesSignatures(file, mod)
	typ, _ := mod.GlobalScope.Type("a")
	assert.Equal(t, typ, wall.NewStructType())
}

func TestCheckFunctionsSignatures(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.FunDef{
				Fun: wall.Token{},
				Id:  wall.Token{Content: []byte("a")},
				Params: []wall.FunParam{{
					Type: &wall.IdTypeNode{Token: wall.Token{Content: []byte("A")}},
				}},
				ReturnType: &wall.IdTypeNode{Token: wall.Token{Content: []byte("B")}},
			},
		},
	}
	mod := wall.NewModule()
	typeIdA := mod.DefType("A", &wall.StructType{})
	typeIdB := mod.DefType("B", &wall.StructType{})
	assert.NoError(t, wall.CheckFunctionsSignatures(file, mod))
	if fun, ok := mod.GlobalScope.Funs["a"]; assert.Equal(t, ok, true) {
		assert.Equal(t, fun, mod.TypeId(&wall.FunctionType{
			Args: []wall.TypeId{
				typeIdA,
			},
			Returns: typeIdB,
		}))
	}
}

func TestCheckTypesContents(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.StructDef{
				Name: wall.Token{Content: []byte("Employee")},
				Fields: []wall.StructField{
					{
						Name: wall.Token{Content: []byte("name")},
						Type: &wall.IdTypeNode{
							Token: wall.Token{Content: []byte("String")},
						},
					},
					{
						Name: wall.Token{Content: []byte("age")},
						Type: &wall.IdTypeNode{
							Token: wall.Token{Content: []byte("int")},
						},
					},
				},
			},
		},
	}
	mod := wall.NewModule()
	employeeTypeId := mod.DefType("Employee", wall.NewStructType())
	stringTypeId := mod.DefType("String", wall.NewStructType())
	intTypeId := mod.DefType("int", wall.NewStructType())
	assert.NoError(t, wall.CheckTypesContents(file, mod))
	if assert.IsType(t, mod.Types[employeeTypeId], &wall.StructType{}) {
		emp := mod.Types[employeeTypeId].(*wall.StructType)
		assert.Equal(t, emp.Fields, map[string]wall.TypeId{
			"name": stringTypeId,
			"age":  intTypeId,
		})
	}
}

type checkBlocksTest struct {
	block      *wall.BlockStmt
	returnType wall.TypeNode
}

var checkBlocksTests = []checkBlocksTest{
	{
		block: &wall.BlockStmt{
			Stmts: []wall.StmtNode{},
		},
		returnType: nil,
	},
	{
		block: &wall.BlockStmt{
			Stmts: []wall.StmtNode{},
		},
		returnType: &wall.IdTypeNode{
			Token: wall.Token{Content: []byte("()")},
		},
	},
	{
		block: &wall.BlockStmt{
			Stmts: []wall.StmtNode{
				&wall.ExprStmt{
					Expr: &wall.LiteralExprNode{
						Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
					},
				},
			},
		},
		returnType: &wall.IdTypeNode{
			Token: wall.Token{Content: []byte("int")},
		},
	},
	{
		block: &wall.BlockStmt{
			Stmts: []wall.StmtNode{
				&wall.ExprStmt{
					Expr: &wall.LiteralExprNode{
						Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
					},
				},
				&wall.ExprStmt{
					Expr: &wall.LiteralExprNode{
						Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
					},
				},
			},
		},
		returnType: &wall.IdTypeNode{
			Token: wall.Token{Content: []byte("float")},
		},
	},
}

func TestCheckBlocks(t *testing.T) {
	for _, test := range checkBlocksTests {
		file := &wall.FileNode{
			Defs: []wall.DefNode{
				&wall.FunDef{
					Id:         wall.Token{Content: []byte("a")},
					Body:       test.block,
					ReturnType: test.returnType,
				},
			},
		}
		mod := wall.NewModule()
		assert.NoError(t, wall.CheckFunctionsSignatures(file, mod))
		assert.NoError(t, wall.CheckBlocks(file, mod))
	}
}

type checkStmtTest struct {
	stmt   wall.StmtNode
	typeid wall.TypeId
}

var checkStmtTests = []checkStmtTest{
	{
		stmt: &wall.VarStmt{
			Id: wall.Token{Content: []byte("")},
			Value: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.UNIT_TYPE_ID,
	},
	{
		stmt: &wall.ExprStmt{
			Expr: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
}

func TestCheckStmt(t *testing.T) {
	for _, test := range checkStmtTests {
		mod := wall.NewModule()
		typ, err := wall.CheckStmt(test.stmt, mod.GlobalScope)
		assert.NoError(t, err)
		assert.Equal(t, typ, test.typeid)
	}
}

type checkExprTest struct {
	expr   wall.ExprNode
	typeid wall.TypeId
}

var checkExprTests = []checkExprTest{
	{
		expr: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0.0")},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.GroupedExprNode{
			Inner: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.UnaryExprNode{
			Operator: wall.Token{Kind: wall.PLUS},
			Operand: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.UnaryExprNode{
			Operator: wall.Token{Kind: wall.PLUS},
			Operand: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.UnaryExprNode{
			Operator: wall.Token{Kind: wall.MINUS},
			Operand: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.UnaryExprNode{
			Operator: wall.Token{Kind: wall.MINUS},
			Operand: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.PLUS},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.MINUS},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.STAR},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.SLASH},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.PLUS},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.MINUS},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.STAR},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.BinaryExprNode{
			Left: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.SLASH},
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
}

func TestCheckExpr(t *testing.T) {
	for _, test := range checkExprTests {
		mod := wall.NewModule()
		typ, err := wall.CheckExpr(test.expr, mod.GlobalScope)
		if assert.NoError(t, err) {
			assert.Equal(t, test.typeid, typ)
		}
	}
}

func TestCheckVarStmt(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.FunDef{
				Id: wall.Token{Content: []byte("")},
				Body: &wall.BlockStmt{
					Stmts: []wall.StmtNode{
						&wall.VarStmt{
							Id: wall.Token{Content: []byte("a")},
							Value: &wall.LiteralExprNode{
								Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
							},
						},
						&wall.ExprStmt{
							Expr: &wall.LiteralExprNode{
								Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
							},
						},
					},
				},
				ReturnType: &wall.IdTypeNode{
					Token: wall.Token{Content: []byte("int")},
				},
			},
		},
	}
	mod := wall.NewModule()
	assert.NoError(t, wall.CheckFunctionsSignatures(file, mod))
	assert.NoError(t, wall.CheckBlocks(file, mod))
}
