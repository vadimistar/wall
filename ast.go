package wall

type AstNode interface {
	pos() Pos
}

type ExprNode interface {
	AstNode
	exprNode()
}

type UnaryExprNode struct {
	Operator Token
	Operand  ExprNode
}

type BinaryExprNode struct {
	Left  ExprNode
	Op    Token
	Right ExprNode
}

type GroupedExprNode struct {
	Left  Token
	Inner ExprNode
	Right Token
}

type LiteralExprNode struct {
	Token
}

func (u UnaryExprNode) pos() Pos {
	return u.Operator.Pos
}
func (b BinaryExprNode) pos() Pos {
	return b.Left.pos()
}
func (g GroupedExprNode) pos() Pos {
	return g.Left.Pos
}
func (l LiteralExprNode) pos() Pos {
	return l.Token.Pos
}

func (u UnaryExprNode) exprNode()   {}
func (b BinaryExprNode) exprNode()  {}
func (g GroupedExprNode) exprNode() {}
func (l LiteralExprNode) exprNode() {}
