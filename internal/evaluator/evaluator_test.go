package evaluator

import (
	"fmt"
	"testing"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
)

func TestEvaluateWalrusStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"edad := 25;", 25},
		{"NOMBRE := \"Wilson\";", "Wilson"},
		{"PI := 3.1416;", 3.1416},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObjectLiteral(t, evaluated, tt.expected)
	}
}

func TestConstantReassignment(t *testing.T) {
	input := `
NOMBRE := "Wilson";
NOMBRE = "Pedro";
`

	eval := NewEvaluator()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	err := eval.EvaluateProgram(program)
	if err == nil {
		t.Fatalf("Expected error for constant reassignment, but got none")
	}

	expectedError := "no se puede reasignar constante: NOMBRE"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestVariableReassignment(t *testing.T) {
	input := `
edad := 25;
edad = 30;
`

	evaluated := testEval(input)
	testObjectLiteral(t, evaluated, 30)
}

func TestTypedVariables(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"edad INT := 25;", 25},
		{"nombre STRING := \"Wilson\";", "Wilson"},
		{"valor := 10;", 10}, // ANY
		{"saludo := \"Hola\";", "Hola"}, // ANY
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObjectLiteral(t, evaluated, tt.expected)
	}
}

func TestTypeErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr string
	}{
		{
			`25 = "30";`,
			"left side of assignment must be assignable, got *ast.NumberLiteral",
		},
		{
			`"Wilson" = 123;`,
			"left side of assignment must be assignable, got *ast.StringLiteral",
		},
	}

	for _, tt := range tests {
		eval := NewEvaluator()
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			// Check if the error is the expected one
			if len(p.Errors()) == 1 && p.Errors()[0] == tt.expectedErr {
				return // Test passed - parser correctly caught the error
			}
			t.Fatalf("Parser errors: %v", p.Errors())
		}

		err := eval.EvaluateProgram(program)
		if err == nil {
			t.Fatalf("Expected error for type mismatch, but got none")
		}

		if err.Error() != tt.expectedErr {
			t.Fatalf("Expected error '%s', got '%s'", tt.expectedErr, err.Error())
		}
	}
}

func testEval(input string) Value {
	eval := NewEvaluator()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("Parser errors: " + fmt.Sprintf("%v", p.Errors()))
	}

	var lastValue Value = &Null{}
	for _, stmt := range program.Statements {
		value, err := eval.evaluateStatement(stmt)
		if err != nil {
			panic("Evaluation error: " + err.Error())
		}
		if value != nil {
			lastValue = value
		}
	}
	return lastValue
}

func testObjectLiteral(t *testing.T, obj Value, expected interface{}) bool {
	switch expected := expected.(type) {
	case int:
		return testIntegerObject(t, obj, int64(expected))
	case int64:
		return testIntegerObject(t, obj, expected)
	case float64:
		return testFloatObject(t, obj, expected)
	case string:
		return testStringObject(t, obj, expected)
	}
	t.Errorf("type of expected not handled. got=%T", expected)
	return false
}

func testIntegerObject(t *testing.T, obj Value, expected int64) bool {
	result, ok := obj.(*Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func testFloatObject(t *testing.T, obj Value, expected float64) bool {
	result, ok := obj.(*Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj Value, expected string) bool {
	result, ok := obj.(*String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
		return false
	}
	return true
}