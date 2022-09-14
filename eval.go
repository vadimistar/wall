package wall

import (
	"fmt"
	"reflect"
	"strconv"
)

type Evaluator struct{}

func NewEvaluator() Evaluator {
	return Evaluator{}
}

func (e *Evaluator) EvaluateExpr(expr ExprNode) (EvalObject, error) {
	switch ex := expr.(type) {
	case UnaryExprNode:
		return e.evaluateUnaryExpr(ex)
	case BinaryExprNode:
		return e.evaluateBinaryExpr(ex)
	case LiteralExprNode:
		return e.evaluateLiteralExpr(ex)
	case GroupedExprNode:
		return e.evaluateGroupedExpr(ex)
	}
	panic("unreachable")
}

type EvalObject interface {
	evalObject()
	String() string
}

type IntObject struct {
	Value int64
}

type FloatObject struct {
	Value float64
}

func (i IntObject) evalObject()   {}
func (f FloatObject) evalObject() {}

func (i IntObject) String() string {
	return fmt.Sprintf("%d", i.Value)
}

func (f FloatObject) String() string {
	return fmt.Sprintf("%f", f.Value)
}

func (e *Evaluator) evaluateUnaryExpr(u UnaryExprNode) (EvalObject, error) {
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

func (e *Evaluator) evaluateBinaryExpr(b BinaryExprNode) (EvalObject, error) {
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

func (e *Evaluator) evaluateLiteralExpr(b LiteralExprNode) (EvalObject, error) {
	switch b.Token.Kind {
	case IDENTIFIER:
		panic("unimplemented")
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

func (e *Evaluator) evaluateGroupedExpr(b GroupedExprNode) (EvalObject, error) {
	return e.EvaluateExpr(b.Inner)
}
