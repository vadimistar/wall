package wall_test

import (
	"os"
	"reflect"
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
)

type parseLiteralExprTest struct {
	tokens   []wall.Token
	expected wall.ParsedExpr
}

var parseLiteralExprTests = []parseLiteralExprTest{
	{[]wall.Token{
		{Kind: wall.IDENTIFIER, Content: "abc"},
		{Kind: wall.EOF},
	}, &wall.ParsedIdExpr{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: "abc"},
	}},
	{[]wall.Token{
		{Kind: wall.INTEGER, Content: "123"},
		{Kind: wall.EOF},
	}, &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.INTEGER, Content: "123"},
	}},
	{[]wall.Token{
		{Kind: wall.FLOAT, Content: "1.0"},
		{Kind: wall.EOF},
	}, &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.FLOAT, Content: "1.0"},
	}},
	{[]wall.Token{
		{Kind: wall.STRING, Content: "ABC"},
		{Kind: wall.EOF},
	}, &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.STRING, Content: "ABC"},
	}},
}

func TestParseLiteralAndIdExpr(t *testing.T) {
	for _, test := range parseLiteralExprTests {
		t.Logf("testing %#v", test.tokens)
		p := wall.NewParser(test.tokens)
		e, err := p.ParseExprAndEof()
		assert.NoError(t, err)
		assert.Equal(t, e, test.expected)
	}
}

var unaryOps = []wall.Token{
	{Kind: wall.MINUS},
	{Kind: wall.AMP},
	{Kind: wall.STAR},
}

func TestParseUnaryExpr(t *testing.T) {
	for _, op := range unaryOps {
		t.Logf("testing %+v", op)
		pr := wall.NewParser([]wall.Token{op, {Kind: wall.IDENTIFIER}, {Kind: wall.EOF}})
		expr, err := pr.ParseExprAndEof()
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(expr), reflect.TypeOf(&wall.ParsedUnaryExpr{}))
	}
}

var binaryOps = []wall.Token{
	{Kind: wall.PLUS},
	{Kind: wall.MINUS},
	{Kind: wall.STAR},
	{Kind: wall.SLASH},
	{Kind: wall.EQEQ},
	{Kind: wall.BANGEQ},
	{Kind: wall.LT},
	{Kind: wall.LTEQ},
	{Kind: wall.GT},
	{Kind: wall.GTEQ},
	{Kind: wall.EQ},
}

func TestParseBinaryExpr(t *testing.T) {
	for _, op := range binaryOps {
		t.Logf("testing %s", op.Kind)
		pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER, Content: "a"}, op, {Kind: wall.IDENTIFIER, Content: "b"}, op, {Kind: wall.IDENTIFIER, Content: "c"}, {Kind: wall.EOF}})
		res, err := pr.ParseExprAndEof()
		assert.NoError(t, err)
		if wall.IsRightAssoc(op.Kind) {
			expected := &wall.ParsedBinaryExpr{
				Left: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				},
				Op: op,
				Right: &wall.ParsedBinaryExpr{
					Left: &wall.ParsedIdExpr{
						Token: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
					},
					Op: op,
					Right: &wall.ParsedIdExpr{
						Token: wall.Token{Kind: wall.IDENTIFIER, Content: "c"},
					},
				},
			}
			assert.Equal(t, res, expected)
			return
		}
		expected := &wall.ParsedBinaryExpr{
			Left: &wall.ParsedBinaryExpr{
				Left: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				},
				Op: op,
				Right: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
				},
			},
			Op: op,
			Right: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER, Content: "c"},
			},
		}
		assert.Equal(t, res, expected)
	}
}

func TestParseGroupedExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.LEFTPAREN}, {Kind: wall.IDENTIFIER}, {Kind: wall.RIGHTPAREN}})
	expr, err := pr.ParseExprAndEof()
	assert.NoError(t, err)
	assert.Equal(t, expr, &wall.ParsedGroupedExpr{
		Left: wall.Token{Kind: wall.LEFTPAREN},
		Inner: &wall.ParsedIdExpr{
			Token: wall.Token{Kind: wall.IDENTIFIER},
		},
		Right: wall.Token{Kind: wall.RIGHTPAREN},
	})
}

type parseCallExprTest struct {
	tokens []wall.Token
	node   *wall.ParsedCallExpr
}

var parseCallExprTests = []parseCallExprTest{
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.RIGHTPAREN}},
		node: &wall.ParsedCallExpr{
			Callee: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ParsedExpr{},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTPAREN}},
		node: &wall.ParsedCallExpr{
			Callee: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ParsedExpr{
				&wall.ParsedLiteralExpr{
					Token: wall.Token{Kind: wall.INTEGER},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.INTEGER}, {Kind: wall.COMMA}, {Kind: wall.RIGHTPAREN}},
		node: &wall.ParsedCallExpr{
			Callee: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ParsedExpr{
				&wall.ParsedLiteralExpr{
					Token: wall.Token{Kind: wall.INTEGER},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.INTEGER}, {Kind: wall.COMMA}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTPAREN}},
		node: &wall.ParsedCallExpr{
			Callee: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ParsedExpr{
				&wall.ParsedLiteralExpr{
					Token: wall.Token{Kind: wall.INTEGER},
				},
				&wall.ParsedLiteralExpr{
					Token: wall.Token{Kind: wall.INTEGER},
				},
			},
		},
	},
}

func TestParseCallExpr(t *testing.T) {
	for _, test := range parseCallExprTests {
		pr := wall.NewParser(test.tokens)
		expr, err := pr.ParseExprAndEof()
		if assert.NoError(t, err) {
			assert.Equal(t, test.node, expr)
		}
	}
}

type parseStructInitExprTest struct {
	tokens   []wall.Token
	expected *wall.ParsedStructInitExpr
}

var parseStructInitExprTests = []parseStructInitExprTest{
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: "Bob"}, {Kind: wall.LEFTBRACE}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name:   &wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: "Bob"}, {Kind: wall.LEFTBRACE}, {Kind: wall.IDENTIFIER, Content: "age"}, {Kind: wall.COLON}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name: &wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{
				{
					Name: wall.Token{Kind: wall.IDENTIFIER, Content: "age"},
					Value: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER},
					},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: "Bob"}, {Kind: wall.LEFTBRACE}, {Kind: wall.IDENTIFIER, Content: "age"}, {Kind: wall.COLON}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name: &wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{
				{
					Name: wall.Token{Kind: wall.IDENTIFIER, Content: "age"},
					Value: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER},
					},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: "Bob"}, {Kind: wall.LEFTBRACE}, {Kind: wall.NEWLINE}, {Kind: wall.IDENTIFIER, Content: "age"}, {Kind: wall.COLON}, {Kind: wall.INTEGER}, {Kind: wall.COMMA}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name: &wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{
				{
					Name: wall.Token{Kind: wall.IDENTIFIER, Content: "age"},
					Value: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER},
					},
				},
			},
		},
	},
}

func TestParseStructInitExpr(t *testing.T) {
	for _, test := range parseStructInitExprTests {
		pr := wall.NewParser(test.tokens)
		expr, err := pr.ParseExprAndEof()
		if assert.NoError(t, err) {
			assert.IsType(t, test.expected, expr)
		}
	}
}

func TestParseVarStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.COLONEQ}, {Kind: wall.INTEGER}})
	stmt, err := pr.ParseStmtAndEof()
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(stmt), reflect.TypeOf(&wall.ParsedVar{}))
}

func TestParseMutVarStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.MUT}, {Kind: wall.IDENTIFIER}, {Kind: wall.COLONEQ}, {Kind: wall.INTEGER}})
	stmt, err := pr.ParseStmtAndEof()
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(stmt), reflect.TypeOf(&wall.ParsedVar{}))
}

func TestParseReturnStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.RETURN}, {Kind: wall.IDENTIFIER}})
	stmt, err := pr.ParseStmtAndEof()
	if assert.NoError(t, err) {
		assert.IsType(t, stmt, &wall.ParsedReturn{
			Arg: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
		})
	}
	pr = wall.NewParser([]wall.Token{{Kind: wall.RETURN}})
	stmt, err = pr.ParseStmtAndEof()
	if assert.NoError(t, err) {
		assert.IsType(t, stmt, &wall.ParsedReturn{})
	}
}

func TestParseBlockStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.LEFTBRACE}, {Kind: wall.NEWLINE}, {Kind: wall.IDENTIFIER}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}})
	got, err := pr.ParseStmtAndEof()
	assert.NoError(t, err)
	expected := &wall.ParsedBlock{
		Left: wall.Token{Kind: wall.LEFTBRACE},
		Stmts: []wall.ParsedStmt{
			&wall.ParsedExprStmt{
				Expr: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER},
				},
			},
		},
		Right: wall.Token{Kind: wall.RIGHTBRACE},
	}
	assert.Equal(t, got, expected)
}

type parseIfStmtTest struct {
	tokens   []wall.Token
	expected *wall.ParsedIf
}

var parseIfStmtTests = []parseIfStmtTest{
	{
		tokens: []wall.Token{
			{Kind: wall.IF}, {Kind: wall.IDENTIFIER}, {Kind: wall.LEFTBRACE}, {Kind: wall.IDENTIFIER}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE},
		},
		expected: &wall.ParsedIf{
			If: wall.Token{Kind: wall.IF},
			Condition: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Body: &wall.ParsedBlock{
				Left:  wall.Token{Kind: wall.LEFTBRACE},
				Stmts: []wall.ParsedStmt{&wall.ParsedExprStmt{Expr: &wall.ParsedIdExpr{Token: wall.Token{Kind: wall.IDENTIFIER}}}},
				Right: wall.Token{Kind: wall.RIGHTBRACE},
			},
			ElseBody: nil,
		},
	},
	{
		tokens: []wall.Token{
			{Kind: wall.IF}, {Kind: wall.IDENTIFIER}, {Kind: wall.LEFTBRACE}, {Kind: wall.NEWLINE}, {Kind: wall.IDENTIFIER}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}, {Kind: wall.ELSE}, {Kind: wall.LEFTBRACE}, {Kind: wall.IDENTIFIER}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE},
		},
		expected: &wall.ParsedIf{
			If: wall.Token{Kind: wall.IF},
			Condition: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Body: &wall.ParsedBlock{
				Left:  wall.Token{Kind: wall.LEFTBRACE},
				Stmts: []wall.ParsedStmt{&wall.ParsedExprStmt{Expr: &wall.ParsedIdExpr{Token: wall.Token{Kind: wall.IDENTIFIER}}}},
				Right: wall.Token{Kind: wall.RIGHTBRACE},
			},
			ElseBody: &wall.ParsedBlock{
				Left: wall.Token{Kind: wall.LEFTBRACE},
				Stmts: []wall.ParsedStmt{
					&wall.ParsedExprStmt{
						Expr: &wall.ParsedIdExpr{
							Token: wall.Token{Kind: wall.IDENTIFIER},
						},
					},
				},
				Right: wall.Token{Kind: wall.RIGHTBRACE},
			},
		},
	},
}

func TestParseWhileStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.WHILE}, {Kind: wall.TRUE}, {Kind: wall.LEFTBRACE}, {Kind: wall.BREAK}, {Kind: wall.NEWLINE}, {Kind: wall.CONTINUE}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}})
	got, err := pr.ParseStmtAndEof()
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.ParsedWhile{
			While: wall.Token{Kind: wall.WHILE},
			Condition: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.TRUE},
			},
			Body: &wall.ParsedBlock{
				Left: wall.Token{Kind: wall.LEFTBRACE},
				Stmts: []wall.ParsedStmt{
					&wall.ParsedBreak{
						Break: wall.Token{Kind: wall.BREAK},
					},
					&wall.ParsedContinue{
						Continue: wall.Token{Kind: wall.CONTINUE},
					},
				},
				Right: wall.Token{Kind: wall.RIGHTBRACE},
			},
		}, got)
	}
}

func TestParseIfStmt(t *testing.T) {
	for _, test := range parseIfStmtTests {
		pr := wall.NewParser(test.tokens)
		stmt, err := pr.ParseStmtAndEof()
		if assert.NoError(t, err) {
			assert.Equal(t, test.expected, stmt)
		}
	}
}

type parseFunDefTest struct {
	tokens   []wall.Token
	expected wall.ParsedDef
}

var parseFunDefTests = []parseFunDefTest{
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: "sum"},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.IDENTIFIER, Content: "a"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: "b"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.ParsedFunDef{
		Fun: wall.Token{Kind: wall.FUN},
		Id:  wall.Token{Kind: wall.IDENTIFIER, Content: "sum"},
		Params: []wall.ParsedFunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
		},
		ReturnType: &wall.ParsedIdType{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
		},
		Body: &wall.ParsedBlock{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.ParsedStmt{},
			Right: wall.Token{Kind: wall.RIGHTBRACE},
		},
	}},
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: "sum"},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.IDENTIFIER, Content: "a"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: "b"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.ParsedFunDef{
		Fun: wall.Token{Kind: wall.FUN},
		Id:  wall.Token{Kind: wall.IDENTIFIER, Content: "sum"},
		Params: []wall.ParsedFunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
		},
		ReturnType: nil,
		Body: &wall.ParsedBlock{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.ParsedStmt{},
			Right: wall.Token{Kind: wall.RIGHTBRACE},
		},
	}},
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: "main"},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.ParsedFunDef{
		Fun:        wall.Token{Kind: wall.FUN},
		Id:         wall.Token{Kind: wall.IDENTIFIER, Content: "main"},
		Params:     []wall.ParsedFunParam{},
		ReturnType: nil,
		Body: &wall.ParsedBlock{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.ParsedStmt{},
			Right: wall.Token{Kind: wall.RIGHTBRACE},
		},
	}},
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: "A"},
		{Kind: wall.DOT},
		{Kind: wall.IDENTIFIER, Content: "main"},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.ParsedFunDef{
		Fun:        wall.Token{Kind: wall.FUN},
		Typename:   &wall.Token{Kind: wall.IDENTIFIER, Content: "A"},
		Dot:        &wall.Token{Kind: wall.DOT},
		Id:         wall.Token{Kind: wall.IDENTIFIER, Content: "main"},
		Params:     []wall.ParsedFunParam{},
		ReturnType: nil,
		Body: &wall.ParsedBlock{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.ParsedStmt{},
			Right: wall.Token{Kind: wall.RIGHTBRACE},
		},
	}},
}

func TestParseFunDef(t *testing.T) {
	for _, test := range parseFunDefTests {
		pr := wall.NewParser(test.tokens)
		got, err := pr.ParseDefAndEof()
		assert.NoError(t, err)
		assert.Equal(t, got, test.expected)
	}
}

type parseExternFunDefTest struct {
	tokens   []wall.Token
	expected *wall.ParsedExternFunDef
}

var parseExternFunDefTests = []parseExternFunDefTest{
	{[]wall.Token{
		{Kind: wall.EXTERN},
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: "sum"},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.IDENTIFIER, Content: "a"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: "b"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.EOF},
	}, &wall.ParsedExternFunDef{
		Extern: wall.Token{Kind: wall.EXTERN},
		Fun:    wall.Token{Kind: wall.FUN},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: "sum"},
		Params: []wall.ParsedFunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
		},
		ReturnType: &wall.ParsedIdType{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
		},
	}},
}

func TestParseExternFunDef(t *testing.T) {
	for _, test := range parseExternFunDefTests {
		pr := wall.NewParser(test.tokens)
		got, err := pr.ParseDefAndEof()
		if assert.NoError(t, err) {
			assert.Equal(t, got, test.expected)
		}
	}
}

func TestParseImportDef(t *testing.T) {
	pr := wall.NewParser([]wall.Token{
		{Kind: wall.IMPORT},
		{Kind: wall.IDENTIFIER, Content: "a"},
	})
	got, err := pr.ParseDefAndEof()
	assert.NoError(t, err)
	expected := &wall.ParsedImport{
		Import: wall.Token{Kind: wall.IMPORT},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
	}
	assert.Equal(t, got, expected)
}

func TestParseStructDef(t *testing.T) {
	pr := wall.NewParser([]wall.Token{
		{Kind: wall.STRUCT},
		{Kind: wall.IDENTIFIER, Content: "Employee"},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.IDENTIFIER, Content: "id"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: "age"},
		{Kind: wall.IDENTIFIER, Content: "int"},
		{Kind: wall.NEWLINE},
		{Kind: wall.RIGHTBRACE},
	})
	got, err := pr.ParseDefAndEof()
	assert.NoError(t, err)
	expected := &wall.ParsedStructDef{
		Struct: wall.Token{Kind: wall.STRUCT},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: "Employee"},
		Fields: []wall.ParsedStructField{
			{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: "id"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
			{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: "age"},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: "int"},
				},
			},
		},
	}
	assert.Equal(t, got, expected)
}

func TestParseFile(t *testing.T) {
	tokens := []wall.Token{
		{Kind: wall.IMPORT},
		{Kind: wall.IDENTIFIER, Content: "a"},
		{Kind: wall.NEWLINE},
	}
	tokens = append(tokens, parseFunDefTests[0].tokens...)
	tokens = append(tokens, wall.Token{Kind: wall.NEWLINE})
	pr := wall.NewParser(tokens)
	got, err := pr.ParseFile()
	assert.NoError(t, err)
	expected := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
				Import: wall.Token{Kind: wall.IMPORT},
				Name:   wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
			},
			parseFunDefTests[0].expected,
		},
	}
	assert.Equal(t, got, expected)
}

func TestParseCompilationUnit(t *testing.T) {
	if err := os.WriteFile("A.wall", []byte("import B\nfun a() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("B.wall", []byte("import C\nfun b() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("C.wall", []byte("import A\nfun c() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}
	A, err := wall.ParseCompilationUnit("A.wall", "import B\nfun a() {}\n", "")
	if err != nil {
		t.Fatal(err)
	}
	_ = A.Defs[1].(*wall.ParsedFunDef)
	importB := A.Defs[0].(*wall.ParsedImport)
	B := importB.File
	_ = B.Defs[1].(*wall.ParsedFunDef)
	importC := B.Defs[0].(*wall.ParsedImport)
	C := importC.File
	_ = C.Defs[1].(*wall.ParsedFunDef)
	importA := C.Defs[0].(*wall.ParsedImport)
	assert.Equal(t, importA.File, A)
}

func TestParseObjectAccessExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER, Content: "a"}, {Kind: wall.DOT}, {Kind: wall.IDENTIFIER, Content: "b"}})
	got, err := pr.ParseExprAndEof()
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.ParsedObjectAccessExpr{
			Object: &wall.ParsedIdExpr{
				Token: wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
			},
			Dot:    wall.Token{Kind: wall.DOT},
			Member: wall.Token{Kind: wall.IDENTIFIER, Content: "b"},
		}, got)
	}
}

func TestParseModuleAccessExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER, Content: "a"}, {Kind: wall.COLONCOLON}, {Kind: wall.IDENTIFIER, Content: "b"}})
	got, err := pr.ParseExprAndEof()
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.ParsedModuleAccessExpr{
			Module:     wall.Token{Kind: wall.IDENTIFIER, Content: "a"},
			Coloncolon: wall.Token{Kind: wall.COLONCOLON},
			Member:     &wall.ParsedIdExpr{Token: wall.Token{Kind: wall.IDENTIFIER, Content: "b"}},
		}, got)
	}
}

func TestParseAsExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.INTEGER}, {Kind: wall.AS}, {Kind: wall.IDENTIFIER}})
	got, err := pr.ParseExprAndEof()
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.ParsedAsExpr{
			Value: &wall.ParsedLiteralExpr{
				Token: wall.Token{Kind: wall.INTEGER},
			},
			As: wall.Token{Kind: wall.AS},
			Type: &wall.ParsedIdType{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
		}, got)
	}
}

func TestParseTypealiasDef(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.TYPEALIAS}, {Kind: wall.IDENTIFIER}, {Kind: wall.IDENTIFIER}})
	got, err := pr.ParseDef()
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.ParsedTypealiasDef{
			Typealias: wall.Token{Kind: wall.TYPEALIAS},
			Name:      wall.Token{Kind: wall.IDENTIFIER},
			Type: &wall.ParsedIdType{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
		}, got)
	}
}

func TestParseThisExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.DOT}, {Kind: wall.IDENTIFIER}})
	got, err := pr.ParseExprAndEof()
	if assert.NoError(t, err) {
		assert.Equal(t, &wall.ParsedObjectAccessExpr{
			Object: nil,
			Dot:    wall.Token{Kind: wall.DOT},
			Member: wall.Token{Kind: wall.IDENTIFIER},
		}, got)
	}
}
