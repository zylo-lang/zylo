// internal/tests/optimizations_test.go

package tests

import (
	"strings"
	"testing"
)

func TestConstantFolding(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Constant folding - simple addition",
			Code: `func main() { x := 5 + 3; show.log(x) }`,
			ExpectedOutput: "8",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				// Should directly assign 8 instead of computing 5 + 3
				return strings.Contains(goCode, "var x") && strings.Contains(goCode, "8")
			},
		},
		{
			Name: "Constant folding - complex arithmetic",
			Code: `func main() { x := 5 + 3 * 2 - 1; show.log(x) }`,
			ExpectedOutput: "10",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				return strings.Contains(goCode, "var x") && strings.Contains(goCode, "10")
			},
		},
		{
			Name: "Constant folding - with variables",
			Code: `func main() { x := 10; y := x + 5; show.log(y) }`,
			ExpectedOutput: "15",
			ShouldCompile: true,
		},
		{
			Name: "Boolean constant folding",
			Code: `func main() { x := true && false; show.log(x) }`,
			ExpectedOutput: "false",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				return strings.Contains(goCode, "false") && !strings.Contains(goCode, "true && false")
			},
		},
	}

	RunTestCases(t, tests)
}

func TestDeadCodeElimination(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Dead code after return",
			Code: `
func test() {
    return 5
    show.log("Never executed")
}`,
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				// Should not contain the "Never executed" print statement
				return !strings.Contains(goCode, `"Never executed"`)
			},
		},
		{
			Name: "Dead code in if false branch",
			Code: `
func main() {
    if true {
        show.log("Always")
    } else {
        show.log("Never")
    }
}`,
			ExpectedOutput: "Always",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				return !strings.Contains(goCode, `"Never"`)
			},
		},
		{
			Name: "Unreachable code after break",
			Code: `
func main() {
    while true {
        show.log("Once")
        break
        show.log("Never reached")
    }
}`,
			ExpectedOutput: "Once",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				// Should not execute code after break in loop
				return strings.Contains(goCode, "break") && !strings.Contains(goCode, `"Never reached"`)
			},
		},
	}

	RunTestCases(t, tests)
}

func TestNativeTypeOptimizations(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Native int arithmetic - no runtime wrapper",
			Code: `
func main() {
    a int := 10
    b int := 20
    c := a + b
    show.log(c)
}`,
			ExpectedOutput: "30",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				// Should use native Go operations, not zyloruntime functions
				return !strings.Contains(goCode, "zyloruntime.Add") &&
					   strings.Contains(goCode, "a + b")
			},
		},
		{
			Name: "Native float arithmetic",
			Code: `
func main() {
    a float := 3.5
    b float := 2.1
    c := a * b
    show.log(c)
}`,
			ExpectedOutput: "7.35",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				return strings.Contains(goCode, "float64") && !strings.Contains(goCode, "zyloruntime.Multiply")
			},
		},
		{
			Name: "Mixed type operations",
			Code: `
func main() {
    x int := 5
    y float := 3.0
    result := x + y
    show.log(result)
}`,
			ExpectedOutput: "8",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestIfOptimization(t *testing.T) {
	tests := []TestCase{
		{
			Name: "If true optimization",
			Code: `
func main() {
    if true {
        show.log("Always")
    } else {
        show.log("Never")
    }
}`,
			ExpectedOutput: "Always",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				// Should eliminate the if and just execute the consequence
				return !strings.Contains(goCode, `"Never"`)
			},
		},
		{
			Name: "If false optimization",
			Code: `
func main() {
    if false {
        show.log("Never")
    } else {
        show.log("Always")
    }
}`,
			ExpectedOutput: "Always",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				return !strings.Contains(goCode, `"Never"`)
			},
		},
		{
			Name: "Constant condition in while",
			Code: `
func main() {
    while false {
        show.log("Never executed")
    }
    show.log("Done")
}`,
			ExpectedOutput: "Done",
			ShouldCompile: true,
			ValidateGeneratedCode: func(goCode string) bool {
				// Should eliminate the entire while loop since condition is false
				return !strings.Contains(goCode, `"Never executed"`)
			},
		},
	}

	RunTestCases(t, tests)
}

func TestVariablePropagation(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Simple variable propagation",
			Code: `
func main() {
    x := 5
    y := x + 3
    show.log(y)
}`,
			ExpectedOutput: "8",
			ShouldCompile: true,
		},
		{
			Name: "Multiple variable uses",
			Code: `
func main() {
    a := 10
    b := a * 2
    c := b + 5
    d := c / 3
    show.log(d)
}`,
			ExpectedOutput: "8",
			ShouldCompile: true,
		},
		{
			Name: "Variable reassignment",
			Code: `
func main() {
    x := 5
    x := x + 1
    show.log(x)
}`,
			ExpectedOutput: "6",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestLoopOptimizations(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Empty loop optimization",
			Code: `
func main() {
    counter := 0
    while counter < 1000000 {
        counter = counter + 1
    }
    show.log("Done after", counter, "iterations")
}`,
			ExpectedOutput: "Done after 1000000 iterations",
			ShouldCompile: true,
		},
		{
			Name: "Loop invariant code motion simulation",
			Code: `
func main() {
    result := 0
    for item in [1, 2, 3, 4, 5] {
        multiplier := 10  // This could be hoisted in advanced optimization
        result = result + item * multiplier
    }
    show.log(result)
}`,
			ExpectedOutput: "150",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestOptimizationValidation(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Optimizations don't change program behavior",
			Code: `
func main() {
    const := 5 + 3  // Should be folded to 8
    variable := const * 2  // Should be folded to 16
    result := variable - 1  // Should be folded to 15
    show.log(result)
}`,
			ExpectedOutput: "15",
			ShouldCompile: true,
		},
		{
			Name: "Maintain execution order in optimizations",
			Code: `
func side_effect() {
    show.log("executed")
    return 42
}
func main() {
    x := side_effect() + 8  // side_effect should still execute
    show.log("result:", x)
}`,
			ExpectedOutput: "executed\nresult: 50",
			ShouldCompile: true,
		},
		{
			Name: "Don't optimize away necessary operations",
			Code: `
func main() {
    a := 1
    b := 2
    c := a + b
    d := c * 3
    show.log(d)
	}`,
			ExpectedOutput: "9",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}
