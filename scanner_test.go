package wall_test

import (
	"reflect"
	"testing"
	"wall"
)

type scanTokensTest struct {
	source   []byte
	expected []wall.TokenKind
}

var scanTokensTests = []scanTokensTest{
	{[]byte(""), []wall.TokenKind{wall.EOF}},
	{[]byte("\t"), []wall.TokenKind{wall.EOF}},
	{[]byte("\r"), []wall.TokenKind{wall.EOF}},
	{[]byte("\r\n"), []wall.TokenKind{wall.NEWLINE, wall.EOF}},
	{[]byte("\n"), []wall.TokenKind{wall.NEWLINE, wall.EOF}},
	{[]byte("abc"), []wall.TokenKind{wall.IDENTIFIER, wall.EOF}},
	{[]byte("123"), []wall.TokenKind{wall.INTEGER, wall.EOF}},
	{[]byte("123*123"), []wall.TokenKind{wall.INTEGER, wall.STAR, wall.INTEGER, wall.EOF}},
	{[]byte("-1"), []wall.TokenKind{wall.MINUS, wall.INTEGER, wall.EOF}},
	{[]byte("0.0"), []wall.TokenKind{wall.FLOAT, wall.EOF}},
	{[]byte("+"), []wall.TokenKind{wall.PLUS, wall.EOF}},
	{[]byte("-"), []wall.TokenKind{wall.MINUS, wall.EOF}},
	{[]byte("*"), []wall.TokenKind{wall.STAR, wall.EOF}},
	{[]byte("/"), []wall.TokenKind{wall.SLASH, wall.EOF}},
	{[]byte("()"), []wall.TokenKind{wall.LEFTPAREN, wall.RIGHTPAREN, wall.EOF}},
	{[]byte("{}"), []wall.TokenKind{wall.LEFTBRACE, wall.RIGHTBRACE, wall.EOF}},
	{[]byte(","), []wall.TokenKind{wall.COMMA, wall.EOF}},
	{[]byte("var"), []wall.TokenKind{wall.VAR, wall.EOF}},
	{[]byte("fun"), []wall.TokenKind{wall.FUN, wall.EOF}},
	{[]byte("import"), []wall.TokenKind{wall.IMPORT, wall.EOF}},
	{[]byte("struct"), []wall.TokenKind{wall.STRUCT, wall.EOF}},
}

func TestScanTokens(t *testing.T) {
	for _, test := range scanTokensTests {
		t.Logf("running test '%s'", test.source)
		tokens, err := wall.ScanTokens("<test>", test.source)
		if err != nil {
			t.Fatal(err)
		}
		kinds := []wall.TokenKind{}
		for _, tok := range tokens {
			kinds = append(kinds, tok.Kind)
		}
		if !reflect.DeepEqual(kinds, test.expected) {
			t.Fatalf("%#v is not equal to %#v", kinds, test.expected)
		}
	}
}

type scannerScanTest struct {
	source  []byte
	kind    wall.TokenKind
	content []byte
}

var scannerScanTests = []scannerScanTest{
	{[]byte("123"), wall.INTEGER, []byte("123")},
	{[]byte("123*123"), wall.INTEGER, []byte("123")},
	{[]byte("0.0"), wall.FLOAT, []byte("0.0")},
	{[]byte("a"), wall.IDENTIFIER, []byte("a")},
}

func TestScanner_Scan(t *testing.T) {
	for _, test := range scannerScanTests {
		t.Logf("running test '%s'", test.source)
		sc := wall.NewScanner("<test>", test.source)
		tok, err := sc.Scan()
		if err != nil {
			t.Fatal(err)
		}
		if tok.Kind != test.kind {
			t.Errorf("token kind %s is not equal to %s", tok.Kind, test.kind)
		}
		if !reflect.DeepEqual(tok.Content, test.content) {
			t.Errorf("token content %s is not equal to %s", tok.Content, test.content)
		}
	}
}
