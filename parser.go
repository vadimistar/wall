package wall

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

func (p *Parser) ParseStmtOrDefAndEof() (AstNode, error) {
	switch p.next().Kind {
	case FUN:
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
		body, err := p.ParseStmt()
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
	case LEFTBRACE:
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
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	return &ExprStmt{
		Expr: expr,
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
