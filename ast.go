package wall

type AstNode interface {
	pos() Pos
}

type FileNode struct {
	Defs []DefNode
}

func (f *FileNode) pos() Pos {
	if len(f.Defs) == 0 {
		return Pos{}
	}
	return f.Defs[0].pos()
}

type DefNode interface {
	AstNode
	defNode()
	id() []byte
}

type FunDef struct {
	Fun        Token
	Id         Token
	Params     []FunParam
	ReturnType TypeNode
	Body       *BlockStmt
}

type FunParam struct {
	Id   Token
	Type TypeNode
}

type ImportDef struct {
	Import Token
	Name   Token
}

type ParsedImportDef struct {
	ImportDef
	ParsedNode *FileNode
}

type StructDef struct {
	Struct Token
	Name   Token
	Fields []StructField
}

type StructField struct {
	Name Token
	Type TypeNode
}

func (f *FunDef) pos() Pos {
	return f.Fun.Pos
}
func (i *ImportDef) pos() Pos {
	return i.Import.Pos
}
func (p *ParsedImportDef) pos() Pos {
	return p.Import.Pos
}
func (s *StructDef) pos() Pos {
	return s.Struct.Pos
}

func (f *FunDef) defNode()          {}
func (i *ImportDef) defNode()       {}
func (p *ParsedImportDef) defNode() {}
func (s *StructDef) defNode()       {}

func (f *FunDef) id() []byte {
	return f.Id.Content
}
func (i *ImportDef) id() []byte {
	return i.Name.Content
}
func (p *ParsedImportDef) id() []byte {
	return p.Name.Content
}
func (s *StructDef) id() []byte {
	return s.Name.Content
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

type BlockStmt struct {
	Left  Token
	Stmts []StmtNode
	Right Token
}

type ReturnStmt struct {
	Return Token
	Arg    ExprNode
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
func (r *ReturnStmt) pos() Pos {
	return r.Return.Pos
}

func (v *VarStmt) stmtNode()    {}
func (e *ExprStmt) stmtNode()   {}
func (b *BlockStmt) stmtNode()  {}
func (r *ReturnStmt) stmtNode() {}

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
