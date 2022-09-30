package wall_test

import (
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
)

func TestCheckImports(t *testing.T) {
	fileA := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
				Name: wall.Token{Content: []byte("B")},
				File: nil,
			},
		},
	}
	fileB := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
				Name: wall.Token{Content: []byte("A")},
				File: fileA,
			},
		},
	}
	fileA.Defs[0].(*wall.ParsedImport).File = fileB
	checkedFileA := wall.NewCheckedFile("")
	assert.NoError(t, wall.CheckImports(fileA, checkedFileA))
	checkedFileB := checkedFileA.Imports[checkedFileA.GlobalScope.Imports["B"]].File
	if assert.NotNil(t, checkedFileB) {
		checkedFileA2 := checkedFileB.Imports[checkedFileB.GlobalScope.Imports["A"]].File
		assert.Equal(t, checkedFileA, checkedFileA2)
	}
}

func TestCheckTypeSignatures(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedStructDef{
				Name: wall.Token{Content: []byte("a")},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	typ := checkedFile.Types[checkedFile.GlobalScope.Types["a"].TypeId]
	assert.Equal(t, typ, wall.NewStructType())
	assert.Equal(t, len(checkedFile.Structs), 1)
}

func TestCheckFunctionSignatures(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Fun: wall.Token{},
				Id:  wall.Token{Content: []byte("a")},
				Params: []wall.ParsedFunParam{{
					Type: &wall.ParsedIdType{Token: wall.Token{Content: []byte("A")}},
				}},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Content: []byte("B")}},
			},
			&wall.ParsedExternFunDef{
				Extern: wall.Token{},
				Fun:    wall.Token{},
				Name:   wall.Token{Content: []byte("b")},
				Params: []wall.ParsedFunParam{{
					Type: &wall.ParsedIdType{Token: wall.Token{Content: []byte("A")}},
				}},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Content: []byte("B")}},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, checkedFile.GlobalScope.DefineType(&wall.Token{Content: []byte("A")}, wall.NewStructType()))
	typeIdA := checkedFile.GlobalScope.Types["A"].TypeId
	assert.NoError(t, checkedFile.GlobalScope.DefineType(&wall.Token{Content: []byte("B")}, wall.NewStructType()))
	typeIdB := checkedFile.GlobalScope.Types["B"].TypeId
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	if fun, ok := checkedFile.GlobalScope.Funs["a"]; assert.Equal(t, ok, true) {
		assert.Equal(t, checkedFile.Types[fun.TypeId], &wall.FunctionType{
			Params: []wall.TypeId{
				typeIdA,
			},
			Returns: typeIdB,
		})
	}
	if fun, ok := checkedFile.GlobalScope.Funs["b"]; assert.Equal(t, ok, true) {
		assert.Equal(t, checkedFile.Types[fun.TypeId], &wall.FunctionType{
			Params: []wall.TypeId{
				typeIdA,
			},
			Returns: typeIdB,
		})
	}
	assert.Equal(t, len(checkedFile.Funs), 1)
}

func TestCheckTypesContents(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedStructDef{
				Name: wall.Token{Content: []byte("Employee")},
				Fields: []wall.ParsedStructField{
					{
						Name: wall.Token{Content: []byte("name")},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: []byte("String")},
						},
					},
					{
						Name: wall.Token{Content: []byte("age")},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: []byte("int")},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, checkedFile.GlobalScope.DefineType(&wall.Token{Content: []byte("String")}, wall.NewStructType()))
	stringTypeId := checkedFile.GlobalScope.Types["String"].TypeId
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckTypeContents(file, checkedFile))
	employeeTypeId := checkedFile.GlobalScope.Types["Employee"].TypeId
	if assert.IsType(t, wall.NewStructType(), checkedFile.Types[employeeTypeId]) {
		emp := checkedFile.Types[employeeTypeId].(*wall.StructType)
		assert.Equal(t, map[string]wall.TypeId{
			"name": stringTypeId,
			"age":  wall.INT_TYPE_ID,
		}, emp.Fields)
	}
}

type checkBlocksTest struct {
	block      *wall.ParsedBlock
	returnType wall.ParsedType
}

var checkBlocksTests = []checkBlocksTest{
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{},
		},
		returnType: nil,
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: []byte("()")},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedReturn{},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: []byte("()")},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedReturn{
					Arg: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
					},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: []byte("int")},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedBlock{},
				&wall.ParsedReturn{
					Arg: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
					},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: []byte("int")},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedBlock{},
				&wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{&wall.ParsedReturn{
						Arg: &wall.ParsedLiteralExpr{
							Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
						},
					}},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: []byte("int")},
		},
	},
}

func TestCheckBlocks(t *testing.T) {
	for _, test := range checkBlocksTests {
		file := &wall.ParsedFile{
			Defs: []wall.ParsedDef{
				&wall.ParsedFunDef{
					Id:         wall.Token{Content: []byte("a")},
					Body:       test.block,
					ReturnType: test.returnType,
				},
			},
		}
		checkedFile := wall.NewCheckedFile("")
		assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
		assert.NoError(t, wall.CheckBlocks(file, checkedFile))
	}
}

var checkStmtTests = []wall.ParsedStmt{
	&wall.ParsedVar{
		Id: wall.Token{Content: []byte("")},
		Value: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
		},
	},
	&wall.ParsedExprStmt{
		Expr: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
		},
	},
}

func TestCheckStmt(t *testing.T) {
	for _, test := range checkStmtTests {
		checkedFile := wall.NewCheckedFile("")
		_, err := wall.CheckStmt(test, checkedFile.GlobalScope, &wall.MayReturn{Type: wall.UNIT_TYPE_ID})
		assert.NoError(t, err)
	}
}

type checkExprTest struct {
	expr   wall.ParsedExpr
	typeid wall.TypeId
}

var checkExprTests = []checkExprTest{
	{
		expr: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0.0")},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.ParsedGroupedExpr{
			Inner: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.ParsedUnaryExpr{
			Operator: wall.Token{Kind: wall.MINUS},
			Operand: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.ParsedUnaryExpr{
			Operator: wall.Token{Kind: wall.MINUS},
			Operand: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.PLUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.MINUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.STAR},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.SLASH},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
			},
		},
		typeid: wall.INT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.PLUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.MINUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.STAR},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
			Op: wall.Token{Kind: wall.SLASH},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0")},
			},
		},
		typeid: wall.FLOAT_TYPE_ID,
	},
}

func TestCheckExpr(t *testing.T) {
	for _, test := range checkExprTests {
		checkedFile := wall.NewCheckedFile("")
		expr, err := wall.CheckExpr(test.expr, checkedFile.GlobalScope)
		if assert.NoError(t, err) {
			assert.Equal(t, test.typeid, expr.TypeId())
		}
	}
}

func TestCheckStringLiteralExpr(t *testing.T) {
	checkedFile := wall.NewCheckedFile("")
	expr, err := wall.CheckExpr(&wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.STRING, Content: []byte("\"ABC\"")},
	}, checkedFile.GlobalScope)
	if assert.NoError(t, err) {
		assert.Equal(t, checkedFile.TypeId(&wall.PointerType{
			Type: wall.CHAR_TYPE_ID,
		}), expr.TypeId())
	}
}

func TestCheckVarStmt(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Id: wall.Token{Content: []byte("")},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedVar{
							Id: wall.Token{Content: []byte("a")},
							Value: &wall.ParsedLiteralExpr{
								Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
							},
						},
						&wall.ParsedReturn{
							Arg: &wall.ParsedIdExpr{
								Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
							},
						},
					},
				},
				ReturnType: &wall.ParsedIdType{
					Token: wall.Token{Content: []byte("int")},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}

func TestCheckCallExpr(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Id: wall.Token{Content: []byte("sum")},
				Params: []wall.ParsedFunParam{
					{
						Id: wall.Token{Content: []byte("a")},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: []byte("int")},
						},
					},
					{
						Id: wall.Token{Content: []byte("b")},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: []byte("int")},
						},
					},
				},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Content: []byte("int")}},
				Body:       &wall.ParsedBlock{Stmts: []wall.ParsedStmt{&wall.ParsedReturn{Arg: &wall.ParsedLiteralExpr{Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")}}}}},
			},
			&wall.ParsedFunDef{
				Params:     []wall.ParsedFunParam{},
				ReturnType: nil,
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedCallExpr{
								Callee: &wall.ParsedIdExpr{
									Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("sum")},
								},
								Args: []wall.ParsedExpr{
									&wall.ParsedLiteralExpr{
										Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
									},
									&wall.ParsedLiteralExpr{
										Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}

func TestCheckIdExpr(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Id:     wall.Token{Content: []byte("a")},
				Params: []wall.ParsedFunParam{},
				Body:   &wall.ParsedBlock{},
			},
			&wall.ParsedFunDef{
				Id: wall.Token{Content: []byte("b")},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedIdExpr{
								Token: wall.Token{Content: []byte("a")},
							},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
	checkedFile.Funs[0].Name.Content = []byte("c")
	assert.Equal(t, &wall.CheckedExprStmt{
		Expr: &wall.CheckedIdExpr{
			Id: checkedFile.Funs[0].Name,
			Type: checkedFile.TypeId(&wall.FunctionType{
				Params:  []wall.TypeId{},
				Returns: wall.UNIT_TYPE_ID,
			}),
		},
	}, checkedFile.Funs[1].Body.Stmts[0])
}

func TestCheckStructInitExpr(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedStructDef{
				Name: wall.Token{Content: []byte("Point")},
				Fields: []wall.ParsedStructField{
					{
						Name: wall.Token{Content: []byte("x")},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: []byte("int")},
						},
					},
					{
						Name: wall.Token{Content: []byte("y")},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: []byte("float")},
						},
					},
				},
			},
			&wall.ParsedFunDef{
				Id: wall.Token{Content: []byte("a")},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedStructInitExpr{
								Name: wall.ParsedIdType{Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("Point")}},
								Fields: []wall.ParsedStructInitField{
									{
										Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("x")},
										Value: &wall.ParsedLiteralExpr{
											Token: wall.Token{Kind: wall.INTEGER},
										},
									},
									{
										Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("y")},
										Value: &wall.ParsedLiteralExpr{
											Token: wall.Token{Kind: wall.FLOAT},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedFile("")
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckTypeContents(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}
