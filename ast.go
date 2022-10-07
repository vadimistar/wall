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
	id() string
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

type ParsedTypealiasDef struct {
	Typealias Token
	Name      Token
	Type      ParsedType
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
func (p *ParsedTypealiasDef) pos() Pos {
	return p.Typealias.Pos
}

func (f *ParsedFunDef) def()       {}
func (i *ParsedImport) def()       {}
func (s *ParsedStructDef) def()    {}
func (e *ParsedExternFunDef) def() {}
func (p *ParsedTypealiasDef) def() {}

func (f *ParsedFunDef) id() string {
	return f.Id.Content
}
func (i *ParsedImport) id() string {
	return i.Name.Content
}
func (s *ParsedStructDef) id() string {
	return s.Name.Content
}
func (e *ParsedExternFunDef) id() string {
	return e.Name.Content
}
func (p *ParsedTypealiasDef) id() string {
	return p.Name.Content
}

type ParsedStmt interface {
	ParsedNode
	stmt()
}

type ParsedVar struct {
	Id      Token
	ColonEq Token
	Value   ParsedExpr
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

type ParsedIf struct {
	If        Token
	Condition ParsedExpr
	Body      *ParsedBlock
	ElseBody  *ParsedBlock
}

type ParsedWhile struct {
	While     Token
	Condition ParsedExpr
	Body      *ParsedBlock
}

type ParsedBreak struct {
	Break Token
}

type ParsedContinue struct {
	Continue Token
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
func (i *ParsedIf) pos() Pos {
	return i.If.Pos
}
func (p ParsedWhile) pos() Pos {
	return p.While.Pos
}
func (p ParsedBreak) pos() Pos {
	return p.Break.Pos
}
func (p ParsedContinue) pos() Pos {
	return p.Continue.Pos
}

func (v *ParsedVar) stmt()      {}
func (e *ParsedExprStmt) stmt() {}
func (b *ParsedBlock) stmt()    {}
func (r *ParsedReturn) stmt()   {}
func (i *ParsedIf) stmt()       {}
func (p *ParsedWhile) stmt()    {}
func (p *ParsedBreak) stmt()    {}
func (p *ParsedContinue) stmt() {}

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
	Name   ParsedType
	Fields []ParsedStructInitField
}

type ParsedStructInitField struct {
	Name  Token
	Value ParsedExpr
}

type ParsedObjectAccessExpr struct {
	Object ParsedExpr
	Dot    Token
	Member Token
}

type ParsedModuleAccessExpr struct {
	Module     Token
	Coloncolon Token
	Member     ParsedExpr
}

type ParsedAsExpr struct {
	Value ParsedExpr
	As    Token
	Type  ParsedType
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
	return s.Name.pos()
}
func (p ParsedObjectAccessExpr) pos() Pos {
	return p.Object.pos()
}
func (p ParsedModuleAccessExpr) pos() Pos {
	return p.Module.Pos
}
func (p ParsedAsExpr) pos() Pos {
	return p.Value.pos()
}

func (u ParsedUnaryExpr) expr()        {}
func (b ParsedBinaryExpr) expr()       {}
func (g ParsedGroupedExpr) expr()      {}
func (l ParsedLiteralExpr) expr()      {}
func (i ParsedIdExpr) expr()           {}
func (c ParsedCallExpr) expr()         {}
func (s ParsedStructInitExpr) expr()   {}
func (a ParsedObjectAccessExpr) expr() {}
func (p ParsedModuleAccessExpr) expr() {}
func (p ParsedAsExpr) expr()           {}

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

type ParsedModuleAccessType struct {
	Module     Token
	Coloncolon Token
	Member     ParsedType
}

func (i *ParsedIdType) pos() Pos {
	return i.Token.Pos
}
func (p *ParsedPointerType) pos() Pos {
	return p.Star.Pos
}
func (p *ParsedModuleAccessType) pos() Pos {
	return p.Module.Pos
}

func (i *ParsedIdType) parsedType()           {}
func (p *ParsedPointerType) parsedType()      {}
func (p *ParsedModuleAccessType) parsedType() {}
