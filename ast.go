package wall

type AstNode interface {
	pos() Pos
}

type ExprNode interface {
	AstNode
	exprNode()
}

type UnaryExprNode struct {
	operator Token
	operand  ExprNode
}

type BinaryExprNode struct {
	left  ExprNode
	op    Token
	right ExprNode
}

type GroupedExprNode struct {
	left  Token
	inner ExprNode
	right Token
}

type LiteralExprNode struct {
	Token
}

func (u UnaryExprNode) pos() Pos {
	return u.operator.Pos
}
func (b BinaryExprNode) pos() Pos {
	return b.left.pos()
}
func (g GroupedExprNode) pos() Pos {
	return g.left.Pos
}
func (l LiteralExprNode) pos() Pos {
	return l.Token.Pos
}

func (u UnaryExprNode) exprNode()   {}
func (b BinaryExprNode) exprNode()  {}
func (g GroupedExprNode) exprNode() {}
func (l LiteralExprNode) exprNode() {}
