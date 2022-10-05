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

func ParseCompilationUnit(filename, source, workpath string) (*ParsedFile, error) {
	parsed, err := resolveImports(Pos{}, filename, make(map[string]*ParsedFile), workpath)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func resolveImports(importPos Pos, filename string, parsedModules map[string]*ParsedFile, workpath string) (*ParsedFile, error) {
	absPath := filepath.Join(workpath, filename)
	if parsedFile, isParsed := parsedModules[absPath]; isParsed {
		return parsedFile, nil
	}
	source, err := os.ReadFile(absPath)
	if err != nil {
		return nil, NewError(importPos, "unresolved import: %s (%s)", filename, err)
	}
	parsedFile, err := ParseFile(filename, string(source))
	if err != nil {
		return nil, err
	}
	parsedModules[absPath] = parsedFile
	for _, def := range parsedFile.Defs {
		switch def := def.(type) {
		case *ParsedImport:
			resolved, err := resolveImports(def.pos(), moduleToFilename(def.id()), parsedModules, workpath)
			if err != nil {
				return nil, err
			}
			def.File = resolved
		}
	}
	return parsedFile, nil
}

// func resolveImport(def *ParsedImport, parsedModules map[string]*ParsedFile, filename, workpath string) (*ParsedImport, error) {
// 	if module, ok := parsedModules[filename]; ok {
// 		return &ParsedImport{
// 			Import: def.Import,
// 			Name:   def.Name,
// 			File:   module,
// 		}, nil
// 	}
// 	source, err := os.ReadFile(filename)
// 	if err != nil {
// 		return nil, NewError(def.pos(), "unresolved import: %s (%s)", def.Name.Content, err)
// 	}
// 	relPath, err := filepath.Rel(workpath, filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	parsedFile, err := ParseFile(relPath, string(source))
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := resolveImports(parsedFile, parsedModules, workpath); err != nil {
// 		return nil, err
// 	}
// 	return &ParsedImport{
// 		Import: def.Import,
// 		Name:   def.Name,
// 		File:   parsedFile,
// 	}, nil
// }

func moduleToFilename(name string) string {
	const WALL_EXTENSION = ".wall"
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

func isUnaryOp(t TokenKind) bool {
	return t == PLUS || t == MINUS || t == STAR || t == AMP
}

func (p *Parser) parsePrimary() (expr ParsedExpr, err error) {
	switch {
	case isUnaryOp(p.next().Kind):
		operator := p.advance()
		operand, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ParsedUnaryExpr{
			Operator: operator,
			Operand:  operand,
		}, nil
	default:
		switch p.next().Kind {
		case INTEGER, FLOAT, STRING, TRUE, FALSE:
			expr = &ParsedLiteralExpr{Token: p.advance()}
		case IDENTIFIER:
			expr, err = p.parseId()
			if err != nil {
				return nil, err
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
		default:
			return nil, NewError(p.next().Pos, "expected a primary expression, but got %s", p.next().Kind)
		}
	}
Loop:
	for {
		switch p.next().Kind {
		case LEFTPAREN:
			expr, err = p.parseArgs(expr)
			if err != nil {
				return nil, err
			}
		case DOT:
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
		case AS:
			as := p.advance()
			typ, err := p.parseType()
			if err != nil {
				return nil, err
			}
			expr = &ParsedAsExpr{
				Value: expr,
				As:    as,
				Type:  typ,
			}
			if err != nil {
				return nil, err
			}
		case LEFTBRACE:
			if (p.peek(1).Kind == RIGHTBRACE) || (p.peek(1).Kind == IDENTIFIER && p.peek(2).Kind == COLON) || (p.peek(1).Kind == NEWLINE && p.peek(2).Kind == IDENTIFIER && p.peek(3).Kind == COLON) {
				expr, err = p.parseStructInitBody(expr)
				if err != nil {
					return nil, err
				}
				continue
			}
			break Loop
		default:
			break Loop
		}
	}
	return
}

func (p *Parser) parseId() (expr ParsedExpr, err error) {
	if p.peek(1).Kind != COLONCOLON {
		return &ParsedIdExpr{Token: p.advance()}, nil
	}
	path := make([]Token, 0, 1)
	coloncolons := make([]Token, 0, 1)
	for p.peek(1).Kind == COLONCOLON {
		module, err := p.match(IDENTIFIER)
		if err != nil {
			return nil, err
		}
		path = append(path, module)
		coloncolons = append(coloncolons, p.advance())
	}
	id, err := p.match(IDENTIFIER)
	if err != nil {
		return nil, err
	}
	expr = &ParsedIdExpr{
		Token: id,
	}
	for i := len(path) - 1; i >= 0; i-- {
		expr = &ParsedModuleAccessExpr{
			Module:     path[i],
			Coloncolon: coloncolons[i],
			Member:     expr,
		}
	}
	return expr, nil
}

func (p *Parser) parseStructInitBody(name ParsedExpr) (*ParsedStructInitExpr, error) {
	typ, err := parsedExprToParsedType(name)
	if err != nil {
		return nil, err
	}
	_, err = p.match(LEFTBRACE)
	if err != nil {
		return nil, err
	}
	fields := make([]ParsedStructInitField, 0)
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
	_, err = p.match(RIGHTBRACE)
	if err != nil {
		return nil, err
	}
	return &ParsedStructInitExpr{
		Name:   typ,
		Fields: fields,
	}, nil
}

func parsedExprToParsedType(p ParsedExpr) (ParsedType, error) {
	switch p := p.(type) {
	case *ParsedGroupedExpr:
		return parsedExprToParsedType(p.Inner)
	case *ParsedIdExpr:
		return &ParsedIdType{
			Token: p.Token,
		}, nil
	case *ParsedModuleAccessExpr:
		member, err := parsedExprToParsedType(p.Member)
		if err != nil {
			return nil, err
		}
		return &ParsedModuleAccessType{
			Module:     p.Module,
			Coloncolon: p.Coloncolon,
			Member:     member,
		}, nil
	}
	return nil, NewError(p.pos(), "an invalid type in the struct initializer")
}

func (p *Parser) parseArgs(callee ParsedExpr) (*ParsedCallExpr, error) {
	_, err := p.match(LEFTPAREN)
	if err != nil {
		return nil, err
	}
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
	return &ParsedCallExpr{
		Callee: callee,
		Args:   args,
	}, nil
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
		expr, err := p.parseId()
		if err != nil {
			return nil, err
		}
		return parsedExprToParsedType(expr)
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
