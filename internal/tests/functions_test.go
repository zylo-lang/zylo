// internal/tests/functions_test.go

package tests

import (
	"testing"
)

func TestBasicFunctions(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Function with return value",
			Code: `
func sumar(a int, b int) {
    return a + b
}
func main() {
    resultado := sumar(5, 3)
    show.log(resultado)
}`,
			ExpectedOutput: "8",
			ShouldCompile: true,
		},
		{
			Name: "Function with type inference",
			Code: `
func multiplicar(x, y) {
    return x * y
}
func main() {
    show.log(multiplicar(4, 5))
}`,
			ExpectedOutput: "20",
			ShouldCompile: true,
		},
		{
			Name: "Void function",
			Code: `
void func saludar(nombre string) {
    show.log("Hola", nombre)
}
func main() {
    saludar("Wilson")
}`,
			ExpectedOutput: "Hola Wilson",
			ShouldCompile: true,
		},
		{
			Name: "Function with multiple parameters",
			Code: `
func calcular(a int, b int, c int) {
    return a + b * c
}
func main() {
    show.log(calcular(2, 3, 4))
}`,
			ExpectedOutput: "14",
			ShouldCompile: true,
		},
		{
			Name: "Function with mixed types",
			Code: `
func info(nombre string, edad int, activo bool) {
    return nombre + " (" + string(edad) + ") " + string(activo)
}
func main() {
    show.log(info("Ana", 25, true))
}`,
			ExpectedOutput: "Ana (25) true",
			ShouldCompile: true,
		},
		{
			Name: "Function returning different types",
			Code: `
func getValue(tipo string) {
    if tipo == "int" {
        return 42
    } elif tipo == "float" {
        return 3.14
    } elif tipo == "string" {
        return "hello"
    } else {
        return true
    }
}
func main() {
    show.log(getValue("int"), getValue("float"), getValue("string"), getValue("bool"))
}`,
			ExpectedOutput: "42 3.14 hello true",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestHigherOrderFunctions(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Function taking function parameter",
			Code: `
func aplicar(x int, f func) {
    return f(x)
}
func duplicar(n) {
    return n * 2
}
func main() {
    resultado := aplicar(5, duplicar)
    show.log(resultado)
}`,
			ExpectedOutput: "10",
			ShouldCompile: true,
		},
		{
			Name: "Anonymous function",
			Code: `
func main() {
    sumar := func(x, y) { return x + y }
    show.log(sumar(3, 4))
}`,
			ExpectedOutput: "7",
			ShouldCompile: true,
		},
		{
			Name: "Function returning function",
			Code: `
func crear_multiplicador(factor int) {
    func multiplicar(x) {
        return x * factor
    }
    return multiplicar
}
func main() {
    duplicar := crear_multiplicador(2)
    show.log(duplicar(7))
}`,
			ExpectedOutput: "14",
			ShouldCompile: true,
		},
		{
			Name: "Chain of higher-order functions",
			Code: `
func aplicar_operacion(x int, op func) {
    return op(x)
}
func sumar_uno(n) { return n + 1 }
func multiplicar_por_dos(n) { return n * 2 }
func main() {
    resultado := aplicar_operacion(5, sumar_uno)
    resultado = aplicar_operacion(resultado, multiplicar_por_dos)
    show.log(resultado)
}`,
			ExpectedOutput: "12",
			ShouldCompile: true,
		},
		{
			Name: "Anonymous function with closure",
			Code: `
func main() {
    factor := 3
    multiplicar_por_factor := func(x) { return x * factor }
    show.log(multiplicar_por_factor(4))
}`,
			ExpectedOutput: "12",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestRecursion(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Simple recursion - factorial",
			Code: `
func factorial(n int) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
func main() {
    show.log(factorial(5))
}`,
			ExpectedOutput: "120",
			ShouldCompile: true,
		},
		{
			Name: "Fibonacci recursion",
			Code: `
func fib(n int) {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}
func main() {
    show.log(fib(7))
}`,
			ExpectedOutput: "13",
			ShouldCompile: true,
		},
		{
			Name: "Deep recursion",
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
			Name: "Mutual recursion",
			Code: `
func es_par(n int) {
    if n == 0 {
        return true
    }
    return es_impar(n - 1)
}
func es_impar(n int) {
    if n == 0 {
        return false
    }
    return es_par(n - 1)
}
func main() {
    show.log("4 es par:", es_par(4))
    show.log("5 es impar:", es_impar(5))
}`,
			ExpectedOutput: "4 es par: true\n5 es impar: true",
			ShouldCompile: true,
		},
		{
			Name: "Tail recursion (concept)",
			Code: `
func factorial_tail(n int, acc int) {
    if n <= 1 {
        return acc
    }
    return factorial_tail(n - 1, n * acc)
}
func factorial_wrapper(n int) {
    return factorial_tail(n, 1)
}
func main() {
    show.log(factorial_wrapper(6))
}`,
			ExpectedOutput: "720",
			ShouldCompile: true,
		},
	}
	RunTestCases(t, tests)
}

func TestFunctionErrors(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Return in void function should fail",
			Code: `
void func test() {
    return 5
}`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_107",
		},
		{
			Name: "Void function with return statement should fail",
			Code: `
void func greet() {
    show.log("Hello")
    return "done"
}`,
			ShouldCompile: false,
			ExpectedError: "ZYLO_ERR_107",
		},
		{
			Name: "Function call with wrong number of arguments",
			Code: `
func add(a int, b int) {
    return a + b
}
func main() {
    result := add(5)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Function call with wrong argument types",
			Code: `
func add(a int, b int) {
    return a + b
}
func main() {
    result := add("hello", 5)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Undefined function call",
			Code: `
func main() {
    result := undefinedFunction(5)
}`,
			ShouldCompile: false,
		},
		{
			Name: "Recursive function without base case",
			Code: `
func infinite_loop(n int) {
    return infinite_loop(n) + 1
}
func main() {
    infinite_loop(1)
}`,
			ShouldCompile: true, // Compile-time detection might be limited
		},
	}

	RunTestCases(t, tests)
}

func TestFunctionScopes(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Variable scope in function",
			Code: `
func test() {
    x := 10
    return x
}
func main() {
    y := 5
    result := test()
    show.log(result + y)
}`,
			ExpectedOutput: "15",
			ShouldCompile: true,
		},
		{
			Name: "Global and local variables",
			Code: `
global := 100
func getGlobal() {
    return global
}
func setLocal() {
    local := 50
    return local
}
func main() {
    show.log(getGlobal())
    show.log(setLocal())
}`,
			ExpectedOutput: "100\n50",
			ShouldCompile: true,
		},
		{
			Name: "Function parameters don't affect global scope",
			Code: `
x := 1
func test(x int) {
    x = x + 10
    return x
}
func main() {
    result := test(5)
    show.log("result:", result, "global x:", x)
}`,
			ExpectedOutput: "result: 15 global x: 1",
			ShouldCompile: true,
		},
		{
			Name: "Nested function scope",
			Code: `
func outer() {
    outer_var := 100
    func inner() {
        inner_var := 50
        return outer_var + inner_var
    }
    return inner()
}
func main() {
    show.log(outer())
}`,
			ExpectedOutput: "150",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}
