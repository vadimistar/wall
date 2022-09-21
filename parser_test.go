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
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(e, &test.expected) {
			t.Fatalf("expression %#v is not equal to %#v", e, &test.expected)
		}
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
		if err != nil {
			t.Fatal(err)
		}
		if reflect.TypeOf(expr) != reflect.TypeOf(&wall.UnaryExprNode{}) {
			t.Fatalf("expected unary expression, but got %#v", expr)
		}
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
		if err != nil {
			t.Fatal(err)
		}
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
			if !reflect.DeepEqual(res, expected) {
				t.Fatalf("expected %#v, but got %#v", expected, res)
			}
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
		if !reflect.DeepEqual(res, expected) {
			t.Fatalf("expected %#v, but got %#v", expected, res)
		}
	}
}

func TestParseGroupedExpr(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.LEFTPAREN}, {Kind: wall.IDENTIFIER}, {Kind: wall.RIGHTPAREN}})
	expr, err := pr.ParseExprAndEof()
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(expr) != reflect.TypeOf(&wall.GroupedExprNode{}) {
		t.Fatalf("expected grouped expression, but got %#v", expr)
	}
}

func TestParseVarStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.VAR}, {Kind: wall.IDENTIFIER}, {Kind: wall.EQ}, {Kind: wall.INTEGER}})
	stmt, err := pr.ParseStmtAndEof()
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(stmt) != reflect.TypeOf(&wall.VarStmt{}) {
		t.Fatalf("expected var statement, but got %#v", stmt)
	}
}

func TestParseBlockStmt(t *testing.T) {
	pr := wall.NewParser([]wall.Token{{Kind: wall.LEFTBRACE}, {Kind: wall.NEWLINE}, {Kind: wall.IDENTIFIER}, {Kind: wall.NEWLINE}, {Kind: wall.RIGHTBRACE}})
	got, err := pr.ParseStmtAndEof()
	if err != nil {
		t.Fatal(err)
	}
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
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, but got %#v", expected, got)
	}
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
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.expected) {
			t.Fatalf("expected %#v, but got %#v", test.expected, got)
		}
	}
}

func TestParseImportDef(t *testing.T) {
	pr := wall.NewParser([]wall.Token{
		{Kind: wall.IMPORT},
		{Kind: wall.IDENTIFIER, Content: []byte("a")},
	})
	got, err := pr.ParseDefAndEof()
	if err != nil {
		t.Fatal(err)
	}
	expected := &wall.ImportDef{
		Import: wall.Token{Kind: wall.IMPORT},
		Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, but got %#v", expected, got)
	}
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
	if err != nil {
		t.Fatal(err)
	}
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
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, but got %#v", expected, got)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	expected := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.ImportDef{
				Import: wall.Token{Kind: wall.IMPORT},
				Name:   wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
			},
			parseFunDefTests[0].expected,
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, but got %#v", expected, got)
	}
}

func TestParseCompilationUnit(t *testing.T) {
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
