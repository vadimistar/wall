package wall_test

import (
	"log"
	"reflect"
	"testing"
	"wall"
)

type evalExprTest struct {
	node   wall.ExprNode
	result wall.EvalObject
}

var evalExprTests = []evalExprTest{
	{&wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
	}, &wall.IntObject{
		Value: 123,
	}},
	{&wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0.123")},
	}, &wall.FloatObject{
		Value: 0.123,
	}},
	{&wall.GroupedExprNode{
		Inner: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
		},
	}, &wall.IntObject{
		Value: 123,
	}},
	{&wall.UnaryExprNode{
		Operator: wall.Token{Kind: wall.PLUS},
		Operand: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
		},
	}, &wall.IntObject{
		Value: 123,
	}},
	{&wall.UnaryExprNode{
		Operator: wall.Token{Kind: wall.PLUS},
		Operand: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
		},
	}, &wall.FloatObject{
		Value: 1.0,
	}},
	{&wall.UnaryExprNode{
		Operator: wall.Token{Kind: wall.MINUS},
		Operand: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
		},
	}, &wall.IntObject{
		Value: -123,
	}},
	{&wall.UnaryExprNode{
		Operator: wall.Token{Kind: wall.MINUS},
		Operand: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
		},
	}, &wall.FloatObject{
		Value: -1.0,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("3")},
		},
		Op: wall.Token{Kind: wall.PLUS},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: 8,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
		Op: wall.Token{Kind: wall.PLUS},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("5.0")},
		},
	}, &wall.FloatObject{
		Value: 3.0 + 5.0,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("3")},
		},
		Op: wall.Token{Kind: wall.MINUS},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: -2,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
		Op: wall.Token{Kind: wall.MINUS},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("5.0")},
		},
	}, &wall.FloatObject{
		Value: 3.0 - 5.0,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("3")},
		},
		Op: wall.Token{Kind: wall.STAR},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: 15,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
		Op: wall.Token{Kind: wall.STAR},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("5.0")},
		},
	}, &wall.FloatObject{
		Value: 3.0 * 5.0,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("30")},
		},
		Op: wall.Token{Kind: wall.SLASH},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: 6,
	}},
	{&wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("10.0")},
		},
		Op: wall.Token{Kind: wall.SLASH},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
	}, &wall.FloatObject{
		Value: 10.0 / 3.0,
	}},
}

func TestEvalExpr(t *testing.T) {
	for _, test := range evalExprTests {
		t.Logf("testing %#v", test.node)
		ev := wall.NewEvaluator()
		res, err := ev.EvaluateExpr(test.node)
		if err != nil {
			log.Fatal(err)
		}
		if !reflect.DeepEqual(res, test.result) {
			log.Fatalf("expected result %#v, but got %#v", test.result, res)
		}
	}
}

func TestEvalVarStmt(t *testing.T) {
	varStmt := &wall.VarStmt{
		Var: wall.Token{},
		Id:  wall.Token{Content: []byte("a")},
		Eq:  wall.Token{},
		Value: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
		},
	}
	idExpr := &wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
	}
	ev := wall.NewEvaluator()
	varRes, err := ev.EvaluateStmt(varStmt)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(varRes, &wall.UnitObject{}) {
		t.Fatalf("expected %#v, but got %#v", &wall.UnitObject{}, varRes)
	}
	idRes, err := ev.EvaluateExpr(idExpr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(idRes, &wall.IntObject{Value: 10}) {
		t.Fatalf("expected %#v, but got %#v", &wall.IntObject{Value: 10}, varRes)
	}
}

func TestEvalAssignExpr(t *testing.T) {
	varStmt := &wall.VarStmt{
		Var: wall.Token{},
		Id:  wall.Token{Content: []byte("a")},
		Eq:  wall.Token{},
		Value: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
		},
	}
	assignExpr := &wall.BinaryExprNode{
		Left: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
		},
		Op: wall.Token{Kind: wall.EQ},
		Right: &wall.LiteralExprNode{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("20")},
		},
	}
	idExpr := &wall.LiteralExprNode{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
	}
	ev := wall.NewEvaluator()
	varRes, err := ev.EvaluateStmt(varStmt)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(varRes, &wall.UnitObject{}) {
		t.Fatalf("expected %#v, but got %#v", &wall.UnitObject{}, varRes)
	}
	assignRes, err := ev.EvaluateExpr(assignExpr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(assignRes, &wall.IntObject{Value: 20}) {
		t.Fatalf("expected %#v, but got %#v", &wall.IntObject{Value: 20}, assignRes)
	}
	idRes, err := ev.EvaluateExpr(idExpr)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(assignRes, &wall.IntObject{Value: 20}) {
		t.Fatalf("expected %#v, but got %#v", &wall.IntObject{Value: 20}, idRes)
	}
}
