package wall

import (
	"strings"
)

func LowerExternFunctions(c *CheckedFile) {
	lowerExternFunctions(c, make(map[*CheckedFile]struct{}))
}

func lowerExternFunctions(c *CheckedFile, checked map[*CheckedFile]struct{}) {
	if _, ok := checked[c]; ok {
		return
	}
	checked[c] = struct{}{}
	for _, f := range c.ExternFuns {
		var builder strings.Builder
		builder.WriteString(string(f.Name.Content))
		builder.WriteString("(")
		for i, param := range f.Params {
			builder.WriteString(string(param.Name.Content))
			if i+1 < len(f.Params) {
				builder.WriteString(", ")
			}
		}
		builder.WriteString(")")
		inlineCText := builder.String()
		inlineCFun := c.GlobalScope.findFunction("inlineC")
		inlineC := &CheckedReturn{
			Value: &CheckedCallExpr{
				Callee: &CheckedIdExpr{
					Id:   inlineCFun.Token,
					Type: inlineCFun.TypeId,
				},
				Args: []CheckedExpr{
					&CheckedLiteralExpr{
						Literal: Token{
							Kind:    STRING,
							Content: inlineCText,
						},
						Type: c.TypeId(&PointerType{
							Type: CHAR_TYPE_ID,
						}),
					},
				},
				Type: f.ReturnType,
			},
		}
		c.Funs = append(c.Funs, &CheckedFunDef{
			Name:       f.Name,
			Params:     f.Params,
			ReturnType: f.ReturnType,
			Body: &CheckedBlock{
				Stmts: []CheckedStmt{inlineC},
			},
		})
	}
	for _, imp := range c.Imports {
		lowerExternFunctions(imp.File, checked)
	}
}
