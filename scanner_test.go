package wall_test

import (
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
)

type scanTokensTest struct {
	source   string
	expected []wall.TokenKind
}

var scanTokensTests = []scanTokensTest{
	{"", []wall.TokenKind{wall.EOF}},
	{"\t", []wall.TokenKind{wall.EOF}},
	{"\r", []wall.TokenKind{wall.EOF}},
	{"\r\n", []wall.TokenKind{wall.NEWLINE, wall.EOF}},
	{"\n", []wall.TokenKind{wall.NEWLINE, wall.EOF}},
	{"abc", []wall.TokenKind{wall.IDENTIFIER, wall.EOF}},
	{"123", []wall.TokenKind{wall.INTEGER, wall.EOF}},
	{"\"abc\"", []wall.TokenKind{wall.STRING, wall.EOF}},
	{"123*123", []wall.TokenKind{wall.INTEGER, wall.STAR, wall.INTEGER, wall.EOF}},
	{"-1", []wall.TokenKind{wall.MINUS, wall.INTEGER, wall.EOF}},
	{"0.0", []wall.TokenKind{wall.FLOAT, wall.EOF}},
	{"+", []wall.TokenKind{wall.PLUS, wall.EOF}},
	{"-", []wall.TokenKind{wall.MINUS, wall.EOF}},
	{"*", []wall.TokenKind{wall.STAR, wall.EOF}},
	{"/", []wall.TokenKind{wall.SLASH, wall.EOF}},
	{"()", []wall.TokenKind{wall.LEFTPAREN, wall.RIGHTPAREN, wall.EOF}},
	{"{}", []wall.TokenKind{wall.LEFTBRACE, wall.RIGHTBRACE, wall.EOF}},
	{",", []wall.TokenKind{wall.COMMA, wall.EOF}},
	{":", []wall.TokenKind{wall.COLON, wall.EOF}},
	{"::", []wall.TokenKind{wall.COLONCOLON, wall.EOF}},
	{":=", []wall.TokenKind{wall.COLONEQ, wall.EOF}},
	{".", []wall.TokenKind{wall.DOT, wall.EOF}},
	{"=", []wall.TokenKind{wall.EQ, wall.EOF}},
	{"==", []wall.TokenKind{wall.EQEQ, wall.EOF}},
	{"!=", []wall.TokenKind{wall.BANGEQ, wall.EOF}},
	{"<", []wall.TokenKind{wall.LT, wall.EOF}},
	{"<=", []wall.TokenKind{wall.LTEQ, wall.EOF}},
	{">", []wall.TokenKind{wall.GT, wall.EOF}},
	{">=", []wall.TokenKind{wall.GTEQ, wall.EOF}},
	{"&", []wall.TokenKind{wall.AMP, wall.EOF}},
	{"fun", []wall.TokenKind{wall.FUN, wall.EOF}},
	{"import", []wall.TokenKind{wall.IMPORT, wall.EOF}},
	{"struct", []wall.TokenKind{wall.STRUCT, wall.EOF}},
	{"return", []wall.TokenKind{wall.RETURN, wall.EOF}},
	{"extern", []wall.TokenKind{wall.EXTERN, wall.EOF}},
	{"true", []wall.TokenKind{wall.TRUE, wall.EOF}},
	{"false", []wall.TokenKind{wall.FALSE, wall.EOF}},
	{"if", []wall.TokenKind{wall.IF, wall.EOF}},
	{"else", []wall.TokenKind{wall.ELSE, wall.EOF}},
	{"as", []wall.TokenKind{wall.AS, wall.EOF}},
	{"while", []wall.TokenKind{wall.WHILE, wall.EOF}},
	{"break", []wall.TokenKind{wall.BREAK, wall.EOF}},
	{"continue", []wall.TokenKind{wall.CONTINUE, wall.EOF}},
	{"typealias", []wall.TokenKind{wall.TYPEALIAS, wall.EOF}},
	{"mut", []wall.TokenKind{wall.MUT, wall.EOF}},
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
	source  string
	kind    wall.TokenKind
	content string
}

var scannerScanTests = []scannerScanTest{
	{"123", wall.INTEGER, "123"},
	{"123*123", wall.INTEGER, "123"},
	{"0.0", wall.FLOAT, "0.0"},
	{"a", wall.IDENTIFIER, "a"},
	{"\"a\"", wall.STRING, "a"},
	{"\"\\a\"", wall.STRING, "\a"},
	{"\"\\a\"", wall.STRING, "\a"},
	{"\"\\b\"", wall.STRING, "\b"},
	{"\"\\f\"", wall.STRING, "\f"},
	{"\"\\n\"", wall.STRING, "\n"},
	{"\"\\r\"", wall.STRING, "\r"},
	{"\"\\t\"", wall.STRING, "\t"},
	{"\"\\v\"", wall.STRING, "\v"},
	{`"\\"`, wall.STRING, "\\"},
	{`"\""`, wall.STRING, "\""},
}

func TestScanner_Scan(t *testing.T) {
	for _, test := range scannerScanTests {
		t.Logf("running test '%s'", test.source)
		sc := wall.NewScanner("<test>", test.source)
		tok, err := sc.Scan()
		assert.NoError(t, err)
		assert.Equal(t, test.kind, tok.Kind)
		assert.Equal(t, test.content, tok.Content)
	}
}
