package wall

type ParsedNode interface {
	pos() Pos
}

type ParsedFile struct {
	Defs []ParsedDef
}

func (f *ParsedFile) pos() Pos {
	if len(f.Defs) == 0 {
		return Pos{}
	}
	return f.Defs[0].pos()
}

type ParsedDef interface {
	ParsedNode
	def()
	id() []byte
}

type ParsedFunDef struct {
	Fun        Token
	Id         Token
	Params     []ParsedFunParam
	ReturnType ParsedType
	Body       *ParsedBlock
}

type ParsedFunParam struct {
	Id   Token
	Type ParsedType
}

type ParsedImport struct {
	Import Token
	Name   Token
	File   *ParsedFile
}

type ParsedStructDef struct {
	Struct Token
	Name   Token
	Fields []ParsedStructField
}

type ParsedStructField struct {
	Name Token
	Type ParsedType
}

type ParsedExternFunDef struct {
	Extern     Token
	Fun        Token
	Name       Token
	Params     []ParsedFunParam
	ReturnType ParsedType
}

func (f *ParsedFunDef) pos() Pos {
	return f.Fun.Pos
}
func (i *ParsedImport) pos() Pos {
	return i.Import.Pos
}
func (s *ParsedStructDef) pos() Pos {
	return s.Struct.Pos
}
func (e *ParsedExternFunDef) pos() Pos {
	return e.Fun.Pos
}

func (f *ParsedFunDef) def()       {}
func (i *ParsedImport) def()       {}
func (s *ParsedStructDef) def()    {}
func (e *ParsedExternFunDef) def() {}

func (f *ParsedFunDef) id() []byte {
	return f.Id.Content
}
func (i *ParsedImport) id() []byte {
	return i.Name.Content
}
func (s *ParsedStructDef) id() []byte {
	return s.Name.Content
}
func (e *ParsedExternFunDef) id() []byte {
	return e.Name.Content
}

type ParsedStmt interface {
	ParsedNode
	stmt()
}

type ParsedVar struct {
	Var   Token
	Id    Token
	Eq    Token
	Value ParsedExpr
}

type ParsedExprStmt struct {
	Expr ParsedExpr
}

type ParsedBlock struct {
	Left  Token
	Stmts []ParsedStmt
	Right Token
}

type ParsedReturn struct {
	Return Token
	Arg    ParsedExpr
}

func (v *ParsedVar) pos() Pos {
	return v.Id.Pos
}
func (e *ParsedExprStmt) pos() Pos {
	return e.Expr.pos()
}
func (b *ParsedBlock) pos() Pos {
	return b.Left.Pos
}
func (r *ParsedReturn) pos() Pos {
	return r.Return.Pos
}

func (v *ParsedVar) stmt()      {}
func (e *ParsedExprStmt) stmt() {}
func (b *ParsedBlock) stmt()    {}
func (r *ParsedReturn) stmt()   {}

type ParsedExpr interface {
	ParsedNode
	expr()
}

type ParsedUnaryExpr struct {
	Operator Token
	Operand  ParsedExpr
}

type ParsedBinaryExpr struct {
	Left  ParsedExpr
	Op    Token
	Right ParsedExpr
}

type ParsedGroupedExpr struct {
	Left  Token
	Inner ParsedExpr
	Right Token
}

type ParsedLiteralExpr struct {
	Token
}

type ParsedIdExpr struct {
	Token
}

type ParsedCallExpr struct {
	Callee ParsedExpr
	Args   []ParsedExpr
}

type ParsedStructInitExpr struct {
	Name   ParsedIdType
	Fields []ParsedStructInitField
}

type ParsedStructInitField struct {
	Name  Token
	Value ParsedExpr
}

func (u ParsedUnaryExpr) pos() Pos {
	return u.Operator.Pos
}
func (b ParsedBinaryExpr) pos() Pos {
	return b.Left.pos()
}
func (g ParsedGroupedExpr) pos() Pos {
	return g.Left.Pos
}
func (l ParsedLiteralExpr) pos() Pos {
	return l.Token.Pos
}
func (i ParsedIdExpr) pos() Pos {
	return i.Token.Pos
}
func (c ParsedCallExpr) pos() Pos {
	return c.Callee.pos()
}
func (s ParsedStructInitExpr) pos() Pos {
	return s.Name.Pos
}

func (u ParsedUnaryExpr) expr()      {}
func (b ParsedBinaryExpr) expr()     {}
func (g ParsedGroupedExpr) expr()    {}
func (l ParsedLiteralExpr) expr()    {}
func (i ParsedIdExpr) expr()         {}
func (c ParsedCallExpr) expr()       {}
func (s ParsedStructInitExpr) expr() {}

type ParsedType interface {
	ParsedNode
	parsedType()
}

type ParsedIdType struct {
	Token
}

type ParsedPointerType struct {
	Star Token
	To   ParsedType
}

func (i *ParsedIdType) pos() Pos {
	return i.Token.Pos
}
func (p *ParsedPointerType) pos() Pos {
	return p.Star.Pos
}

func (i *ParsedIdType) parsedType()      {}
func (p *ParsedPointerType) parsedType() {}
