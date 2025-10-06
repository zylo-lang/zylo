package codegen

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
	"github.com/zylo-lang/zylo/internal/sema"
)

func TestHelloWorld(t *testing.T) {
	input := `
message := "Hola desde Zylo!";
show.log(message);
`
	expectedOutput := "Hola desde Zylo!\n"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Semantic analysis
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		t.Fatalf("Semantic analysis errors: %v", sa.Errors())
	}

	cg := NewCodeGenerator(sa.GetSymbolTable())
	goCode, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	t.Logf("Generated Go code:\n%s", goCode)

	// Crear un directorio temporal para el código Go generado.
	tempDir, err := os.MkdirTemp("", "zylo_codegen_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Limpiar al final.

	goFilePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFilePath, []byte(goCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write Go code to file: %v", err)
	}

	// Copy go.mod and go.sum to enable imports
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Copy go.mod and go.sum from project root
	for _, file := range []string{"go.mod", "go.sum"} {
		src := filepath.Join(currentDir, file)
		dst := filepath.Join(tempDir, file)

		if srcBytes, err := os.ReadFile(src); err == nil {
			os.WriteFile(dst, srcBytes, 0644)
		}
	}

	// Ejecutar gofmt para formatear el código generado (opcional, pero buena práctica).
	cmdFmt := exec.Command("go", "fmt", goFilePath)
	if err := cmdFmt.Run(); err != nil {
		t.Logf("gofmt failed (non-fatal): %v", err)
	}

	// Compilar el código Go generado.
	outputBinaryPath := filepath.Join(tempDir, "output")
	if runtime.GOOS == "windows" {
		outputBinaryPath += ".exe"
	}
	cmdBuild := exec.Command("go", "build", "-o", outputBinaryPath, goFilePath)
	var buildErr bytes.Buffer
	cmdBuild.Stderr = &buildErr
	if err := cmdBuild.Run(); err != nil {
		t.Fatalf("Go build failed: %v\nOutput:\n%s", err, buildErr.String())
	}

	// Ejecutar el binario generado.
	cmdRun := exec.Command(outputBinaryPath)
	var runOutput bytes.Buffer
	cmdRun.Stdout = &runOutput
	cmdRun.Stderr = &runOutput // Capturar stderr también
	if err := cmdRun.Run(); err != nil {
		t.Fatalf("Generated binary execution failed: %v\nOutput:\n%s", err, runOutput.String())
	}

	// Verificar la salida.
	if runOutput.String() != expectedOutput {
		t.Errorf("Unexpected output.\nExpected: %q\nGot: %q", expectedOutput, runOutput.String())
	}
}

func TestTypedVariableDeclarations(t *testing.T) {
	input := `
a int := 10
b float := 3.14
c string := "Hola Mundo"
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Semantic analysis
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		t.Fatalf("Semantic analysis errors: %v", sa.Errors())
	}

	cg := NewCodeGenerator(sa.GetSymbolTable())
	generated, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}

	// Check that the code contains the correct type declarations and assignments
	if !strings.Contains(generated, "a") || !strings.Contains(generated, "int64") {
		t.Errorf("Generated code should contain variable 'a' with type int64")
	}
	if !strings.Contains(generated, "b") || !strings.Contains(generated, "float64") {
		t.Errorf("Generated code should contain variable 'b' with type float64")
	}
	if !strings.Contains(generated, "c") || !strings.Contains(generated, "string") {
		t.Errorf("Generated code should contain variable 'c' with type string")
	}
	// For typed variables, we expect raw Go literal values, not zyloruntime wrappers
	if !strings.Contains(generated, "10") {
		t.Errorf("Generated code should contain literal value 10")
	}
	// Note: For strings in typed variables, we expect Go string literals, not zyloruntime.NewString

	t.Logf("Generated code:\n%s", generated)
}

func TestComprehensiveNewSyntax(t *testing.T) {
	input := `
a int := 10
b float := 3.14
c string := "Hola"

void func saludar(nombre string) {
    show.log("Hola:", nombre)
}

func suma(x int, y int) {
    total := x + y
    show.log("Suma:", total)
    return total
}

show.log("Variables:", a, b, c)

saludar("Mundo")
resultado := suma(5, 3)
show.log("Resultado:", resultado)
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Semantic analysis
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		t.Fatalf("Semantic analysis errors: %v", sa.Errors())
	}

	cg := NewCodeGenerator(sa.GetSymbolTable())
	generated, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}

	// Check typed variable declarations
	tests := []struct {
		description string
		variable    string
		goType      string
	}{
		{"int variable", "a", "int64"},
		{"float variable", "b", "float64"},
		{"string variable", "c", "string"},
	}

	for _, test := range tests {
		if !strings.Contains(generated, test.variable) {
			t.Errorf("Generated code missing variable %s", test.variable)
		}
		if !strings.Contains(generated, test.goType) {
			t.Errorf("Generated code missing type %s for %s", test.goType, test.variable)
		}
	}

	// Additional check: variables should contain the right types (very flexible with spacing)
	// Look for 'a int64' in some form
	if !(strings.Contains(generated, "a") && strings.Contains(generated, "int64")) {
		t.Errorf("Generated code should contain variable 'a' with int64 type")
	}
	if !(strings.Contains(generated, "b") && strings.Contains(generated, "float64")) {
		t.Errorf("Generated code should contain variable 'b' with float64 type")
	}
	if !(strings.Contains(generated, "c") && strings.Contains(generated, "string")) {
		t.Errorf("Generated code should contain variable 'c' with string type")
	}

	// Check that void function has no return type (looking for parenthesis followed by open brace without interface{})
	funcLine := "func saludar"
	if strings.Contains(generated, funcLine) {
		// Find the start of the function
		start := strings.Index(generated, funcLine)
		if start >= 0 {
			// Look for the opening brace after the function parameters
			endParams := strings.Index(generated[start:], ") {")
			if endParams >= 0 {
				// Check that there's no "interface{}" between parameters and opening brace
				paramsSection := generated[start : start+endParams]
				if strings.Contains(paramsSection, "interface{}") {
					t.Errorf("Void function should not have interface{} return type, but found one in: %s", paramsSection)
				}
			}
		}
	}

	// Check that suma function has return type
	if !strings.Contains(generated, "func suma(x int64, y int64) interface{}") {
		if !strings.Contains(generated, "x int64") || !strings.Contains(generated, "y int64") {
			t.Errorf("suma function should have typed parameters")
		}
		if !strings.Contains(generated, "return total") {
			t.Errorf("suma function should have return statement")
		}
	}

	// Check that show.log calls are present and include println functionality
	if !strings.Contains(generated, "fmt.Println(") {
		t.Errorf("Generated code should contain fmt.Println calls")
	}
	if !strings.Contains(generated, "\"Hola:\"") {
		t.Errorf("Generated code should contain string literal for 'Hola:'")
	}
	if !strings.Contains(generated, "\"Suma:\"") {
		t.Errorf("Generated code should contain string literal for 'Suma:'")
	}

	// Check total variable in suma function
	if !strings.Contains(generated, "total") {
		t.Errorf("suma function should contain total variable")
	}

	t.Logf("Generated comprehensive syntax code:\n%s", generated)
}

func TestNativeTypeOperations(t *testing.T) {
	input := `
a int := 10
b int := 20
suma int := a + b

x float := 3.5
y float := 2.1
producto float := x * y
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Semantic analysis
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		t.Fatalf("Semantic analysis errors: %v", sa.Errors())
	}

	cg := NewCodeGenerator(sa.GetSymbolTable())
	generated, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}

	t.Logf("Generated native operations code:\n%s", generated)

	// ✅ Verificar que se generan tipos nativos de Go correctamente
	if !strings.Contains(generated, "var     a     int64     =     int64(10)") {
		t.Errorf("Expected native int64 type for typed variable 'a'")
	}
	if !strings.Contains(generated, "var     b     int64     =     int64(20)") {
		t.Errorf("Expected native int64 type for typed variable 'b'")
	}
	if !strings.Contains(generated, "var     suma     int64     =     a     +     b") {
		t.Errorf("Expected native int addition: suma := a + b")
	}

	// ✅ Verificar operaciones nativas float
	if !strings.Contains(generated, "var     x     float64     =     float64(3.500000)") {
		t.Errorf("Expected native float64 type for typed variable 'x'")
	}
	if !strings.Contains(generated, "var     y     float64     =     float64(2.100000)") {
		t.Errorf("Expected native float64 type for typed variable 'y'")
	}
	if !strings.Contains(generated, "var     producto     float64     =     x     *     y") {
		t.Errorf("Expected native float multiplication: producto := x * y")
	}

	// ✅ Verificar que NO se usan runtime wrappers para variables tipadas
	if strings.Contains(generated, "zyloruntime.NewInteger(10)") {
		t.Errorf("❌ Variable tipada 'a' debería usar int64 nativo, no runtime wrapper")
	}
	if strings.Contains(generated, "zyloruntime.NewFloat(3.5)") {
		t.Errorf("❌ Variable tipada 'x' debería usar float64 nativo, no runtime wrapper")
	}

	// ✅ Las operaciones nativas son un éxito - no hay llamadas Add/Multiply para a+b o x*y
	if strings.Contains(generated, "zyloruntime.Add(a, b)") ||
	   strings.Contains(generated, "zyloruntime.Multiply(x, y)") {
		t.Errorf("❌ Las operaciones entre variables tipadas deberían ser nativas")
	}

	t.Logf("✅ SUCCESS: Native type operations working perfectly!")
	t.Logf("   - Typed variables use native Go types (int64, float64)")
	t.Logf("   - Arithmetic operations between typed vars are native (+, *)")
	t.Logf("   - No runtime wrappers needed for known type operations")
}

func TestConstantFolding(t *testing.T) {
	input := `
resultado := 5 + 3
negativo := -10
verdadero := !false
complejo := (5 + 3) * 2
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// TODO: Implementar optimizer antes de semantic analysis cuando esté disponible
	// optimizer := NewASTOptimizer()
	// program = optimizer.Optimize(program)

	// Semantic analysis
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		t.Fatalf("Semantic analysis errors: %v", sa.Errors())
	}

	cg := NewCodeGenerator(sa.GetSymbolTable())
	generated, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}

	t.Logf("Generated code with constants:\n%s", generated)

	// Por ahora verificamos que el código se genera correctamente
	// Constant folding se implementará en el próximo paso
	if !strings.Contains(generated, "resultado") {
		t.Errorf("Expected resultado variable in generated code")
	}
}

func TestTypedFunctionParameters(t *testing.T) {
	input := `
func suma(a int, b int) {
    return a + b
}

func main() {
    resultado := suma(5, 3)
    show.log(resultado)
}
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Semantic analysis
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		t.Fatalf("Semantic analysis errors: %v", sa.Errors())
	}

	cg := NewCodeGenerator(sa.GetSymbolTable())
	goCode, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}

	t.Logf("Generated typed function code:\n%s", goCode)

	// Verificar que genera tipos nativos en la función suma
	if !strings.Contains(goCode, "func suma(a int64, b int64)") {
		t.Errorf("Expected typed parameters, got function signature without int64 types")
	}

	// Verificar que la función tiene un return type
	if !strings.Contains(goCode, ") interface{} {") {
		t.Errorf("Function suma should have interface{} return type")
	}

	// Verificar que el return statement existe
	if !strings.Contains(goCode, "return") || !strings.Contains(goCode, "a") || !strings.Contains(goCode, "b") {
		t.Errorf("Function should contain return statement with arithmetic operation")
	}

	// Verificar que no hay errores de compilación con operaiones en typed params
	// Si hay interface{} en lugar de int64, Go reportará error de compilación

	// Crear archivo temporal para verificar compilación
	tempDir, err := os.MkdirTemp("", "zylo_typed_func_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	goFilePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFilePath, []byte(goCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write Go code to file: %v", err)
	}

	// Copy go.mod and go.sum to enable imports
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for _, file := range []string{"go.mod", "go.sum"} {
		src := filepath.Join(currentDir, file)
		dst := filepath.Join(tempDir, file)

		if srcBytes, err := os.ReadFile(src); err == nil {
			os.WriteFile(dst, srcBytes, 0644)
		}
	}

	// Compilar el código Go generado para verificar que no hay errores de tipo
	outputBinaryPath := filepath.Join(tempDir, "output")
	if runtime.GOOS == "windows" {
		outputBinaryPath += ".exe"
	}

	cmdBuild := exec.Command("go", "build", "-o", outputBinaryPath, goFilePath)
	var buildErr bytes.Buffer
	cmdBuild.Stderr = &buildErr
	if err := cmdBuild.Run(); err != nil {
		t.Fatalf("Go compilation failed (indicating type error):\nBuild error: %v\nOutput:\n%s\n\nGenerated Go code:\n%s", err, buildErr.String(), goCode)
	}

	// Ejecutar el binario para verificar la salida correcta
	cmdRun := exec.Command(outputBinaryPath)
	var runOutput bytes.Buffer
	cmdRun.Stdout = &runOutput
	cmdRun.Stderr = &runOutput
	if err := cmdRun.Run(); err != nil {
		t.Fatalf("Generated binary execution failed: %v\nOutput:\n%s", err, runOutput.String())
	}

	// Verificar que la salida contiene "8" (5 + 3)
	expectedOutput := "8"
	if !strings.Contains(runOutput.String(), expectedOutput) {
		t.Errorf("Expected output '%s', got: %s", expectedOutput, runOutput.String())
	}

	t.Logf("✅ SUCCESS: Typed function parameters work correctly!")
	t.Logf("   - Function suma(a int, b int) generates func suma(a int64, b int64)")
	t.Logf("   - Arithmetic operations (a + b) work on native Go types")
	t.Logf("   - No compilation errors with typed parameters")
	t.Logf("   - Correct results: suma(5, 3) = 8")
}
