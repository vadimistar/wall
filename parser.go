package wall

import (
	"os"
	"path/filepath"
)

func ParseFile(filename string, source string) (*ParsedFile, error) {
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

func ParseCompilationUnit(filename string, source string) (*ParsedFile, error) {
	file, err := ParseFile(filename, source)
	if err != nil {
		return nil, err
	}
	if err := resolveImports(file, make(map[string]*ParsedFile)); err != nil {
		return nil, err
	}
	return file, nil
}

func resolveImports(file *ParsedFile, parsedModules map[string]*ParsedFile) error {
	absImportedFilename, err := filepath.Abs(file.pos().Filename)
	if err != nil {
		return err
	}
	parsedModules[absImportedFilename] = file
	for i, def := range file.Defs {
		switch importDef := def.(type) {
		case *ParsedImport:
			resolvedImport, err := resolveImport(importDef, parsedModules)
			if err != nil {
				return err
			}
			file.Defs[i] = resolvedImport
		}
	}
	return nil
}

func resolveImport(def *ParsedImport, parsedModules map[string]*ParsedFile) (*ParsedImport, error) {
	importedFilename := filepath.Join(filepath.Dir(def.Import.Filename), moduleToFilename(string(def.Name.Content)))
	absImportedFilename, err := filepath.Abs(importedFilename)
	if err != nil {
		return nil, err
	}
	if module, ok := parsedModules[absImportedFilename]; ok {
		return &ParsedImport{
			Import: def.Import,
			Name:   def.Name,
			File:   module,
		}, nil
	}
	source, err := os.ReadFile(importedFilename)
	if err != nil {
		return nil, NewError(def.pos(), "unresolved import: %s (%s)", def.Name.Content, err)
	}
	parsedFile, err := ParseFile(importedFilename, string(source))
	if err != nil {
		return nil, err
	}
	parsedModules[absImportedFilename] = parsedFile
	if err := resolveImports(parsedFile, parsedModules); err != nil {
		return nil, err
	}
	return &ParsedImport{
		Import: def.Import,
		Name:   def.Name,
		File:   parsedFile,
	}, nil
}

func moduleToFilename(name string) string {
	const WALL_EXTENSION = ".wl"
	return name + WALL_EXTENSION
}

func (p *Parser) ParseFile() (*ParsedFile, error) {
	defs := make([]ParsedDef, 0)
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
	return &ParsedFile{
		Defs: defs,
	}, nil
}

func (p *Parser) ParseStmtOrDefAndEof() (ParsedNode, error) {
	switch p.next().Kind {
	case FUN, IMPORT, STRUCT:
		return p.ParseDefAndEof()
	default:
		return p.ParseStmtAndEof()
	}
}

func (p *Parser) ParseDef() (ParsedDef, error) {
	switch p.next().Kind {
	case EXTERN:
		extern := p.advance()
		if p.next().Kind == FUN {
			fun := p.advance()
			name, err := p.match(IDENTIFIER)
			if err != nil {
				return nil, err
			}
			params, err := p.parseFunParams()
			if err != nil {
				return nil, err
			}
			var returnType ParsedType = nil
			if p.next().Kind != NEWLINE {
				returnType, err = p.parseType()
				if err != nil {
					return nil, err
				}
			}
			return &ParsedExternFunDef{
				Extern:     extern,
				Fun:        fun,
				Name:       name,
				Params:     params,
				ReturnType: returnType,
			}, err
		}
		return nil, NewError(p.next().Pos, "expected FUN, but got %s", p.next().Kind)
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
		var returnType ParsedType = nil
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
		return &ParsedFunDef{
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
		return &ParsedImport{
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
		return &ParsedStructDef{
			Struct: kw,
			Name:   name,
			Fields: fields,
		}, nil
	case TYPEALIAS:
		typealias := p.advance()
		name, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		typ, err := p.parseType()
		if err != nil {
			return nil, err
		}
		return &ParsedTypealiasDef{
			Typealias: typealias,
			Name:      name,
			Type:      typ,
		}, nil
	}
	return nil, NewError(p.next().Pos, "expected definition, but got %s", p.next().Kind)
}

func (p *Parser) ParseDefAndEof() (ParsedDef, error) {
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

func (p *Parser) ParseStmt() (ParsedStmt, error) {
	switch p.next().Kind {
	case VAR:
		varToken := p.advance()
		id, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		if p.next().Kind == EQ {
			_, err = p.match(EQ)
			if err != nil {
				return nil, err
			}
			val, err := p.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &ParsedVar{
				Var:   varToken,
				Id:    id,
				Type:  nil,
				Value: val,
			}, nil
		}
		t, err := p.parseType()
		if err != nil {
			return nil, err
		}
		return &ParsedVar{
			Var:   varToken,
			Id:    id,
			Type:  t,
			Value: nil,
		}, nil
	case RETURN:
		kw := p.advance()
		if p.next().Kind != NEWLINE && p.next().Kind != EOF && p.next().Kind != RIGHTBRACE {
			arg, err := p.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &ParsedReturn{
				Return: kw,
				Arg:    arg,
			}, nil
		}
		return &ParsedReturn{
			Return: kw,
			Arg:    nil,
		}, nil
	case LEFTBRACE:
		return p.parseBlock()
	case IF:
		kw := p.advance()
		cond, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		if p.next().Kind == ELSE {
			p.advance()
			elseBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			return &ParsedIf{
				If:        kw,
				Condition: cond,
				Body:      body,
				ElseBody:  elseBody,
			}, nil
		}
		return &ParsedIf{
			If:        kw,
			Condition: cond,
			Body:      body,
		}, nil
	case WHILE:
		kw := p.advance()
		cond, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &ParsedWhile{
			While:     kw,
			Condition: cond,
			Body:      body,
		}, nil
	case BREAK:
		return &ParsedBreak{
			Break: p.advance(),
		}, nil
	case CONTINUE:
		return &ParsedContinue{
			Continue: p.advance(),
		}, nil
	}
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	return &ParsedExprStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseBlock() (*ParsedBlock, error) {
	left := p.advance()
	stmts := make([]ParsedStmt, 0)
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
	return &ParsedBlock{
		Left:  left,
		Stmts: stmts,
		Right: right,
	}, nil
}

func (p *Parser) ParseStmtAndEof() (ParsedStmt, error) {
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

func (p *Parser) ParseExpr() (ParsedExpr, error) {
	lhs, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	return p.parseExpr(lhs, 0)
}

func (p *Parser) ParseExprAndEof() (ParsedExpr, error) {
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

func (p *Parser) parseExpr(lhs ParsedExpr, minPrec int) (ParsedExpr, error) {
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
		lhs = &ParsedBinaryExpr{
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
	case LT, LTEQ, GT, GTEQ:
		return 6
	case EQEQ, BANGEQ:
		return 5
	case EQ:
		return 1
	}
	return -1
}

func (p *Parser) parsePrimary() (expr ParsedExpr, err error) {
	switch t := p.next(); t.Kind {
	case INTEGER, FLOAT, STRING, TRUE, FALSE:
		t := p.advance()
		expr = &ParsedLiteralExpr{Token: t}
	case IDENTIFIER:
		t := p.advance()
		if p.next().Kind == COLONCOLON {
			coloncolon := p.advance()
			member, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			return &ParsedModuleAccessExpr{
				Module:     t,
				Coloncolon: coloncolon,
				Member:     member,
			}, nil
		}
		if (p.next().Kind == LEFTBRACE && p.peek(1).Kind == RIGHTBRACE) ||
			(p.next().Kind == LEFTBRACE && p.peek(1).Kind == IDENTIFIER && p.peek(2).Kind == COLON) ||
			(p.next().Kind == LEFTBRACE && p.peek(1).Kind == NEWLINE && p.peek(2).Kind == IDENTIFIER && p.peek(3).Kind == COLON) {
			fields := make([]ParsedStructInitField, 0)
			p.advance()
			for p.next().Kind != RIGHTBRACE {
				if p.next().Kind == NEWLINE {
					p.advance()
					continue
				}
				name, err := p.match(IDENTIFIER)
				if err != nil {
					return nil, err
				}
				_, err = p.match(COLON)
				if err != nil {
					return nil, err
				}
				value, err := p.ParseExpr()
				if err != nil {
					return nil, err
				}
				fields = append(fields, ParsedStructInitField{
					Name:  name,
					Value: value,
				})
				if p.next().Kind == RIGHTBRACE {
					break
				}
				_, err = p.match(COMMA)
				if err != nil {
					return nil, err
				}
			}
			p.advance()
			expr = &ParsedStructInitExpr{
				Name: ParsedIdType{
					Token: t,
				},
				Fields: fields,
			}
		} else {
			expr = &ParsedIdExpr{Token: t}
		}
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
		expr = &ParsedGroupedExpr{
			Left:  left,
			Inner: inner,
			Right: right,
		}
	case PLUS, MINUS, AMP, STAR:
		operator := p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		expr = &ParsedUnaryExpr{
			Operator: operator,
			Operand:  operand,
		}
	default:
		return nil, NewError(p.next().Pos, "expected primary expression, but got %s", p.next().Kind)
	}
	for {
		if p.next().Kind == LEFTPAREN {
			p.advance()
			args := make([]ParsedExpr, 0)
			for p.next().Kind != RIGHTPAREN {
				arg, err := p.ParseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.next().Kind == RIGHTPAREN {
					break
				}
				if _, err := p.match(COMMA); err != nil {
					return nil, err
				}
			}
			if _, err := p.match(RIGHTPAREN); err != nil {
				return nil, err
			}
			expr = &ParsedCallExpr{
				Callee: expr,
				Args:   args,
			}
		} else if p.next().Kind == DOT {
			dot := p.advance()
			member, err := p.match(IDENTIFIER)
			if err != nil {
				return nil, err
			}
			expr = &ParsedObjectAccessExpr{
				Object: expr,
				Dot:    dot,
				Member: member,
			}
		} else if p.next().Kind == AS {
			as := p.advance()
			typ, err := p.parseType()
			if err != nil {
				return nil, err
			}
			return &ParsedAsExpr{
				Value: expr,
				As:    as,
				Type:  typ,
			}, nil
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) next() Token {
	return p.peek(0)
}

func (p *Parser) peek(offset int) Token {
	if p.index+offset >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.index+offset]
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

func (p *Parser) parseFunParams() (params []ParsedFunParam, err error) {
	params = make([]ParsedFunParam, 0)
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
		params = append(params, ParsedFunParam{
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

func (p *Parser) parseType() (ParsedType, error) {
	switch p.next().Kind {
	case LEFTPAREN:
		l := p.advance()
		_, err := p.match(RIGHTPAREN)
		if err != nil {
			return nil, err
		}
		return &ParsedIdType{
			Token: Token{Kind: IDENTIFIER, Content: "()", Pos: l.Pos},
		}, nil
	case IDENTIFIER:
		tok := p.advance()
		return &ParsedIdType{
			Token: tok,
		}, nil
	case STAR:
		star := p.advance()
		to, err := p.parseType()
		if err != nil {
			return nil, err
		}
		return &ParsedPointerType{
			Star: star,
			To:   to,
		}, nil
	}
	return nil, NewError(p.next().Pos, "expected type, but got %s", p.next().Kind)
}

func (p *Parser) parseStructBody() (fields []ParsedStructField, err error) {
	_, err = p.match(LEFTBRACE)
	if err != nil {
		return fields, err
	}
	fields = make([]ParsedStructField, 0)
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
		fields = append(fields, ParsedStructField{
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
