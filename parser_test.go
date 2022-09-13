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
}

func TestParseBinaryExpr(t *testing.T) {
	for _, op := range binaryOps {
		t.Logf("testing %+v", op)
		pr := wall.NewParser([]wall.Token{{Kind: wall.IDENTIFIER}, op, {Kind: wall.IDENTIFIER}, {Kind: wall.EOF}})
		expr, err := pr.ParseExprAndEof()
		if err != nil {
			t.Fatal(err)
		}
		if reflect.TypeOf(expr) != reflect.TypeOf(&wall.BinaryExprNode{}) {
			t.Fatalf("expected binary expression, but got %#v", expr)
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
		t.Fatalf("expected binary expression, but got %#v", expr)
	}
}
