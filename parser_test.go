package wall_test

import (
	"reflect"
	"testing"
	"wall"
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
