package wall

type AstNode interface {
	pos() Pos
}

type DefNode interface {
	AstNode
	defNode()
}

type FunDef struct {
	Fun        Token
	Id         Token
	Params     []FunParam
	ReturnType TypeNode
	Body       StmtNode
}

type FunParam struct {
	Id   Token
	Type TypeNode
}

func (f *FunDef) pos() Pos {
	return f.Fun.Pos
}

func (f *FunDef) defNode() {}

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

type BlockStmt struct {
	Left  Token
	Stmts []StmtNode
	Right Token
}

func (v *VarStmt) pos() Pos {
	return v.Id.Pos
}
func (e *ExprStmt) pos() Pos {
	return e.Expr.pos()
}
func (b *BlockStmt) pos() Pos {
	return b.Left.Pos
}

func (v *VarStmt) stmtNode()   {}
func (e *ExprStmt) stmtNode()  {}
func (b *BlockStmt) stmtNode() {}

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

type TypeNode interface {
	AstNode
	typeNode()
}

type IdTypeNode struct {
	Token
}

func (i *IdTypeNode) pos() Pos {
	return i.Token.Pos
}

func (i *IdTypeNode) typeNode() {}
