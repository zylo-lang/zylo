package parser

import (
	"testing"
	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/lexer"
)

func TestVarStatements(t *testing.T) {
	input := `
x := 5
y := 10
foobar := 838383
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testVarStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch expected.(type) {
	case int64:
		intVal, ok := exp.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("exp not *ast.NumberLiteral. got=%T", exp)
			return false
		}
		if intVal.Value != expected {
			t.Errorf("intVal.Value not %d. got=%d", expected, intVal.Value)
			return false
		}
	case float64:
		// Nota: El AST actual solo soporta int64 para NumberLiteral.
		// Esto requerirá una refactorización futura para manejar floats correctamente.
		// Por ahora, solo verificamos que el tipo sea NumberLiteral.
		_, ok := exp.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("exp not *ast.NumberLiteral. got=%T", exp)
			return false
		}
	case string:
		strVal, ok := exp.(*ast.StringLiteral)
		if !ok {
			t.Errorf("exp not *ast.StringLiteral. got=%T", exp)
			return false
		}
		if strVal.Value != expected {
			t.Errorf("strVal.Value not %q. got=%q", expected, strVal.Value)
			return false
		}
	case bool:
		boolVal, ok := exp.(*ast.BooleanLiteral)
		if !ok {
			t.Errorf("exp not *ast.BooleanLiteral. got=%T", exp)
			return false
		}
		if boolVal.Value != expected {
			t.Errorf("boolVal.Value not %t. got=%t", expected, boolVal.Value)
			return false
		}
	default:
		t.Errorf("unhandled type for literal test: %T", expected)
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testVarStatement(t *testing.T, s ast.Statement, name string) bool {
	varStmt, ok := s.(*ast.VarStatement)
	if !ok {
		t.Errorf("s not *ast.VarStatement. got=%T", s)
		return false
	}

	if varStmt.Name.Value != name {
		t.Errorf("varStmt.Name.Value not '%s'. got=%s", name, varStmt.Name.Value)
		return false
	}

	if varStmt.Name.TokenLiteral() != name {
		t.Errorf("varStmt.Name.TokenLiteral() not '%s'. got=%s", name, varStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func TestIfStatement(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value != "x" {
		t.Errorf("consequence is not %s. got=%s", "x", stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value)
	}

	if stmt.Alternative != nil {
		t.Errorf("stmt.Alternative.Statements was not nil. got=%+v", stmt.Alternative)
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value != "x" {
		t.Errorf("consequence is not %s. got=%s", "x", stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value)
	}

	if stmt.Alternative == nil {
		t.Errorf("stmt.Alternative.Statements was nil")
	}
}

func TestExpressionStatement(t *testing.T) {
	input := `show("Hola")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	callExp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	ident, ok := callExp.Function.(*ast.Identifier)
	if !ok {
		t.Fatalf("callExp.Function is not ast.Identifier. got=%T",
			callExp.Function)
	}

	if ident.Value != "show" {
		t.Errorf("function name is not 'show'. got=%s", ident.Value)
	}

	if len(callExp.Arguments) != 1 {
		t.Fatalf("wrong number of arguments. got=%d", len(callExp.Arguments))
	}

	testLiteralExpression(t, callExp.Arguments[0], "Hola")
}

func TestWalrusStatements(t *testing.T) {
	input := `
edad := 25
NOMBRE := "Wilson"
PI STRING := "3.1416"
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	// Debug
	for i, stmt := range program.Statements {
		t.Logf("Statement %d: %T - %s", i, stmt, stmt.String())
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
		expectedConstant   bool
	}{
		{"edad", false},
		{"NOMBRE", true},
		{"PI", true},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testWalrusStatement(t, stmt, tt.expectedIdentifier, tt.expectedConstant) {
			return
		}
	}
}

func testWalrusStatement(t *testing.T, s ast.Statement, name string, isConstant bool) bool {
	varStmt, ok := s.(*ast.VarStatement)
	if !ok {
		t.Errorf("s not *ast.VarStatement. got=%T", s)
		return false
	}

	if varStmt.Name.Value != name {
		t.Errorf("varStmt.Name.Value not '%s'. got=%s", name, varStmt.Name.Value)
		return false
	}

	if varStmt.IsConstant != isConstant {
		t.Errorf("varStmt.IsConstant not %t. got=%t", isConstant, varStmt.IsConstant)
		return false
	}

	return true
}

func TestConstantReassignmentError(t *testing.T) {
	input := `
NOMBRE := "Wilson"
NOMBRE = "Pedro"
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// This should parse successfully, but evaluation should fail
	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}
}

func TestWalrusAssignInExpressionError(t *testing.T) {
	input := `x + := 5`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatalf("Expected parser error for ':=' in expression context, but got none")
	}

	expectedError := "':=' should not appear in expression context"
	found := false
	for _, err := range errors {
		if err == expectedError {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected error '%s', got %v", expectedError, errors)
	}
}

func TestDotInExpressionError(t *testing.T) {
	input := `x + .y`

	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatalf("Expected parser error for '.' in expression context, but got none")
	}

	expectedError := "no prefix parse function for DOT found"
	found := false
	for _, err := range errors {
		if err == expectedError {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected error '%s', got %v", expectedError, errors)
	}
}

func TestIfElifElseStatement(t *testing.T) {
	input := `if x { } elif y { } else { }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Alternative == nil {
		t.Errorf("stmt.Alternative should not be nil")
	}

	// The alternative should be a block with the elif statement
	if len(stmt.Alternative.Statements) != 1 {
		t.Fatalf("stmt.Alternative.Statements should contain 1 statement. got=%d",
			len(stmt.Alternative.Statements))
	}

	elifStmt, ok := stmt.Alternative.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("stmt.Alternative.Statements[0] is not ast.IfStatement. got=%T",
			stmt.Alternative.Statements[0])
	}

	if elifStmt.Alternative == nil {
		t.Errorf("elifStmt.Alternative should not be nil")
	}
}

// TestTypedVariableDeclaration tests the new typed variable syntax: identifier type := value
func TestTypedVariableDeclaration(t *testing.T) {
	input := `
a int := 10
b float := 3.14
c string := "Hola"
PI float := 3.14
YELL CONSTANT := "texto"
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 5 {
		t.Fatalf("program.Statements does not contain 5 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
		expectedType       string
		expectedConstant   bool
	}{
		{"a", "int", false},
		{"b", "float", false},
		{"c", "string", false},
		{"PI", "float", true},
		{"YELL", "CONSTANT", true},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testTypedVarStatement(t, stmt, tt.expectedIdentifier, tt.expectedType, tt.expectedConstant) {
			return
		}
	}
}

// testTypedVarStatement verifies that a typed variable declaration has the correct properties
func testTypedVarStatement(t *testing.T, s ast.Statement, name, expectedType string, expectedConstant bool) bool {
	varStmt, ok := s.(*ast.VarStatement)
	if !ok {
		t.Errorf("s not *ast.VarStatement. got=%T", s)
		return false
	}

	if varStmt.Name.Value != name {
		t.Errorf("varStmt.Name.Value not '%s'. got=%s", name, varStmt.Name.Value)
		return false
	}

	if varStmt.Name.TypeAnnotation != expectedType {
		t.Errorf("varStmt.Name.TypeAnnotation not '%s'. got=%s", expectedType, varStmt.Name.TypeAnnotation)
		return false
	}

	if varStmt.IsConstant != expectedConstant {
		t.Errorf("varStmt.IsConstant not %t. got=%t", expectedConstant, varStmt.IsConstant)
		return false
	}

	return true
}

// TestTypedFunctionParameters tests functions with typed parameters
func TestTypedFunctionParameters(t *testing.T) {
	input := `
void func miFuncion(x int, y string) {
    show.log("x vale:", x)
    show.log("y vale:", y)
}

func suma(a int, b int) {
    return a + b
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}

	// Test void function with typed parameters
	voidFuncStmt, ok := program.Statements[0].(*ast.FuncStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.FuncStatement. got=%T", program.Statements[0])
	}

	if voidFuncStmt.IsVoid != true {
		t.Errorf("voidFuncStmt.IsVoid should be true")
	}

	if len(voidFuncStmt.Parameters) != 2 {
		t.Fatalf("void function should have 2 parameters. got=%d", len(voidFuncStmt.Parameters))
	}

	expectedParams1 := []struct {
		name string
		typ  string
	}{
		{"x", "int"},
		{"y", "string"},
	}

	for i, param := range voidFuncStmt.Parameters {
		if param.Value != expectedParams1[i].name {
			t.Errorf("Parameter %d name not %s. got=%s", i, expectedParams1[i].name, param.Value)
		}
		if param.TypeAnnotation != expectedParams1[i].typ {
			t.Errorf("Parameter %d type not %s. got=%s", i, expectedParams1[i].typ, param.TypeAnnotation)
		}
	}

	// Test regular function with typed parameters and return type
	sumFuncStmt, ok := program.Statements[1].(*ast.FuncStatement)
	if !ok {
		t.Fatalf("program.Statements[1] is not ast.FuncStatement. got=%T", program.Statements[1])
	}

	if sumFuncStmt.IsVoid != false {
		t.Errorf("sumFuncStmt.IsVoid should be false")
	}

	expectedParams2 := []struct {
		name string
		typ  string
	}{
		{"a", "int"},
		{"b", "int"},
	}

	for i, param := range sumFuncStmt.Parameters {
		if param.Value != expectedParams2[i].name {
			t.Errorf("Parameter %d name not %s. got=%s", i, expectedParams2[i].name, param.Value)
		}
		if param.TypeAnnotation != expectedParams2[i].typ {
			t.Errorf("Parameter %d type not %s. got=%s", i, expectedParams2[i].typ, param.TypeAnnotation)
		}
	}
}
