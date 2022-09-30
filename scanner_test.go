package wall_test

import (
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
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
	{[]byte("\"abc\""), []wall.TokenKind{wall.STRING, wall.EOF}},
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
	{[]byte(":"), []wall.TokenKind{wall.COLON, wall.EOF}},
	{[]byte("var"), []wall.TokenKind{wall.VAR, wall.EOF}},
	{[]byte("fun"), []wall.TokenKind{wall.FUN, wall.EOF}},
	{[]byte("import"), []wall.TokenKind{wall.IMPORT, wall.EOF}},
	{[]byte("struct"), []wall.TokenKind{wall.STRUCT, wall.EOF}},
	{[]byte("return"), []wall.TokenKind{wall.RETURN, wall.EOF}},
}

func TestScanTokens(t *testing.T) {
	for _, test := range scanTokensTests {
		t.Logf("running test '%s'", test.source)
		tokens, err := wall.ScanTokens("<test>", test.source)
		assert.NoError(t, err)
		kinds := []wall.TokenKind{}
		for _, tok := range tokens {
			kinds = append(kinds, tok.Kind)
		}
		assert.Equal(t, kinds, test.expected)
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
	{[]byte("\"a\""), wall.STRING, []byte("a")},
}

func TestScanner_Scan(t *testing.T) {
	for _, test := range scannerScanTests {
		t.Logf("running test '%s'", test.source)
		sc := wall.NewScanner("<test>", test.source)
		tok, err := sc.Scan()
		assert.NoError(t, err)
		assert.Equal(t, tok.Kind, test.kind)
		assert.Equal(t, tok.Content, test.content)
	}
}
