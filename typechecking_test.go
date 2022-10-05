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
				Name: wall.Token{Content: "B"},
				File: nil,
			},
		},
	}
	fileB := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
				Name: wall.Token{Content: "A"},
				File: fileA,
			},
		},
	}
	fileA.Defs[0].(*wall.ParsedImport).File = fileB
	checkedFileA := wall.NewCheckedCompilationUnit("")
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
				Name: wall.Token{Content: "a"},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	typ := (*checkedFile.Types)[checkedFile.GlobalScope.Types["a"].TypeId]
	assert.Equal(t, typ, wall.NewStructType())
	assert.Equal(t, len(checkedFile.Structs), 1)
}

func TestCheckFunctionSignatures(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Fun: wall.Token{},
				Id:  wall.Token{Content: "a"},
				Params: []wall.ParsedFunParam{{
					Type: &wall.ParsedIdType{Token: wall.Token{Content: "A"}},
				}},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Content: "B"}},
			},
			&wall.ParsedExternFunDef{
				Extern: wall.Token{},
				Fun:    wall.Token{},
				Name:   wall.Token{Content: "b"},
				Params: []wall.ParsedFunParam{{
					Type: &wall.ParsedIdType{Token: wall.Token{Content: "A"}},
				}},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Content: "B"}},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, checkedFile.GlobalScope.DefineType(&wall.Token{Content: "A"}, wall.NewStructType()))
	typeIdA := checkedFile.GlobalScope.Types["A"].TypeId
	assert.NoError(t, checkedFile.GlobalScope.DefineType(&wall.Token{Content: "B"}, wall.NewStructType()))
	typeIdB := checkedFile.GlobalScope.Types["B"].TypeId
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	if fun, ok := checkedFile.GlobalScope.Funs["a"]; assert.Equal(t, ok, true) {
		assert.Equal(t, (*checkedFile.Types)[fun.TypeId], &wall.FunctionType{
			Params: []wall.TypeId{
				typeIdA,
			},
			Returns: typeIdB,
		})
	}
	if fun, ok := checkedFile.GlobalScope.Funs["b"]; assert.Equal(t, ok, true) {
		assert.Equal(t, (*checkedFile.Types)[fun.TypeId], &wall.FunctionType{
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
				Name: wall.Token{Content: "Employee"},
				Fields: []wall.ParsedStructField{
					{
						Name: wall.Token{Content: "name"},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: "String"},
						},
					},
					{
						Name: wall.Token{Content: "age"},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: "int"},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, checkedFile.GlobalScope.DefineType(&wall.Token{Content: "String"}, wall.NewStructType()))
	stringTypeId := checkedFile.GlobalScope.Types["String"].TypeId
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckTypeContents(file, checkedFile))
	employeeTypeId := checkedFile.GlobalScope.Types["Employee"].TypeId
	if assert.IsType(t, wall.NewStructType(), (*checkedFile.Types)[employeeTypeId]) {
		emp := (*checkedFile.Types)[employeeTypeId].(*wall.StructType)
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
			Token: wall.Token{Content: "()"},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedReturn{},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: "()"},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedReturn{
					Arg: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER, Content: "10"},
					},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: "int32"},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedReturn{
					Arg: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.TRUE},
					},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: "bool"},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedReturn{
					Arg: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.FALSE},
					},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: "bool"},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedBlock{},
				&wall.ParsedReturn{
					Arg: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER, Content: "10"},
					},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: "int32"},
		},
	},
	{
		block: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedBlock{},
				&wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{&wall.ParsedReturn{
						Arg: &wall.ParsedLiteralExpr{
							Token: wall.Token{Kind: wall.INTEGER, Content: "10"},
						},
					}},
				},
			},
		},
		returnType: &wall.ParsedIdType{
			Token: wall.Token{Content: "int32"},
		},
	},
}

func TestCheckBlocks(t *testing.T) {
	for _, test := range checkBlocksTests {
		file := &wall.ParsedFile{
			Defs: []wall.ParsedDef{
				&wall.ParsedFunDef{
					Id:         wall.Token{Content: "a"},
					Body:       test.block,
					ReturnType: test.returnType,
				},
			},
		}
		checkedFile := wall.NewCheckedCompilationUnit("")
		assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
		assert.NoError(t, wall.CheckBlocks(file, checkedFile))
	}
}

var checkStmtTests = []wall.ParsedStmt{
	&wall.ParsedVar{
		Id: wall.Token{Content: ""},
		Value: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
		},
	},
	&wall.ParsedExprStmt{
		Expr: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
		},
	},
}

func TestCheckStmt(t *testing.T) {
	for _, test := range checkStmtTests {
		checkedFile := wall.NewCheckedCompilationUnit("")
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
			Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
		},
		typeid: wall.INT32_TYPE_ID,
	},
	{
		expr: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: "0.0"},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedGroupedExpr{
			Inner: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedUnaryExpr{
			Operator: wall.Token{Kind: wall.MINUS},
			Operand: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.INT32_TYPE_ID,
	},
	{
		expr: &wall.ParsedUnaryExpr{
			Operator: wall.Token{Kind: wall.MINUS},
			Operand: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.PLUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.INT32_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.MINUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.INT32_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.STAR},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.INT32_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.SLASH},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.INT32_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.EQEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.BANGEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.LT},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.LTEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.GT},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
			Op: wall.Token{Kind: wall.GTEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.PLUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.MINUS},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.STAR},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.SLASH},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.FLOAT64_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.EQEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.BANGEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.LT},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.LTEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.GT},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
	{
		expr: &wall.ParsedBinaryExpr{
			Left: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
			Op: wall.Token{Kind: wall.GTEQ},
			Right: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.FLOAT, Content: "0"},
			},
		},
		typeid: wall.BOOL_TYPE_ID,
	},
}

func TestCheckExpr(t *testing.T) {
	for _, test := range checkExprTests {
		checkedFile := wall.NewCheckedCompilationUnit("")
		expr, err := wall.CheckExpr(test.expr, checkedFile.GlobalScope)
		if assert.NoError(t, err) {
			assert.Equal(t, test.typeid, expr.TypeId())
		}
	}
}

func TestCheckStringLiteralExpr(t *testing.T) {
	checkedFile := wall.NewCheckedCompilationUnit("")
	expr, err := wall.CheckExpr(&wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.STRING, Content: "\"ABC\""},
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
				Id: wall.Token{Content: ""},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedVar{
							Id: wall.Token{Content: "a"},
							Value: &wall.ParsedLiteralExpr{
								Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
							},
						},
						&wall.ParsedReturn{
							Arg: &wall.ParsedIdExpr{
								Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
							},
						},
						&wall.ParsedVar{
							Id: wall.Token{Content: "b"},
							Type: &wall.ParsedIdType{
								Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int32"},
							},
						},
						&wall.ParsedReturn{
							Arg: &wall.ParsedIdExpr{
								Token: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
							},
						},
					},
				},
				ReturnType: &wall.ParsedIdType{
					Token: wall.Token{Content: "int32"},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}

func TestCheckWhileStmt(t *testing.T) {
	checkedFile := wall.NewCheckedCompilationUnit("")
	got, err := wall.CheckStmt(&wall.ParsedWhile{
		While: wall.Token{Kind: wall.WHILE},
		Condition: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.TRUE},
		},
		Body: &wall.ParsedBlock{
			Stmts: []wall.ParsedStmt{
				&wall.ParsedBreak{},
				&wall.ParsedContinue{},
			},
		},
	}, checkedFile.GlobalScope, &wall.MayReturn{
		Type: wall.UNIT_TYPE_ID,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.CheckedWhile{
			Cond: &wall.CheckedLiteralExpr{
				Literal: wall.Token{Kind: wall.TRUE},
				Type:    wall.BOOL_TYPE_ID,
			},
			Body: &wall.CheckedBlock{
				Stmts: []wall.CheckedStmt{
					&wall.CheckedBreak{},
					&wall.CheckedContinue{},
				},
			},
		}, got)
	}
}

func TestCheckCallExpr(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Id: wall.Token{Content: "sum"},
				Params: []wall.ParsedFunParam{
					{
						Id: wall.Token{Content: "a"},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: "int32"},
						},
					},
					{
						Id: wall.Token{Content: "b"},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: "int32"},
						},
					},
				},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Content: "int32"}},
				Body:       &wall.ParsedBlock{Stmts: []wall.ParsedStmt{&wall.ParsedReturn{Arg: &wall.ParsedLiteralExpr{Token: wall.Token{Kind: wall.INTEGER, Content: "10"}}}}},
			},
			&wall.ParsedFunDef{
				Params:     []wall.ParsedFunParam{},
				ReturnType: nil,
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedCallExpr{
								Callee: &wall.ParsedIdExpr{
									Token: wall.Token{Kind: wall.IDENTIFIER, Content: "sum"},
								},
								Args: []wall.ParsedExpr{
									&wall.ParsedLiteralExpr{
										Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
									},
									&wall.ParsedLiteralExpr{
										Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}

func TestCheckIdExpr(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Id:     wall.Token{Content: "a"},
				Params: []wall.ParsedFunParam{},
				Body:   &wall.ParsedBlock{},
			},
			&wall.ParsedFunDef{
				Id: wall.Token{Content: "b"},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedIdExpr{
								Token: wall.Token{Content: "a"},
							},
						},
					},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
	checkedFile.Funs[0].Name.Content = "c"
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
				Name: wall.Token{Content: "Point"},
				Fields: []wall.ParsedStructField{
					{
						Name: wall.Token{Content: "x"},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: "int32"},
						},
					},
					{
						Name: wall.Token{Content: "y"},
						Type: &wall.ParsedIdType{
							Token: wall.Token{Content: "float64"},
						},
					},
				},
			},
			&wall.ParsedFunDef{
				Id: wall.Token{Content: "a"},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedStructInitExpr{
								Name: &wall.ParsedIdType{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "Point"}},
								Fields: []wall.ParsedStructInitField{
									{
										Name: wall.Token{Kind: wall.IDENTIFIER, Content: "x"},
										Value: &wall.ParsedLiteralExpr{
											Token: wall.Token{Kind: wall.INTEGER},
										},
									},
									{
										Name: wall.Token{Kind: wall.IDENTIFIER, Content: "y"},
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
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckTypeContents(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}

func TestCheckStructAccessExpr(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedStructDef{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: "Point"},
				Fields: []wall.ParsedStructField{
					{
						Name: wall.Token{Kind: wall.IDENTIFIER, Content: "x"},
						Type: &wall.ParsedIdType{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int32"}},
					},
					{
						Name: wall.Token{Kind: wall.IDENTIFIER, Content: "y"},
						Type: &wall.ParsedIdType{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int32"}},
					},
				},
			},
			&wall.ParsedFunDef{
				Id:         wall.Token{Kind: wall.IDENTIFIER, Content: "main"},
				Params:     []wall.ParsedFunParam{},
				ReturnType: &wall.ParsedIdType{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int32"}},
				Body: &wall.ParsedBlock{
					Left: wall.Token{Kind: wall.LEFTBRACE},
					Stmts: []wall.ParsedStmt{
						&wall.ParsedVar{
							Id:    wall.Token{Kind: wall.IDENTIFIER, Content: "p"},
							Type:  &wall.ParsedIdType{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "Point"}},
							Value: nil,
						},
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedObjectAccessExpr{
								Object: &wall.ParsedIdExpr{
									Token: wall.Token{Kind: wall.IDENTIFIER, Content: "p"},
								},
								Dot:    wall.Token{Kind: wall.DOT},
								Member: wall.Token{Kind: wall.IDENTIFIER, Content: "x"},
							},
						},
						&wall.ParsedReturn{
							Arg: &wall.ParsedObjectAccessExpr{
								Object: &wall.ParsedIdExpr{
									Token: wall.Token{Kind: wall.IDENTIFIER, Content: "p"},
								},
								Dot:    wall.Token{Kind: wall.DOT},
								Member: wall.Token{Kind: wall.IDENTIFIER, Content: "y"},
							},
						},
					},
					Right: wall.Token{Kind: wall.RIGHTBRACE},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckFunctionSignatures(file, checkedFile))
	assert.NoError(t, wall.CheckTypeContents(file, checkedFile))
	assert.NoError(t, wall.CheckBlocks(file, checkedFile))
}

func TestCheckAddressOp(t *testing.T) {
	checkedFile := wall.NewCheckedCompilationUnit("")
	parsedExpr := &wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.AMP},
		Operand: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
		},
	}
	_, err := wall.CheckExpr(parsedExpr, checkedFile.GlobalScope)
	assert.Error(t, err)
	parsedExpr = &wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.AMP},
		Operand: &wall.ParsedIdExpr{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
		},
	}
	checkedFile.GlobalScope.DefineVar(&wall.Token{Kind: wall.IDENTIFIER, Content: "a"}, wall.INT_TYPE_ID)
	checkedExpr, err := wall.CheckExpr(parsedExpr, checkedFile.GlobalScope)
	if assert.NoError(t, err) {
		assert.Equal(t, checkedExpr.TypeId(), checkedFile.TypeId(&wall.PointerType{
			Type: wall.INT_TYPE_ID,
		}))
	}
}

func TestCheckDerefOp(t *testing.T) {
	checkedFile := wall.NewCheckedCompilationUnit("")
	parsedExpr := &wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.STAR},
		Operand: &wall.ParsedIdExpr{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
		},
	}
	checkedFile.GlobalScope.DefineVar(&wall.Token{Kind: wall.IDENTIFIER, Content: "a"}, checkedFile.TypeId(&wall.PointerType{
		Type: wall.INT_TYPE_ID,
	}))
	checkedExpr, err := wall.CheckExpr(parsedExpr, checkedFile.GlobalScope)
	if assert.NoError(t, err) {
		assert.Equal(t, checkedExpr.TypeId(), wall.INT_TYPE_ID)
	}
}

func TestCheckAssignExpr(t *testing.T) {
	checkedFile := wall.NewCheckedCompilationUnit("")
	parsedExpr := &wall.ParsedBinaryExpr{
		Left: &wall.ParsedIdExpr{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
		},
		Op: wall.Token{Kind: wall.EQ},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: "0"},
		},
	}
	checkedFile.GlobalScope.DefineVar(&wall.Token{Kind: wall.IDENTIFIER, Content: "a"}, wall.INT32_TYPE_ID)
	checkedExpr, err := wall.CheckExpr(parsedExpr, checkedFile.GlobalScope)
	if assert.NoError(t, err) {
		assert.Equal(t, checkedExpr.TypeId(), wall.INT32_TYPE_ID)
	}
}

func TestCheckModuleAccessExpr(t *testing.T) {
	fileA := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
				Name: wall.Token{Content: "B"},
				File: nil,
			},
			&wall.ParsedFunDef{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedExprStmt{
							Expr: &wall.ParsedCallExpr{
								Callee: &wall.ParsedModuleAccessExpr{
									Module: wall.Token{Kind: wall.IDENTIFIER, Content: "B"},
									Member: &wall.ParsedIdExpr{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "b"}},
								},
								Args: []wall.ParsedExpr{},
							},
						},
					},
				},
			},
		},
	}
	fileB := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedFunDef{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{},
				},
			},
		},
	}
	fileA.Defs[0].(*wall.ParsedImport).File = fileB
	checkedFileA := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckImports(fileA, checkedFileA))
	assert.NoError(t, wall.CheckFunctionSignatures(fileA, checkedFileA))
	assert.NoError(t, wall.CheckBlocks(fileA, checkedFileA))
}

func TestCheckModuleAccessType(t *testing.T) {
	fileA := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
				Name: wall.Token{Content: "B"},
				File: nil,
			},
			&wall.ParsedFunDef{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				Body: &wall.ParsedBlock{
					Stmts: []wall.ParsedStmt{
						&wall.ParsedVar{
							Id: wall.Token{Kind: wall.IDENTIFIER, Content: "v"},
							Type: &wall.ParsedIdType{
								Token: wall.Token{Kind: wall.IDENTIFIER, Content: "A"},
							},
						},
					},
				},
			},
		},
	}
	fileB := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedStructDef{
				Name:   wall.Token{Kind: wall.IDENTIFIER, Content: "A"},
				Fields: []wall.ParsedStructField{},
			},
		},
	}
	fileA.Defs[0].(*wall.ParsedImport).File = fileB
	checkedFileA := wall.NewCheckedCompilationUnit("")
	assert.NoError(t, wall.CheckImports(fileA, checkedFileA))
	assert.NoError(t, wall.CheckTypeContents(fileA, checkedFileA))
	assert.NoError(t, wall.CheckBlocks(fileA, checkedFileA))
}

func TestCheckAsExpr(t *testing.T) {
	checkedFile := wall.NewCheckedCompilationUnit("")
	got, err := wall.CheckExpr(&wall.ParsedAsExpr{
		Value: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER},
		},
		As: wall.Token{Kind: wall.AS},
		Type: &wall.ParsedIdType{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: "float64"},
		},
	}, checkedFile.GlobalScope)
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.CheckedAsExpr{
			Value: &wall.CheckedLiteralExpr{
				Literal: wall.Token{Kind: wall.INTEGER},
				Type:    wall.INT32_TYPE_ID,
			},
			Type: wall.FLOAT64_TYPE_ID,
		}, got)
	}
}

func TestCheckTypealiasDef(t *testing.T) {
	file := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedTypealiasDef{
				Typealias: wall.Token{Kind: wall.TYPEALIAS},
				Name:      wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
		},
	}
	checkedFile := wall.NewCheckedCompilationUnit("")
	if assert.NoError(t, wall.CheckTypeSignatures(file, checkedFile)) && assert.NoError(t, wall.CheckTypeContents(file, checkedFile)) {
		assert.Equal(t, wall.INT_TYPE_ID, checkedFile.GlobalScope.Types["a"].TypeId)
	}
}
