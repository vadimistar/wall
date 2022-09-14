package wall

import (
	"bytes"
	"fmt"

	"github.com/cznic/mathutil"
)

type TokenKind int

const (
	EOF TokenKind = iota
	NEWLINE
	IDENTIFIER
	INTEGER
	FLOAT
	PLUS
	MINUS
	STAR
	SLASH
	LEFTPAREN
	RIGHTPAREN
	EQ

	// keywords
	VAR
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
	case EQ:
		return "="
	case VAR:
		return "VAR"
	}
	panic("unreachable")
}

type Pos struct {
	filename string
	line     uint
}

func (p Pos) String() string {
	return fmt.Sprintf("%s:%d", p.filename, p.line)
}

type Token struct {
	Pos
	Kind    TokenKind
	Content []byte
}

func ScanTokens(filename string, source []byte) ([]Token, error) {
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
	source []byte
	start  int
	end    int
}

func NewScanner(filename string, source []byte) Scanner {
	const DEFAULT_LINE uint = 1
	return Scanner{
		pos: Pos{
			filename: filename,
			line:     DEFAULT_LINE,
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
	case '=':
		s.advance()
		t = s.token(EQ)
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
	if bytes.Equal(t.Content, []byte("var")) {
		t.Kind = VAR
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

func (s *Scanner) skipWhitespace() {
	for {
		switch s.next() {
		case ' ', '\t', '\r':
			s.advance()
		default:
			return
		}
	}
}

func (s *Scanner) next() byte {
	if s.end >= len(s.source) {
		return 0
	}
	return s.source[s.end]
}

func (s *Scanner) advance() byte {
	c := s.next()
	s.end++
	return c
}

func (s *Scanner) token(t TokenKind) Token {
	end := mathutil.Clamp(s.end, 0, len(s.source))
	content := s.source[s.start:end]
	s.start = end
	return Token{
		Pos: Pos{
			filename: s.pos.filename,
			line:     s.pos.line,
		},
		Kind:    t,
		Content: content,
	}
}
