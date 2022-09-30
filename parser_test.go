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
		{Kind: wall.IDENTIFIER, Content: []byte("abc")},
		{Kind: wall.EOF},
	}, &wall.ParsedIdExpr{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("abc")},
	}},
	{[]wall.Token{
		{Kind: wall.INTEGER, Content: []byte("123")},
		{Kind: wall.EOF},
	}, &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
	}},
	{[]wall.Token{
		{Kind: wall.FLOAT, Content: []byte("1.0")},
		{Kind: wall.EOF},
	}, &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
	}},
	{[]wall.Token{
		{Kind: wall.STRING, Content: []byte("ABC")},
		{Kind: wall.EOF},
	}, &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.STRING, Content: []byte("ABC")},
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
	{Kind: wall.PLUS},
	{Kind: wall.MINUS},
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
	{Kind: wall.EQ},
}

func TestParseBinaryExpr(t *testing.T) {
	for _, op := range binaryOps {
		t.Logf("testing %s", op.Kind)
		pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER, Content: []byte("a")}, op, {Kind: wall.IDENTIFIER, Content: []byte("b")}, op, {Kind: wall.IDENTIFIER, Content: []byte("c")}, {Kind: wall.EOF}})
		res, err := pr.ParseExprAndEof()
		assert.NoError(t, err)
		if wall.IsRightAssoc(op.Kind) {
			expected := &wall.ParsedBinaryExpr{
				Left: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				},
				Op: op,
				Right: &wall.ParsedBinaryExpr{
					Left: &wall.ParsedIdExpr{
						Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
					},
					Op: op,
					Right: &wall.ParsedIdExpr{
						Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("c")},
					},
				},
			}
			assert.Equal(t, res, expected)
			return
		}
		expected := &wall.ParsedBinaryExpr{
			Left: &wall.ParsedBinaryExpr{
				Left: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				},
				Op: op,
				Right: &wall.ParsedIdExpr{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				},
			},
			Op: op,
			Right: &wall.ParsedIdExpr{
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
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: []byte("Bob")}, {Kind: wall.LEFTBRACE}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name:   wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: []byte("Bob")}, {Kind: wall.LEFTBRACE}, {Kind: wall.IDENTIFIER, Content: []byte("age")}, {Kind: wall.COLON}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name: wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{
				{
					Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("age")},
					Value: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER},
					},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: []byte("Bob")}, {Kind: wall.LEFTBRACE}, {Kind: wall.IDENTIFIER, Content: []byte("age")}, {Kind: wall.COLON}, {Kind: wall.INTEGER}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name: wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{
				{
					Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("age")},
					Value: &wall.ParsedLiteralExpr{
						Token: wall.Token{Kind: wall.INTEGER},
					},
				},
			},
		},
	},
	{
		tokens: []wall.Token{{Kind: wall.IDENTIFIER, Content: []byte("Bob")}, {Kind: wall.LEFTBRACE}, {Kind: wall.NEWLINE}, {Kind: wall.IDENTIFIER, Content: []byte("age")}, {Kind: wall.COLON}, {Kind: wall.INTEGER}, {Kind: wall.COMMA}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}},
		expected: &wall.ParsedStructInitExpr{
			Name: wall.ParsedIdType{},
			Fields: []wall.ParsedStructInitField{
				{
					Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("age")},
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

func TestParseVarStmtWithValue(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.VAR}, {Kind: wall.IDENTIFIER}, {Kind: wall.EQ}, {Kind: wall.INTEGER}})
	stmt, err := pr.ParseStmtAndEof()
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(stmt), reflect.TypeOf(&wall.ParsedVar{}))
}

func TestParseVarStmtWithType(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.VAR}, {Kind: wall.IDENTIFIER}, {Kind: wall.IDENTIFIER}})
	stmt, err := pr.ParseStmtAndEof()
	if assert.NoError(t, err) {
		assert.Equal(t, reflect.TypeOf(stmt), reflect.TypeOf(&wall.ParsedVar{}))
	}
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
	}, &wall.ParsedFunDef{
		Fun: wall.Token{Kind: wall.FUN},
		Id:  wall.Token{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		Params: []wall.ParsedFunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
		},
		ReturnType: &wall.ParsedIdType{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
		},
		Body: &wall.ParsedBlock{
			Left:  wall.Token{Kind: wall.LEFTBRACE},
			Stmts: []wall.ParsedStmt{},
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
	}, &wall.ParsedFunDef{
		Fun: wall.Token{Kind: wall.FUN},
		Id:  wall.Token{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		Params: []wall.ParsedFunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
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
		{Kind: wall.IDENTIFIER, Content: []byte("main")},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.LEFTBRACE},
		{Kind: wall.RIGHTBRACE},
	}, &wall.ParsedFunDef{
		Fun:        wall.Token{Kind: wall.FUN},
		Id:         wall.Token{Kind: wall.IDENTIFIER, Content: []byte("main")},
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
		{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		{Kind: wall.LEFTPAREN},
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.COMMA},
		{Kind: wall.IDENTIFIER, Content: []byte("b")},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.RIGHTPAREN},
		{Kind: wall.IDENTIFIER, Content: []byte("int")},
		{Kind: wall.EOF},
	}, &wall.ParsedExternFunDef{
		Extern: wall.Token{Kind: wall.EXTERN},
		Fun:    wall.Token{Kind: wall.FUN},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("sum")},
		Params: []wall.ParsedFunParam{
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Id: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("b")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
		},
		ReturnType: &wall.ParsedIdType{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
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
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
	})
	got, err := pr.ParseDefAndEof()
	assert.NoError(t, err)
	expected := &wall.ParsedImport{
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
	expected := &wall.ParsedStructDef{
		Struct: wall.Token{Kind: wall.STRUCT},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("Employee")},
		Fields: []wall.ParsedStructField{
			{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("id")},
				Type: &wall.ParsedIdType{
					Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("int")},
				},
			},
			{
				Name: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("age")},
				Type: &wall.ParsedIdType{
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
	expected := &wall.ParsedFile{
		Defs: []wall.ParsedDef{
			&wall.ParsedImport{
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
