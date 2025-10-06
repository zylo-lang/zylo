// internal/tests/edge_cases_test.go

package tests

import (
	"testing"
)

func TestArithmeticEdgeCases(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Division by zero",
			Code: `func main() { resultado := 10 / 0 }`,
			ShouldCompile: true, // Might be runtime error
		},
		{
			Name: "Modulo by zero",
			Code: `func main() { resultado := 10 % 0 }`,
			ShouldCompile: true, // Might be runtime error
		},
		{
			Name: "Very large integers",
			Code: `
func main() {
    large := 9223372036854775807
    show.log(large)
}`,
			ExpectedOutput: "9223372036854775807",
			ShouldCompile: true,
		},
		{
			Name: "Very small integers",
			Code: `
func main() {
    small := -9223372036854775808
    show.log(small)
}`,
			ExpectedOutput: "-9223372036854775808",
			ShouldCompile: true,
		},
		{
			Name: "Floating point precision",
			Code: `
func main() {
    result := 0.1 + 0.2
    show.log(result)
}`,
			ExpectedOutput: "0.3",
			ShouldCompile: true,
		},
		{
			Name: "Mixed type arithmetic",
			Code: `
func main() {
    a := 5
    b := 3.14
    c := a + b
    show.log(c)
}`,
			ExpectedOutput: "8.14",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestVariableEdgeCases(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Undefined variable",
			Code: `func main() { show.log(inexistente) }`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_002",
		},
		{
			Name: "Variable used before declaration",
			Code: `
func main() {
    show.log(x)
    x := 5
}`,
			ShouldCompile: false,
		},
		{
			Name: "Shadowing with different types",
			Code: `
func main() {
    x := 5
    x := "hello"
    show.log(x)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Very long variable names",
			Code: `
func main() {
    veryLongVariableNameThatShouldStillWork := 42
    anotherVeryLongVariableNameForTesting := "hello"
    show.log(veryLongVariableNameThatShouldStillWork, anotherVeryLongVariableNameForTesting)
}`,
			ExpectedOutput: "42 hello",
			ShouldCompile: true,
		},
		{
			Name: "Variables starting with numbers should fail",
			Code: `func main() { 1invalid := 5 }`,
			ShouldCompile: false,
		},
		{
			Name: "Empty identifier should fail",
			Code: `
func main() {
    var  := 5
}`,
			ShouldCompile: false,
		},
	}

	RunTestCases(t, tests)
}

func TestTypeMismatches(t *testing.T) {
	tests := []TestCase{
		{
			Name: "String + int should work or convert",
			Code: `
func main() {
    result := "hello" + 42
    show.log(result)
}`,
			ExpectedOutput: "hello42",
			ShouldCompile: true,
		},
		{
			Name: "Boolean to int conversion",
			Code: `
func main() {
    x := true + 1
    show.log(x)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Invalid type operation",
			Code: `func main() { result := "text" - 42 }`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_101",
		},
		{
			Name: "Float to int conversion in operation",
			Code: `
func main() {
    x := 5
    y := 3.14
    result := x + y
    show.log(result)
}`,
			ExpectedOutput: "8.14",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestCollectionEdgeCases(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Index out of range",
			Code: `
func main() {
    lista := [1, 2, 3]
    elemento := lista[10]
}`,
			ShouldCompile: true, // Runtime error
		},
		{
			Name: "Negative index out of range",
			Code: `
func main() {
    lista := [1, 2, 3]
    elemento := lista[-10]
}`,
			ShouldCompile: true, // Runtime error
		},
		{
			Name: "Empty collection access",
			Code: `
func main() {
    empty_list := []
    empty_map := {}
    show.log(empty_list[0], empty_map["key"])
}`,
			ShouldCompile: true,
		},
		{
			Name: "Nested out of bounds",
			Code: `
func main() {
    nested := [[1, 2], [3, 4]]
    item := nested[1][5]
    show.log(item)
}`,
			ShouldCompile: true,
		},
		{
			Name: "Map with complex keys",
			Code: `
func main() {
    key := [1, 2, 3]
    data := {key: "value"}
}`,
			ShouldCompile: false, // Non-string keys might not be allowed
		},
	}

	RunTestCases(t, tests)
}

func TestFunctionEdgeCases(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Function with too many parameters",
			Code: `
func add(a, b) {
    return a + b
}
func main() {
    result := add(1, 2, 3, 4, 5)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Function with too few parameters",
			Code: `
func add(a, b, c) {
    return a + b + c
}
func main() {
    result := add(1, 2)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Recursive function stack overflow risk",
			Code: `
func recursive(n) {
    return recursive(n + 1)
}
func main() {
    recursive(0)
}`,
			ShouldCompile: true, // Runtime issue
		},
		{
			Name: "Function returning wrong type",
			Code: `
func test() {
    return "string"
}
func main() {
    x := test()
    y := x + 5
}`,
			ShouldCompile: false,
		},
		{
			Name: "Void function with implicit return",
			Code: `
void func greet() {
    show.log("Hello")
}
func main() {
    result := greet()
    show.log(result)
}`,
			ShouldCompile: false,
		},
	}

	RunTestCases(t, tests)
}

func TestUnicodeAndEncoding(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Unicode strings",
			Code: `
func main() {
    texto := "Hola 世界 🌍"
    show.log(texto)
}`,
			ExpectedOutput: "Hola 世界 🌍",
			ShouldCompile: true,
		},
		{
			Name: "Unicode variable names",
			Code: `
func main() {
    café := "coffee"
    show.log(café)
}`,
			ExpectedOutput: "coffee",
			ShouldCompile: true,
		},
		{
			Name: "Emoji in strings",
			Code: `
func main() {
    message := "🔥 Hot! 🚀"
    show.log(message)
}`,
			ExpectedOutput: "🔥 Hot! 🚀",
			ShouldCompile: true,
		},
		{
			Name: "Unicode operators in comments",
			Code: `
func main() {
    // Comment with unicode: ± × ÷ ≠ ≤ ≥
    x := 5
    show.log(x)
}`,
			ExpectedOutput: "5",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestPrecedenceAndAssociativity(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Complex expression precedence",
			Code: `func main() { show.log(2 + 3 * 4 - 5 / 2) }`,
			ExpectedOutput: "11",
			ShouldCompile: true,
		},
		{
			Name: "Logical and comparison precedence",
			Code: `func main() { show.log(5 > 3 && 2 < 4 || true) }`,
			ExpectedOutput: "true",
			ShouldCompile: true,
		},
		{
			Name: "Mixed operators precedence",
			Code: `func main() { show.log(2 * 3 + 4 % 5) }`,
			ExpectedOutput: "7",
			ShouldCompile: true,
		},
		{
			Name: "Parentheses overriding precedence",
			Code: `func main() { show.log((2 + 3) * (4 - 1)) }`,
			ExpectedOutput: "15",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestEmptyAndMinimalPrograms(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Empty program",
			Code: ``,
			ShouldCompile: true,
			ExpectedOutput: "",
		},
		{
			Name: "Program with only whitespace",
			Code: `
			
			`,
			ShouldCompile: true,
			ExpectedOutput: "",
		},
		{
			Name: "Program with only comments",
			Code: `
// This is a comment
# This is also a comment
`,
			ShouldCompile: true,
			ExpectedOutput: "",
		},
		{
			Name: "Minimal program with main",
			Code: `func main() {}`,
			ShouldCompile: true,
			ExpectedOutput: "",
		},
		{
			Name: "Program with only variable declaration",
			Code: `
x := 42
func main() {
    show.log(x)
}`,
			ExpectedOutput: "42",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestDeepNesting(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Deeply nested expressions",
			Code: `func main() { show.log((((5 + 3) * 2) - 4) / 2) }`,
			ExpectedOutput: "7",
			ShouldCompile: true,
		},
		{
			Name: "Deeply nested function calls",
			Code: `
func a(x) { return x + 1 }
func b(x) { return x * 2 }
func c(x) { return x - 3 }
func d(x) { return x / 2 }
func main() {
    result := d(c(b(a(5))))
    show.log(result)
}`,
			ExpectedOutput: "4",
			ShouldCompile: true,
		},
		{
			Name: "Deep recursion call stack",
			Code: `
func countdown(n int) {
    if n <= 0 {
        return
    }
    countdown(n - 1)
}
func main() {
    countdown(100)
    show.log("Done")
}`,
			ExpectedOutput: "Done",
			ShouldCompile: true,
		},
		{
			Name: "Nested loops",
			Code: `
func main() {
    counter := 0
    while counter < 3 {
        inner := 0
        while inner < 2 {
            show.log(counter, inner)
            inner = inner + 1
        }
        counter = counter + 1
    }
}`,
			ExpectedOutput: "0 0\n0 1\n1 0\n1 1\n2 0\n2 1",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestLargePrograms(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Large array processing",
			Code: `
func main() {
    large := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20]
    sum := 0
    for num in large {
        sum = sum + num
    }
    show.log(sum)
}`,
			ExpectedOutput: "210",
			ShouldCompile: true,
		},
		{
			Name: "Many variables",
			Code: `
func main() {
    a1 := 1
    a2 := 2
    a3 := 3
    a4 := 4
    a5 := 5
    a6 := 6
    a7 := 7
    a8 := 8
    a9 := 9
    a10 := 10
    total := a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10
    show.log(total)
}`,
			ExpectedOutput: "55",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestWeirdSyntax(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Extreme spacing",
			Code: `
func
main
(

)
{
x
:
=
5

show
.
log
(
x
)
}
`,
			ExpectedOutput: "5",
			ShouldCompile: true,
		},
		{
			Name: "Mixed quote types in strings (should be invalid)",
			Code: `
func main() {
    x := 'single quotes'
    show.log(x)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Invalid escape sequences",
			Code: `
func main() {
    x := "invalid \z escape"
    show.log(x)
}`,
			ShouldCompile: false,
		},
	}

	RunTestCases(t, tests)
}
