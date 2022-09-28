package wall_test

import (
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
)

type evalExprTest struct {
	node   wall.ParsedExpr
	result wall.EvalObject
}

var evalExprTests = []evalExprTest{
	{&wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
	}, &wall.IntObject{
		Value: 123,
	}},
	{&wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.FLOAT, Content: []byte("0.123")},
	}, &wall.FloatObject{
		Value: 0.123,
	}},
	{&wall.ParsedGroupedExpr{
		Inner: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
		},
	}, &wall.IntObject{
		Value: 123,
	}},
	{&wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.PLUS},
		Operand: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
		},
	}, &wall.IntObject{
		Value: 123,
	}},
	{&wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.PLUS},
		Operand: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
		},
	}, &wall.FloatObject{
		Value: 1.0,
	}},
	{&wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.MINUS},
		Operand: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("123")},
		},
	}, &wall.IntObject{
		Value: -123,
	}},
	{&wall.ParsedUnaryExpr{
		Operator: wall.Token{Kind: wall.MINUS},
		Operand: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("1.0")},
		},
	}, &wall.FloatObject{
		Value: -1.0,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("3")},
		},
		Op: wall.Token{Kind: wall.PLUS},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: 8,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
		Op: wall.Token{Kind: wall.PLUS},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("5.0")},
		},
	}, &wall.FloatObject{
		Value: 3.0 + 5.0,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("3")},
		},
		Op: wall.Token{Kind: wall.MINUS},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: -2,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
		Op: wall.Token{Kind: wall.MINUS},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("5.0")},
		},
	}, &wall.FloatObject{
		Value: 3.0 - 5.0,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("3")},
		},
		Op: wall.Token{Kind: wall.STAR},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: 15,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("3.0")},
		},
		Op: wall.Token{Kind: wall.STAR},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("5.0")},
		},
	}, &wall.FloatObject{
		Value: 3.0 * 5.0,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("30")},
		},
		Op: wall.Token{Kind: wall.SLASH},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("5")},
		},
	}, &wall.IntObject{
		Value: 6,
	}},
	{&wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.FLOAT, Content: []byte("10.0")},
		},
		Op: wall.Token{Kind: wall.SLASH},
		Right: &wall.ParsedLiteralExpr{
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
		assert.NoError(t, err)
		assert.Equal(t, res, test.result)
	}
}

func TestEvalVarStmt(t *testing.T) {
	varStmt := &wall.ParsedVar{
		Var: wall.Token{},
		Id:  wall.Token{Content: []byte("a")},
		Eq:  wall.Token{},
		Value: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("10")},
		},
	}
	idExpr := &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
	}
	ev := wall.NewEvaluator()
	varRes, err := ev.EvaluateStmt(varStmt)
	assert.NoError(t, err)
	assert.Equal(t, varRes, &wall.UnitObject{})
	idRes, err := ev.EvaluateExpr(idExpr)
	assert.NoError(t, err)
	assert.Equal(t, idRes, &wall.IntObject{Value: 10})
}

func TestEvalAssignExpr(t *testing.T) {
	varStmt := &wall.ParsedVar{
		Var: wall.Token{},
		Id:  wall.Token{Content: []byte("a")},
		Eq:  wall.Token{},
		Value: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("0")},
		},
	}
	assignExpr := &wall.ParsedBinaryExpr{
		Left: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
		},
		Op: wall.Token{Kind: wall.EQ},
		Right: &wall.ParsedLiteralExpr{
			Token: wall.Token{Kind: wall.INTEGER, Content: []byte("20")},
		},
	}
	idExpr := &wall.ParsedLiteralExpr{
		Token: wall.Token{Kind: wall.IDENTIFIER, Content: []byte("a")},
	}
	ev := wall.NewEvaluator()
	varRes, err := ev.EvaluateStmt(varStmt)
	assert.NoError(t, err)
	assert.Equal(t, varRes, &wall.UnitObject{})
	assignRes, err := ev.EvaluateExpr(assignExpr)
	assert.NoError(t, err)
	assert.Equal(t, assignRes, &wall.IntObject{Value: 20})
	idRes, err := ev.EvaluateExpr(idExpr)
	assert.NoError(t, err)
	assert.Equal(t, idRes, &wall.IntObject{Value: 20})
}
