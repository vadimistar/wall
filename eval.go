package wall

import (
	"fmt"
	"reflect"
	"strconv"
)

type Evaluator struct {
	scopes []scope
}

type scope struct {
	vars map[string]EvalObject
}

func newScope() scope {
	return scope{
		vars: map[string]EvalObject{},
	}
}

func NewEvaluator() Evaluator {
	return Evaluator{
		scopes: append(make([]scope, 0), newScope()),
	}
}

func (e *Evaluator) EvaluateStmt(stmt StmtNode) (EvalObject, error) {
	switch st := stmt.(type) {
	case *VarStmt:
		return e.evaluateVarStmt(st)
	case *ExprStmt:
		return e.evaluateExprStmt(st)
	case *BlockStmt:
		return e.evaluateBlockStmt(st)
	}
	panic(fmt.Sprintf("unreachable: %T", stmt))
}

func (e *Evaluator) EvaluateExpr(expr ExprNode) (EvalObject, error) {
	switch ex := expr.(type) {
	case *UnaryExprNode:
		return e.evaluateUnaryExpr(ex)
	case *BinaryExprNode:
		return e.evaluateBinaryExpr(ex)
	case *LiteralExprNode:
		return e.evaluateLiteralExpr(ex)
	case *GroupedExprNode:
		return e.evaluateGroupedExpr(ex)
	}
	panic(fmt.Sprintf("unreachable: %T", expr))
}

type EvalObject interface {
	evalObject()
	String() string
}

type UnitObject struct{}

type IntObject struct {
	Value int64
}

type FloatObject struct {
	Value float64
}

func (u UnitObject) evalObject()  {}
func (i IntObject) evalObject()   {}
func (f FloatObject) evalObject() {}

func (u UnitObject) String() string {
	return "()"
}

func (i IntObject) String() string {
	return fmt.Sprintf("%d", i.Value)
}

func (f FloatObject) String() string {
	return fmt.Sprintf("%f", f.Value)
}

func (e *Evaluator) evaluateUnaryExpr(u *UnaryExprNode) (EvalObject, error) {
	operand, err := e.EvaluateExpr(u.Operand)
	if err != nil {
		return nil, err
	}
	switch u.Operator.Kind {
	case PLUS:
		return operand, nil
	case MINUS:
		switch obj := operand.(type) {
		case *IntObject:
			return &IntObject{Value: -obj.Value}, nil
		case *FloatObject:
			return &FloatObject{Value: -obj.Value}, nil
		}
	}
	return nil, NewError(u.Operator.Pos, "operator %s is not implemented for %T", u.Operator.Kind, operand)
}

func (e *Evaluator) evaluateBinaryExpr(b *BinaryExprNode) (EvalObject, error) {
	if b.Op.Kind == EQ {
		return e.evaluateAssignExpr(b)
	}
	left, err := e.EvaluateExpr(b.Left)
	if err != nil {
		return nil, err
	}
	right, err2 := e.EvaluateExpr(b.Right)
	if err2 != nil {
		return nil, err2
	}
	if reflect.TypeOf(left) == reflect.TypeOf(right) {
		switch b.Op.Kind {
		case PLUS:
			switch left := left.(type) {
			case *IntObject:
				right := right.(*IntObject)
				return &IntObject{Value: left.Value + right.Value}, nil
			case *FloatObject:
				right := right.(*FloatObject)
				return &FloatObject{Value: left.Value + right.Value}, nil
			}
		case MINUS:
			switch left := left.(type) {
			case *IntObject:
				right := right.(*IntObject)
				return &IntObject{Value: left.Value - right.Value}, nil
			case *FloatObject:
				right := right.(*FloatObject)
				return &FloatObject{Value: left.Value - right.Value}, nil
			}
		case STAR:
			switch left := left.(type) {
			case *IntObject:
				right := right.(*IntObject)
				return &IntObject{Value: left.Value * right.Value}, nil
			case *FloatObject:
				right := right.(*FloatObject)
				return &FloatObject{Value: left.Value * right.Value}, nil
			}
		case SLASH:
			switch left := left.(type) {
			case *IntObject:
				right := right.(*IntObject)
				return &IntObject{Value: left.Value / right.Value}, nil
			case *FloatObject:
				right := right.(*FloatObject)
				return &FloatObject{Value: left.Value / right.Value}, nil
			}
		}
	}
	return nil, NewError(b.Op.Pos, "operator %s is not implemented for types %T and %T", b.Op.Kind, left, right)
}

func (e *Evaluator) evaluateAssignExpr(b *BinaryExprNode) (EvalObject, error) {
	if reflect.TypeOf(b.Left) != reflect.TypeOf(&LiteralExprNode{}) {
		return nil, NewError(b.Left.pos(), "can't assign not to id")
	}
	left := b.Left.(*LiteralExprNode)
	if left.Token.Kind != IDENTIFIER {
		return nil, NewError(b.Left.pos(), "can't assign not to id: %s", left.Token.Content)
	}
	id := string(left.Token.Content)
	right, err := e.EvaluateExpr(b.Right)
	if err != nil {
		return nil, err
	}
	if _, ok := e.lookupVar(id); ok {
		e.assignToVar(id, right)
		return right, nil
	}
	return nil, NewError(left.pos(), "not declared: %s", id)
}

func (e *Evaluator) evaluateLiteralExpr(b *LiteralExprNode) (EvalObject, error) {
	switch b.Token.Kind {
	case IDENTIFIER:
		if val, ok := e.lookupVar(string(b.Token.Content)); ok {
			return val, nil
		}
		return nil, NewError(b.Token.Pos, "undeclared name: %s", b.Token.Content)
	case INTEGER:
		val, err := strconv.ParseInt(string(b.Token.Content), 10, 64)
		if err != nil {
			return nil, err
		}
		return &IntObject{Value: val}, nil
	case FLOAT:
		val, err := strconv.ParseFloat(string(b.Token.Content), 64)
		if err != nil {
			return nil, err
		}
		return &FloatObject{Value: val}, nil
	}
	panic("unreachable")
}

func (e *Evaluator) evaluateGroupedExpr(b *GroupedExprNode) (EvalObject, error) {
	return e.EvaluateExpr(b.Inner)
}

func (e *Evaluator) evaluateVarStmt(stmt *VarStmt) (EvalObject, error) {
	val, err := e.EvaluateExpr(stmt.Value)
	if err != nil {
		return nil, err
	}
	e.declareVar(string(stmt.Id.Content), val)
	return &UnitObject{}, nil
}

func (e *Evaluator) evaluateExprStmt(stmt *ExprStmt) (EvalObject, error) {
	return e.EvaluateExpr(stmt.Expr)
}

func (e *Evaluator) evaluateBlockStmt(block *BlockStmt) (EvalObject, error) {
	if len(block.Stmts) == 0 {
		return &UnitObject{}, nil
	}
	e.pushScope()
	var res EvalObject
	for i, stmtNode := range block.Stmts {
		stmt, err := e.EvaluateStmt(stmtNode)
		if err != nil {
			return nil, err
		}
		if i == len(block.Stmts)-1 {
			res = stmt
		}
	}
	e.popScope()
	return res, nil
}

func (e *Evaluator) lookupVar(id string) (EvalObject, bool) {
	for i := len(e.scopes) - 1; i >= 0; i-- {
		if obj, ok := e.scopes[i].vars[id]; ok {
			return obj, true
		}
	}
	return nil, false
}

func (e *Evaluator) assignToVar(id string, val EvalObject) {
	for i := len(e.scopes) - 1; i >= 0; i-- {
		if _, ok := e.scopes[i].vars[id]; ok {
			e.scopes[i].vars[id] = val
		}
	}
}

func (e *Evaluator) declareVar(id string, val EvalObject) {
	e.scopes[len(e.scopes)-1].vars[id] = val
}

func (e *Evaluator) pushScope() {
	e.scopes = append(e.scopes, newScope())
}

func (e *Evaluator) popScope() {
	e.scopes = e.scopes[:len(e.scopes)-1]
}
