package wall

import (
	"os"
	"path/filepath"
)

func ParseFile(filename string, source []byte) (*FileNode, error) {
	tokens, err := ScanTokens(filename, source)
	if err != nil {
		return nil, err
	}
	psr := NewParser(tokens)
	return psr.ParseFile()
}

type Parser struct {
	tokens []Token
	index  int
}

func NewParser(tokens []Token) Parser {
	if len(tokens) == 0 {
		tokens = append(tokens, Token{})
	}
	if tokens[len(tokens)-1].Kind != EOF {
		tokens = append(tokens, Token{Kind: EOF})
	}
	return Parser{
		tokens: tokens,
		index:  0,
	}
}

func ParseCompilationUnit(filename string, source []byte) (*FileNode, error) {
	file, err := ParseFile(filename, source)
	if err != nil {
		return nil, err
	}
	if err := resolveImports(file, make(map[string]*FileNode)); err != nil {
		return nil, err
	}
	return file, nil
}

func resolveImports(file *FileNode, parsedModules map[string]*FileNode) error {
	for i, def := range file.Defs {
		switch importDef := def.(type) {
		case *ImportDef:
			resolvedImport, err := resolveImport(importDef, parsedModules)
			if err != nil {
				return err
			}
			file.Defs[i] = resolvedImport
		}
	}
	return nil
}

func resolveImport(def *ImportDef, parsedModules map[string]*FileNode) (*ParsedImportDef, error) {
	importedFilename := filepath.Join(filepath.Dir(def.Import.Filename), toModuleFilename(string(def.Name.Content)))
	absImportedFilename, err := filepath.Abs(importedFilename)
	if err != nil {
		return nil, err
	}
	if module, ok := parsedModules[absImportedFilename]; ok {
		return &ParsedImportDef{
			ImportDef:  *def,
			ParsedNode: module,
		}, nil
	}
	source, err := os.ReadFile(importedFilename)
	if err != nil {
		return nil, NewError(def.pos(), "unresolved import: %s (%s)", def.Name.Content, err)
	}
	parsedFile, err := ParseFile(importedFilename, source)
	if err != nil {
		return nil, err
	}
	parsedModules[absImportedFilename] = parsedFile
	if err := resolveImports(parsedFile, parsedModules); err != nil {
		return nil, err
	}
	return &ParsedImportDef{
		ImportDef:  *def,
		ParsedNode: parsedFile,
	}, nil
}

func toModuleFilename(name string) string {
	const WALL_EXTENSION = ".wl"
	return name + WALL_EXTENSION
}

func (p *Parser) ParseFile() (*FileNode, error) {
	defs := make([]DefNode, 0)
	for p.next().Kind != EOF {
		if p.next().Kind == NEWLINE {
			p.advance()
			continue
		}
		def, err := p.ParseDef()
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
		_, err = p.match(NEWLINE)
		if err != nil {
			return nil, err
		}
	}
	_ = p.advance()
	return &FileNode{
		Defs: defs,
	}, nil
}

func (p *Parser) ParseStmtOrDefAndEof() (AstNode, error) {
	switch p.next().Kind {
	case FUN, IMPORT, STRUCT:
		return p.ParseDefAndEof()
	default:
		return p.ParseStmtAndEof()
	}
}

func (p *Parser) ParseDef() (DefNode, error) {
	switch p.next().Kind {
	case FUN:
		fun := p.advance()
		id, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		params, err := p.parseFunParams()
		if err != nil {
			return nil, err
		}
		var returnType TypeNode = nil
		if p.next().Kind != LEFTBRACE {
			returnType, err = p.parseType()
			if err != nil {
				return nil, err
			}
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &FunDef{
			Fun:        fun,
			Id:         id,
			Params:     params,
			ReturnType: returnType,
			Body:       body,
		}, err
	case IMPORT:
		kw := p.advance()
		name, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		return &ImportDef{
			Import: kw,
			Name:   name,
		}, nil
	case STRUCT:
		kw := p.advance()
		name, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		fields, err := p.parseStructBody()
		if err != nil {
			return nil, err
		}
		return &StructDef{
			Struct: kw,
			Name:   name,
			Fields: fields,
		}, nil
	}
	return nil, NewError(p.next().Pos, "expected definition, but got %s", p.next().Kind)
}

func (p *Parser) ParseDefAndEof() (DefNode, error) {
	def, err := p.ParseDef()
	if err != nil {
		return nil, err
	}
	_, err = p.match(EOF)
	if err != nil {
		return nil, err
	}
	return def, nil
}

func (p *Parser) ParseStmt() (StmtNode, error) {
	switch p.next().Kind {
	case VAR:
		varToken := p.advance()
		id, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		eq, err := p.match(EQ)
		if err != nil {
			return nil, err
		}
		val, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		return &VarStmt{
			Var:   varToken,
			Id:    id,
			Eq:    eq,
			Value: val,
		}, nil
	case RETURN:
		kw := p.advance()
		if p.next().Kind != NEWLINE && p.next().Kind != EOF && p.next().Kind != RIGHTBRACE {
			arg, err := p.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &ReturnStmt{
				Return: kw,
				Arg:    arg,
			}, nil
		}
		return &ReturnStmt{
			Return: kw,
			Arg:    nil,
		}, nil
	case LEFTBRACE:
		return p.parseBlock()
	}
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	return &ExprStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseBlock() (*BlockStmt, error) {
	left := p.advance()
	stmts := make([]StmtNode, 0)
	for p.next().Kind != RIGHTBRACE {
		if p.next().Kind == NEWLINE {
			p.advance()
			continue
		}
		stmt, err := p.ParseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		_, err = p.match(NEWLINE)
		if err != nil {
			return nil, err
		}
	}
	right := p.advance()
	return &BlockStmt{
		Left:  left,
		Stmts: stmts,
		Right: right,
	}, nil
}

func (p *Parser) ParseStmtAndEof() (StmtNode, error) {
	stmt, err := p.ParseStmt()
	if err != nil {
		return nil, err
	}
	_, err = p.match(EOF)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) ParseExpr() (ExprNode, error) {
	lhs, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	return p.parseExpr(lhs, 0)
}

func (p *Parser) ParseExprAndEof() (ExprNode, error) {
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.match(EOF)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseExpr(lhs ExprNode, minPrec int) (ExprNode, error) {
	for precedence(p.next().Kind) >= minPrec {
		op := p.advance()
		rhs, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		next := p.next()
		for (precedence(next.Kind) > precedence(op.Kind)) ||
			(IsRightAssoc(next.Kind) && (precedence(next.Kind) == precedence(op.Kind))) {
			if IsRightAssoc(next.Kind) {
				rhs, err = p.parseExpr(rhs, precedence(op.Kind))
			} else {
				rhs, err = p.parseExpr(rhs, precedence(op.Kind)+1)
			}
			if err != nil {
				return nil, err
			}
			next = p.next()
		}
		lhs = &BinaryExprNode{
			Left:  lhs,
			Op:    op,
			Right: rhs,
		}
	}
	return lhs, nil
}

func IsRightAssoc(t TokenKind) bool {
	return t == EQ
}

func precedence(t TokenKind) int {
	switch t {
	case STAR, SLASH:
		return 20
	case PLUS, MINUS:
		return 10
	case EQ:
		return 1
	}
	return -1
}

func (p *Parser) parsePrimary() (ExprNode, error) {
	switch t := p.next(); t.Kind {
	case IDENTIFIER, INTEGER, FLOAT:
		t := p.advance()
		return &LiteralExprNode{Token: t}, nil
	case LEFTPAREN:
		left := p.advance()
		inner, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		right, err := p.match(RIGHTPAREN)
		if err != nil {
			return nil, err
		}
		return &GroupedExprNode{
			Left:  left,
			Inner: inner,
			Right: right,
		}, nil
	case PLUS, MINUS:
		operator := p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &UnaryExprNode{
			Operator: operator,
			Operand:  operand,
		}, nil
	}
	return nil, NewError(p.next().Pos, "expected primary expression, but got %s", p.next().Kind)
}

func (p *Parser) next() Token {
	if p.index >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.index]
}

func (p *Parser) advance() Token {
	t := p.next()
	p.index++
	return t
}

func (p *Parser) match(k TokenKind) (Token, error) {
	t := p.next()
	if t.Kind != k {
		return Token{Kind: k}, NewError(t.Pos, "expected %s, but got %s", k, t.Kind)
	}
	p.index++
	return t, nil
}

func (p *Parser) parseFunParams() (params []FunParam, err error) {
	params = make([]FunParam, 0)
	_, err = p.match(LEFTPAREN)
	if err != nil {
		return params, err
	}
	for p.next().Kind != RIGHTPAREN {
		id, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		typ, err := p.parseType()
		if err != nil {
			return params, err
		}
		params = append(params, FunParam{
			Id:   id,
			Type: typ,
		})
		if p.next().Kind == RIGHTPAREN {
			continue
		}
		if p.next().Kind == COMMA {
			p.advance()
			continue
		}
		return params, NewError(p.next().Pos, "expected ')' or ',', but got %s", p.next().Kind)
	}
	_, err = p.match(RIGHTPAREN)
	if err != nil {
		return params, err
	}
	return params, nil
}

func (p *Parser) parseType() (TypeNode, error) {
	switch p.next().Kind {
	case IDENTIFIER:
		tok := p.advance()
		return &IdTypeNode{
			Token: tok,
		}, nil
	}
	return nil, NewError(p.next().Pos, "expected type, but got %s", p.next().Kind)
}

func (p *Parser) parseStructBody() (fields []StructField, err error) {
	_, err = p.match(LEFTBRACE)
	if err != nil {
		return fields, err
	}
	fields = make([]StructField, 0)
	for p.next().Kind != RIGHTBRACE {
		if p.next().Kind == NEWLINE {
			p.advance()
			continue
		}
		name, err := p.match(IDENTIFIER)
		if err != nil {
			return fields, err
		}
		typ, err := p.parseType()
		if err != nil {
			return fields, err
		}
		fields = append(fields, StructField{
			Name: name,
			Type: typ,
		})
		if p.next().Kind == COMMA || p.next().Kind == NEWLINE {
			p.advance()
			continue
		}
		if p.next().Kind == RIGHTBRACE {
			break
		}
		return fields, NewError(p.next().Pos, "expected comma, newline or }, but got %s", p.next().Kind)
	}
	_, err = p.match(RIGHTBRACE)
	if err != nil {
		return nil, err
	}
	return fields, nil
}
