package wall

import (
	"fmt"

	"github.com/cznic/mathutil"
)

type TokenKind int

const (
	EOF TokenKind = iota
	NEWLINE
	IDENTIFIER
	INTEGER
	STRING
	FLOAT
	PLUS
	MINUS
	STAR
	SLASH
	LEFTPAREN
	RIGHTPAREN
	LEFTBRACE
	RIGHTBRACE
	EQ
	COMMA
	COLON
	COLONCOLON
	DOT
	EQEQ
	BANGEQ
	LT
	LTEQ
	GT
	GTEQ
	AMP

	// keywords
	VAR
	FUN
	IMPORT
	STRUCT
	RETURN
	EXTERN
	TRUE
	FALSE
	IF
	ELSE
	AS
	WHILE
	BREAK
	CONTINUE
	TYPEALIAS
)

func (t TokenKind) String() string {
	switch t {
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case IDENTIFIER:
		return "IDENTIFIER"
	case INTEGER:
		return "INTEGER"
	case FLOAT:
		return "FLOAT"
	case STRING:
		return "STRING"
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case STAR:
		return "*"
	case SLASH:
		return "/"
	case LEFTPAREN:
		return "("
	case RIGHTPAREN:
		return ")"
	case LEFTBRACE:
		return "{"
	case RIGHTBRACE:
		return "}"
	case EQ:
		return "="
	case COMMA:
		return ","
	case COLON:
		return ":"
	case COLONCOLON:
		return "::"
	case DOT:
		return "."
	case EQEQ:
		return "=="
	case BANGEQ:
		return "!="
	case LT:
		return "<"
	case LTEQ:
		return "<="
	case GT:
		return ">"
	case GTEQ:
		return ">="
	case AMP:
		return "&"
	case VAR:
		return "VAR"
	case FUN:
		return "FUN"
	case IMPORT:
		return "IMPORT"
	case STRUCT:
		return "STRUCT"
	case RETURN:
		return "RETURN"
	case EXTERN:
		return "EXTERN"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case AS:
		return "AS"
	case WHILE:
		return "WHILE"
	case BREAK:
		return "BREAK"
	case CONTINUE:
		return "CONTINUE"
	case TYPEALIAS:
		return "TYPEALIAS"
	}
	panic("unreachable")
}

type Pos struct {
	Filename string
	Line     uint
}

func (p Pos) String() string {
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}

type Token struct {
	Pos
	Kind    TokenKind
	Content string
}

func ScanTokens(filename string, source string) ([]Token, error) {
	sc := NewScanner(filename, source)
	tokens := []Token{}
	for {
		tok, err := sc.Scan()
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, tok)
		if tok.Kind == EOF {
			break
		}
	}
	return tokens, nil
}

type Scanner struct {
	pos    Pos
	source string
	start  int
	end    int
}

func NewScanner(filename string, source string) Scanner {
	if len(source) == 0 {
		source = source + string([]byte{0})
	}
	const DEFAULT_LINE uint = 1
	return Scanner{
		pos: Pos{
			Filename: filename,
			Line:     DEFAULT_LINE,
		},
		source: source,
	}
}

func (s *Scanner) Scan() (Token, error) {
	s.skipWhitespace()
	s.start = s.end
	var t Token
	switch c := s.next(); c {
	case 0:
		s.advance()
		t = s.token(EOF)
	case '\n':
		s.advance()
		s.pos.Line++
		t = s.token(NEWLINE)
	case '+':
		s.advance()
		t = s.token(PLUS)
	case '-':
		s.advance()
		t = s.token(MINUS)
	case '*':
		s.advance()
		t = s.token(STAR)
	case '/':
		s.advance()
		t = s.token(SLASH)
	case '(':
		s.advance()
		t = s.token(LEFTPAREN)
	case ')':
		s.advance()
		t = s.token(RIGHTPAREN)
	case '{':
		s.advance()
		t = s.token(LEFTBRACE)
	case '}':
		s.advance()
		t = s.token(RIGHTBRACE)
	case '=':
		s.advance()
		if s.next() == '=' {
			s.advance()
			t = s.token(EQEQ)
		} else {
			t = s.token(EQ)
		}
	case '!':
		s.advance()
		if s.next() == '=' {
			s.advance()
			t = s.token(BANGEQ)
		} else {
			return s.token(EOF), NewError(s.pos, "unexpected character: %c", c)
		}
	case '<':
		s.advance()
		if s.next() == '=' {
			s.advance()
			t = s.token(LTEQ)
		} else {
			t = s.token(LT)
		}
	case '>':
		s.advance()
		if s.next() == '=' {
			s.advance()
			t = s.token(GTEQ)
		} else {
			t = s.token(GT)
		}
	case ',':
		s.advance()
		t = s.token(COMMA)
	case ':':
		s.advance()
		if s.next() == ':' {
			s.advance()
			t = s.token(COLONCOLON)
		} else {
			t = s.token(COLON)
		}
	case '.':
		s.advance()
		t = s.token(DOT)
	case '"':
		s.advance()
		return s.string()
	case '&':
		s.advance()
		t = s.token(AMP)
	default:
		if isId(c) {
			return s.id(), nil
		}
		if isNum(c) {
			return s.num(), nil
		}
		return s.token(EOF), NewError(s.pos, "unexpected character: %c", c)
	}
	return t, nil
}

func isId(c byte) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || c == '_'
}

func isNum(c byte) bool {
	return '0' <= c && c <= '9'
}

func (s *Scanner) id() Token {
	for {
		c := s.next()
		if !isId(c) && !isNum(c) {
			break
		}
		s.advance()
	}
	t := s.token(IDENTIFIER)
	switch t.Content {
	case "var":
		t.Kind = VAR
	case "fun":
		t.Kind = FUN
	case "import":
		t.Kind = IMPORT
	case "struct":
		t.Kind = STRUCT
	case "return":
		t.Kind = RETURN
	case "extern":
		t.Kind = EXTERN
	case "true":
		t.Kind = TRUE
	case "false":
		t.Kind = FALSE
	case "if":
		t.Kind = IF
	case "else":
		t.Kind = ELSE
	case "as":
		t.Kind = AS
	case "while":
		t.Kind = WHILE
	case "break":
		t.Kind = BREAK
	case "continue":
		t.Kind = CONTINUE
	case "typealias":
		t.Kind = TYPEALIAS
	}
	return t
}

func (s *Scanner) num() Token {
	for isNum(s.next()) {
		s.advance()
	}
	if s.next() == '.' {
		s.advance()
		for isNum(s.next()) {
			s.advance()
		}
		return s.token(FLOAT)
	}
	return s.token(INTEGER)
}

func (s *Scanner) string() (Token, error) {
	for s.next() != '"' && s.next() != 0 {
		if s.advance() == '\n' {
			s.pos.Line++
		}
	}
	if s.next() == 0 {
		return Token{}, NewError(s.pos, "a string literal is not terminated")
	}
	s.advance()
	t := s.token(STRING)
	t.Content = t.Content[1 : len(t.Content)-1]
	return t, nil
}

func (s *Scanner) skipWhitespace() {
	for {
		switch s.next() {
		case ' ', '\t', '\r':
			s.advance()
		case '/':
			if s.peek(1) == '/' {
				for s.next() != '\n' && s.next() != 0 {
					s.advance()
				}
			} else if s.peek(1) == '*' {
				for s.next() != 0 && !(s.next() == '*' && s.peek(1) == '/') {
					s.advance()
				}
				s.advance()
				s.advance()
			} else {
				return
			}
		default:
			return
		}
	}
}

func (s *Scanner) next() byte {
	return s.peek(0)
}

func (s *Scanner) peek(offset int) byte {
	if s.end+offset >= len(s.source) {
		return 0
	}
	return s.source[s.end+offset]
}

func (s *Scanner) advance() byte {
	c := s.next()
	s.end++
	return c
}

func (s *Scanner) token(t TokenKind) Token {
	end := mathutil.Clamp(s.end, 0, len(s.source))
	start := mathutil.Clamp(s.start, 0, len(s.source)-1)
	content := s.source[start:end]
	return Token{
		Pos: Pos{
			Filename: s.pos.Filename,
			Line:     s.pos.Line,
		},
		Kind:    t,
		Content: content,
	}
}
