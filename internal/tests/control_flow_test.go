// internal/tests/control_flow_test.go

package tests

import (
	"testing"
)

func TestBasicIfStatements(t *testing.T) {
	tests := []TestCase{
		{
			Name: "If statement true",
			Code: `
func main() {
    edad := 20
    if edad >= 18 {
        show.log("Mayor de edad")
    }
}`,
			ExpectedOutput: "Mayor de edad",
			ShouldCompile: true,
		},
		{
			Name: "If statement false - no output",
			Code: `
func main() {
    edad := 15
    if edad >= 18 {
        show.log("Mayor")
    }
}`,
			ExpectedOutput: "",
			ShouldCompile: true,
		},
		{
			Name: "If-else both branches",
			Code: `
func main() {
    edad := 15
    if edad >= 18 {
        show.log("Mayor")
    } else {
        show.log("Menor")
    }
}`,
			ExpectedOutput: "Menor",
			ShouldCompile: true,
		},
		{
			Name: "If-else true branch",
			Code: `
func main() {
    edad := 25
    if edad >= 18 {
        show.log("Mayor")
    } else {
        show.log("Menor")
    }
}`,
			ExpectedOutput: "Mayor",
			ShouldCompile: true,
		},
		{
			Name: "If-elif-else chain",
			Code: `
func main() {
    edad := 16
    if edad >= 18 {
        show.log("Mayor")
    } elif edad >= 16 {
        show.log("Adolescente")
    } else {
        show.log("Menor")
    }
}`,
			ExpectedOutput: "Adolescente",
			ShouldCompile: true,
		},
		{
			Name: "Multiple elif conditions",
			Code: `
func main() {
    score := 85
    if score >= 90 {
        show.log("A")
    } elif score >= 80 {
        show.log("B")
    } elif score >= 70 {
        show.log("C")
    } else {
        show.log("F")
    }
}`,
			ExpectedOutput: "B",
			ShouldCompile: true,
		},
		{
			Name: "Complex condition in if",
			Code: `
func main() {
    x := 5
    y := 10
    if x < y && x > 0 {
        show.log("Valid range")
    }
}`,
			ExpectedOutput: "Valid range",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestWhileLoops(t *testing.T) {
	tests := []TestCase{
		{
			Name: "While loop basic",
			Code: `
func main() {
    contador := 0
    while contador < 3 {
        show.log(contador)
        contador = contador + 1
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
		{
			Name: "While loop zero iterations",
			Code: `
func main() {
    contador := 5
    while contador < 3 {
        show.log(contador)
        contador = contador + 1
    }
}`,
			ExpectedOutput: "",
			ShouldCompile: true,
		},
		{
			Name: "While loop with complex condition",
			Code: `
func main() {
    x := 1
    while x < 10 && x % 3 != 0 {
        show.log(x)
        x = x + 1
    }
}`,
			ExpectedOutput: "1\n2",
			ShouldCompile: true,
		},
		{
			Name: "While loop modifying multiple variables",
			Code: `
func main() {
    a := 0
    b := 1
    while a < 3 {
        show.log(a + b)
        a = a + 1
        b = b * 2
    }
}`,
			ExpectedOutput: "1\n3\n5",
			ShouldCompile: true,
		},
		{
			Name: "While loop with boolean flag",
			Code: `
func main() {
    continuar := true
    count := 0
    while continuar {
        show.log(count)
        count = count + 1
        if count >= 3 {
            continuar = false
        }
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestBreakStatements(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Break in simple while loop",
			Code: `
func main() {
    contador := 0
    while contador < 10 {
        if contador == 3 {
            break
        }
        show.log(contador)
        contador = contador + 1
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
		{
			Name: "Break at start of loop",
			Code: `
func main() {
    x := 1
    while x > 0 {
        show.log("Start")
        break
        show.log("Never reached")
    }
    show.log("End")
}`,
			ExpectedOutput: "Start\nEnd",
			ShouldCompile: true,
		},
		{
			Name: "Break with complex condition",
			Code: `
func main() {
    i := 0
    while i < 10 {
        if i > 5 && i % 2 == 0 {
            break
        }
        show.log(i)
        i = i + 1
    }
}`,
			ExpectedOutput: "0\n1\n2\n3\n4\n5\n6",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestContinueStatements(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Continue skips iteration",
			Code: `
func main() {
    contador := 0
    while contador < 5 {
        contador = contador + 1
        if contador == 3 {
            continue
        }
        show.log(contador)
    }
}`,
			ExpectedOutput: "1\n2\n4\n5",
			ShouldCompile: true,
		},
		{
			Name: "Continue at start of loop",
			Code: `
func main() {
    i := 0
    while i < 5 {
        i = i + 1
        if i == 1 || i == 4 {
            continue
        }
        show.log(i)
    }
}`,
			ExpectedOutput: "2\n3\n5",
			ShouldCompile: true,
		},
		{
			Name: "Continue with complex condition",
			Code: `
func main() {
    x := 0
    while x < 8 {
        x = x + 1
        if (x % 2 == 0 || x % 3 == 0) && x != 6 {
            continue
        }
        if x == 6 {
            break
        }
        show.log(x)
    }
}`,
			ExpectedOutput: "1\n5",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestTraditionalForLoops(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Basic for loop",
			Code: `
func main() {
    for i := 0; i < 3; i = i + 1 {
        show.log("count:", i)
    }
}`,
			ExpectedOutput: "count: 0\ncount: 1\ncount: 2",
			ShouldCompile: true,
		},
		{
			Name: "For loop without init",
			Code: `
func main() {
    i := 0
    for ; i < 3; i = i + 1 {
        show.log(i)
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
		{
			Name: "For loop without post",
			Code: `
func main() {
    for i := 0; i < 3; {
        show.log(i)
        i = i + 1
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
		{
			Name: "For loop without init and post",
			Code: `
func main() {
    i := 0
    for ; i < 3; {
        show.log(i)
        i = i + 1
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
		{
			Name: "Nested for loops",
			Code: `
func main() {
    for i := 0; i < 2; i = i + 1 {
        for j := 0; j < 2; j = j + 1 {
            show.log("i:", i, "j:", j)
        }
    }
}`,
			ExpectedOutput: "i: 0 j: 0\ni: 0 j: 1\ni: 1 j: 0\ni: 1 j: 1",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestControlFlowErrors(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Break outside loop should fail",
			Code: `
func main() {
    break
}`,
			ShouldCompile: false,
			ExpectedError: "break",
		},
		{
			Name: "Continue outside loop should fail",
			Code: `
func main() {
    continue
}`,
			ShouldCompile: false,
			ExpectedError: "continue",
		},
		{
			Name: "Break in function without loop should fail",
			Code: `
func test() {
    break
}
func main() {
    test()
}`,
			ShouldCompile: false,
			ExpectedError: "break",
		},
		{
			Name: "Nested if break should work",
			Code: `
func main() {
    x := 0
    while x < 5 {
        if x == 3 {
            break
        }
        show.log(x)
        x = x + 1
    }
}`,
			ExpectedOutput: "0\n1\n2",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}
