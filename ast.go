package wall

type AstNode interface {
	pos() Pos
}

type StmtNode interface {
	AstNode
	stmtNode()
}

type VarStmt struct {
	Var   Token
	Id    Token
	Eq    Token
	Value ExprNode
}

type ExprStmt struct {
	Expr ExprNode
}

func (v VarStmt) pos() Pos {
	return v.Id.Pos
}
func (e ExprStmt) pos() Pos {
	return e.Expr.pos()
}

func (v VarStmt) stmtNode()  {}
func (e ExprStmt) stmtNode() {}

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
