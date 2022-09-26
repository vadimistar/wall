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
	expected wall.LiteralExprNode
}

var parseLiteralExprTests = []parseLiteralExprTest{
	{[]wall.Token{
		{Kind: wall.IDENTIFIER, Content: []byte("abc")},
		{Kind: wall.EOF},
	}, wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("abc")},
	}},
	{[]wall.Token{
		{Kind: wall.INTEGER, Content: []byte("123")},
		{Kind: wall.EOF},
	}, wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
	}},
	{[]wall.Token{
		{Kind: wall.FLOAT, Content: []byte("1.0")},
		{Kind: wall.EOF},
	}, wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
	}},
}

func TestParseLiteralExpr(t *testing.T) {
	for _, test := range parseLiteralExprTests {
		t.Logf("testing %#v", test.tokens)
		p := wall.NewParser(test.tokens)
		e, err := p.ParseExprAndEof()
		assert.NoError(t, err)
		assert.Equal(t, e, &test.expected)
	}
}

var unaryOps = []wall.Token{
	{Kind: wall.PLUS},
	{Kind: wall.MINUS},
}

func TestParseUnaryExpr(t *testing.T) {
	for _, op := range unaryOps {
		t.Logf("testing %+v", op)
		pr := wall.NewParser([]wall.Token{op, {Kind: wall.IDENTIFIER}, {Kind: wall.EOF}})
		expr, err := pr.ParseExprAndEof()
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(expr), reflect.TypeOf(&wall.UnaryExprNode{}))
	}
}

var binaryOps = []wall.Token{
	{Kind: wall.PLUS},
	{Kind: wall.MINUS},
	{Kind: wall.EQ},
}

func TestParseBinaryExpr(t *testing.T) {
	for _, op := range binaryOps {
		t.Logf("testing %s", op.Kind)
		pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER, Content: []byte("a")}, op, {Kind: wall.IDENTIFIER, Content: []byte("b")}, op, {Kind: wall.IDENTIFIER, Content: []byte("c")}, {Kind: wall.EOF}})
		res, err := pr.ParseExprAndEof()
		assert.NoError(t, err)
		if wall.IsRightAssoc(op.Kind) {
			expected := &wall.BinaryExprNode{
				Left: &wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				},
				Op: op,
				Right: &wall.BinaryExprNode{
					Left: &wall.LiteralExprNode{
						Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
					},
					Op: op,
					Right: &wall.LiteralExprNode{
						Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("c")},
					},
				},
			}
			assert.Equal(t, res, expected)
			return
		}
		expected := &wall.BinaryExprNode{
			Left: &wall.BinaryExprNode{
				Left: &wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				},
				Op: op,
				Right: &wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				},
			},
			Op: op,
			Right: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("c")},
			},
		}
		assert.Equal(t, res, expected)
	}
}

func TestParseGroupedExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.LEFTPAREN}, {Kind: wall.IDENTIFIER}, {Kind: wall.RIGHTPAREN}})
	expr, err := pr.ParseExprAndEof()
	assert.NoError(t, err)
	assert.Equal(t, expr, &wall.GroupedExprNode{
		Left: wall.Token{Kind: wall.LEFTPAREN},
		Inner: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.IDENTIFIER},
		},
		Right: wall.Token{Kind: wall.RIGHTPAREN},
	})
}

type parseCallExprTest struct {
	tokens []wall.Token
	node   *wall.CallExprNode
}

var parseCallExprTests = []parseCallExprTest{
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.RIGHTPAREN}},
		node: &wall.CallExprNode{
			Callee: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ExprNode{},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTPAREN}},
		node: &wall.CallExprNode{
			Callee: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ExprNode{
				&wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.INTEGER},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.INTEGER}, {Kind: wall.COMMA}, {Kind: wall.RIGHTPAREN}},
		node: &wall.CallExprNode{
			Callee: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ExprNode{
				&wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.INTEGER},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER}, {Kind: wall.LEFTPAREN}, {Kind: wall.INTEGER}, {Kind: wall.COMMA}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTPAREN}},
		node: &wall.CallExprNode{
			Callee: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
			Args: []wall.ExprNode{
				&wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.INTEGER},
				},
				&wall.LiteralExprNode{
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
			assert.Equal(t, expr, test.node)
		}
	}
}

func TestParseVarStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.VAR}, {Kind: wall.IDENTIFIER}, {Kind: wall.EQ}, {Kind: wall.INTEGER}})
	stmt, err := pr.ParseStmtAndEof()
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(stmt), reflect.TypeOf(&wall.VarStmt{}))
}

func TestParseReturnStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.RETURN}, {Kind: wall.IDENTIFIER}})
	stmt, err := pr.ParseStmtAndEof()
	if assert.NoError(t, err) {
		assert.IsType(t, stmt, &wall.ReturnStmt{
			Arg: &wall.LiteralExprNode{
				Token: wall.Token{Kind: wall.IDENTIFIER},
			},
		})
	}
	pr = wall.NewParser([]wall.Token{{Kind: wall.RETURN}})
	stmt, err = pr.ParseStmtAndEof()
	if assert.NoError(t, err) {
		assert.IsType(t, stmt, &wall.ReturnStmt{})
	}
}

func TestParseBlockStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.LEFTBRACE}, {Kind: wall.NEWLINE}, {Kind: wall.IDENTIFIER}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}})
	got, err := pr.ParseStmtAndEof()
	assert.NoError(t, err)
	expected := &wall.BlockStmt{
		Left: wall.Token{Kind: wall.LEFTBRACE},
		Stmts: []wall.StmtNode{
			&wall.ExprStmt{
				Expr: &wall.LiteralExprNode{
					Token: wall.Token{Kind: wall.IDENTIFIER},
				},
			},
		},
		Right: wall.Token{Kind: wall.RIGHTBRACE},
	}
	assert.Equal(t, got, expected)
}

type parseFunDefTest struct {
	tokens   []wall.Token
	expected wall.DefNode
}

var parseFunDefTests = []parseFunDefTest{
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: []byte("b")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.FunDef{
		Fun: wall.Token{Kind: wall.FUN},
		Id:  wall.Token{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		Params: []wall.FunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				Type: &wall.IdTypeNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				Type: &wall.IdTypeNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
		},
		ReturnType: &wall.IdTypeNode{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
		},
		Body: &wall.BlockStmt{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.StmtNode{},
			Right: wall.Token{Kind: wall.RIGHTBRACE},
		},
	}},
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: []byte("b")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.FunDef{
		Fun: wall.Token{Kind: wall.FUN},
		Id:  wall.Token{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		Params: []wall.FunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				Type: &wall.IdTypeNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				Type: &wall.IdTypeNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
		},
		ReturnType: nil,
		Body: &wall.BlockStmt{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.StmtNode{},
			Right: wall.Token{Kind: wall.RIGHTBRACE},
		},
	}},
	{[]wall.Token{
		{Kind: wall.FUN},
		{Kind: wall.IDENTIFIER, Content: []byte("main")},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.FunDef{
		Fun:        wall.Token{Kind: wall.FUN},
		Id:         wall.Token{Kind: wall.IDENTIFIER, Content: []byte("main")},
		Params:     []wall.FunParam{},
		ReturnType: nil,
		Body: &wall.BlockStmt{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.StmtNode{},
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

func TestParseImportDef(t *testing.T) {
	pr := wall.NewParser([]wall.Token{
		{Kind: wall.IMPORT},
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
	})
	got, err := pr.ParseDefAndEof()
	assert.NoError(t, err)
	expected := &wall.ImportDef{
		Import: wall.Token{Kind: wall.IMPORT},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
	}
	assert.Equal(t, got, expected)
}

func TestParseStructDef(t *testing.T) {
	pr := wall.NewParser([]wall.Token{
		{Kind: wall.STRUCT},
		{Kind: wall.IDENTIFIER, Content: []byte("Employee")},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.IDENTIFIER, Content: []byte("id")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: []byte("age")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.NEWLINE},
		{Kind: wall.RIGHTBRACE},
	})
	got, err := pr.ParseDefAndEof()
	assert.NoError(t, err)
	expected := &wall.StructDef{
		Struct: wall.Token{Kind: wall.STRUCT},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("Employee")},
		Fields: []wall.StructField{
			{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("id")},
				Type: &wall.IdTypeNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("age")},
				Type: &wall.IdTypeNode{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
		},
	}
	assert.Equal(t, got, expected)
}

func TestParseFile(t *testing.T) {
	tokens := []wall.Token{
		{Kind: wall.IMPORT},
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
		{Kind: wall.NEWLINE},
	}
	tokens = append(tokens, parseFunDefTests[0].tokens...)
	tokens = append(tokens, wall.Token{Kind: wall.NEWLINE})
	pr := wall.NewParser(tokens)
	got, err := pr.ParseFile()
	assert.NoError(t, err)
	expected := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.ImportDef{
				Import: wall.Token{Kind: wall.IMPORT},
				Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
			},
			parseFunDefTests[0].expected,
		},
	}
	assert.Equal(t, got, expected)
}

func TestParseCompilationUnit(t *testing.T) {
	if err := os.WriteFile("A.wl", []byte("import B\nfun a() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("B.wl", []byte("import C\nfun b() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("C.wl", []byte("import A\nfun c() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}
	A, err := wall.ParseCompilationUnit("A.wl", []byte("import B\nfun a() {}\n"))
	if err != nil {
		t.Fatal(err)
	}
	_ = A.Defs[1].(*wall.FunDef)
	importB := A.Defs[0].(*wall.ParsedImportDef)
	B := importB.ParsedNode
	_ = B.Defs[1].(*wall.FunDef)
	importC := B.Defs[0].(*wall.ParsedImportDef)
	C := importC.ParsedNode
	_ = C.Defs[1].(*wall.FunDef)
	importA := C.Defs[0].(*wall.ParsedImportDef)
	assert.Equal(t, importA.ParsedNode, A)
}
