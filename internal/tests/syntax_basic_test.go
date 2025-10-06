// internal/tests/syntax_basic_test.go

package tests

import (
	"testing"
)

func TestVariableDeclarations(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Typed variable int",
			Code: `
func main() {
    edad int := 25
    show.log(edad)
}`,
			ExpectedOutput: "25",
			ShouldCompile: true,
		},
		{
			Name: "Typed variable float",
			Code: `
func main() {
    altura float := 1.75
    show.log(altura)
}`,
			ExpectedOutput: "1.75",
			ShouldCompile: true,
		},
		{
			Name: "Typed variable string",
			Code: `
func main() {
    mensaje string := "Hola Mundo"
    show.log(mensaje)
}`,
			ExpectedOutput: "Hola Mundo",
			ShouldCompile: true,
		},
		{
			Name: "Typed variable bool",
			Code: `
func main() {
    activo bool := true
    show.log(activo)
}`,
			ExpectedOutput: "true",
			ShouldCompile: true,
		},
		{
			Name: "Variable with type inference",
			Code: `
func main() {
    x := 42
    y := 3.14
    z := "texto"
    w := true
    show.log(x, y, z, w)
}`,
			ExpectedOutput: "42 3.14 texto true",
			ShouldCompile: true,
		},
		{
			Name: "Variable without initialization should fail",
			Code: `
func main() {
    var contador int
    show.log(contador)
}`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_003",
		},
		{
			Name: "UPPERCASE constant",
			Code: `
PI float := 3.14159
func main() {
    show.log(PI)
}`,
			ExpectedOutput: "3.14159",
			ShouldCompile: true,
		},
		{
			Name: "Constant reassignment should fail",
			Code: `
PI := 3.14
func main() {
    PI = 3.14159
}`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_203",
		},
		{
			Name: "Uppercase type should fail",
			Code: `
func main() {
    edad Int := 25
}`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_011",
		},
		{
			Name: "Mixed type operations int + float",
			Code: `
func main() {
    resultado := 5 + 3.14
    show.log(resultado)
}`,
			ExpectedOutput: "8.14",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestArithmeticOperators(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Addition",
			Code: `func main() { show.log(10 + 5) }`,
			ExpectedOutput: "15",
		},
		{
			Name: "Subtraction",
			Code: `func main() { show.log(10 - 5) }`,
			ExpectedOutput: "5",
		},
		{
			Name: "Multiplication",
			Code: `func main() { show.log(10 * 5) }`,
			ExpectedOutput: "50",
		},
		{
			Name: "Division returns float",
			Code: `func main() { show.log(10 / 3) }`,
			ExpectedOutput: "3.333",
		},
		{
			Name: "Modulo",
			Code: `func main() { show.log(10 % 3) }`,
			ExpectedOutput: "1",
		},
		{
			Name: "Complex expression with precedence",
			Code: `func main() { show.log(5 + 3 * 2) }`,
			ExpectedOutput: "11",
		},
		{
			Name: "Parentheses override precedence",
			Code: `func main() { show.log((5 + 3) * 2) }`,
			ExpectedOutput: "16",
		},
		{
			Name: "Negative numbers",
			Code: `func main() { show.log(-5 + 3) }`,
			ExpectedOutput: "-2",
		},
		{
			Name: "Complex arithmetic chain",
			Code: `func main() { show.log(2 + 3 * 4 - 6 / 2) }`,
			ExpectedOutput: "11",
		},
		{
			Name: "Arithmetic with variables",
			Code: `
func main() {
    a := 10
    b := 3
    show.log(a + b * 2)
}`,
			ExpectedOutput: "16",
		},
	}

	RunTestCases(t, tests)
}

func TestComparisonOperators(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Equal",
			Code: `func main() { show.log(5 == 5) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Not equal",
			Code: `func main() { show.log(5 != 3) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Greater than",
			Code: `func main() { show.log(7 > 3) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Less than",
			Code: `func main() { show.log(3 < 7) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Greater or equal",
			Code: `func main() { show.log(7 >= 7) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Less or equal",
			Code: `func main() { show.log(5 <= 5) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Equal false",
			Code: `func main() { show.log(5 == 6) }`,
			ExpectedOutput: "false",
		},
		{
			Name: "Not equal false",
			Code: `func main() { show.log(5 != 5) }`,
			ExpectedOutput: "false",
		},
		{
			Name: "Greater than false",
			Code: `func main() { show.log(3 > 7) }`,
			ExpectedOutput: "false",
		},
		{
			Name: "Less than false",
			Code: `func main() { show.log(7 < 3) }`,
			ExpectedOutput: "false",
		},
	}

	RunTestCases(t, tests)
}

func TestLogicalOperators(t *testing.T) {
	tests := []TestCase{
		{
			Name: "AND operator",
			Code: `func main() { show.log(true && false) }`,
			ExpectedOutput: "false",
		},
		{
			Name: "OR operator",
			Code: `func main() { show.log(true || false) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "NOT operator",
			Code: `func main() { show.log(!true) }`,
			ExpectedOutput: "false",
		},
		{
			Name: "NOT keyword",
			Code: `func main() { show.log(not true) }`,
			ExpectedOutput: "false",
		},
		{
			Name: "Short-circuit AND true",
			Code: `
func expensive() {
    show.log("executed")
    return true
}
func main() {
    resultado := true && expensive()
    show.log(resultado)
}`,
			ExpectedOutput: "executed\ntrue",
		},
		{
			Name: "Short-circuit AND false",
			Code: `
func expensive() {
    show.log("executed")
    return true
}
func main() {
    resultado := false && expensive()
    show.log(resultado)
}`,
			ExpectedOutput: "false",
			ShouldNotContain: "executed",
		},
		{
			Name: "Short-circuit OR true",
			Code: `
func expensive() {
    show.log("executed")
    return true
}
func main() {
    resultado := true || expensive()
    show.log(resultado)
}`,
			ExpectedOutput: "true",
			ShouldNotContain: "executed",
		},
		{
			Name: "Short-circuit OR false",
			Code: `
func expensive() {
    show.log("executed")
    return false
}
func main() {
    resultado := false || expensive()
    show.log(resultado)
}`,
			ExpectedOutput: "executed\nfalse",
		},
		{
			Name: "Complex logical expression",
			Code: `func main() { show.log((5 > 3) && (2 < 4) || false) }`,
			ExpectedOutput: "true",
		},
		{
			Name: "Logical with variables",
			Code: `
func main() {
    a := true
    b := false
    show.log(a && !b)
}`,
			ExpectedOutput: "true",
		},
	}

	RunTestCases(t, tests)
}
