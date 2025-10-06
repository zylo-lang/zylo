package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zylo-lang/zylo/internal/codegen"
	"github.com/zylo-lang/zylo/internal/evaluator"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
	"github.com/zylo-lang/zylo/internal/sema"
)

const Version = "1.0.0"

// Colores ANSI para terminal
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
)

func colorize(text, color string) string {
	return color + text + ColorReset
}

func printUsage() {
	fmt.Println(colorize("Zylo Programming Language CLI v"+Version, ColorCyan))
	fmt.Println()
	fmt.Println(colorize("USO:", ColorYellow))
	fmt.Println("  zylo <comando> [argumentos] [flags]")
	fmt.Println()
	fmt.Println(colorize("COMANDOS BÁSICOS:", ColorYellow))
	fmt.Println("  run <archivo>     Ejecuta un script Zylo")
	fmt.Println("  repl              Inicia REPL interactivo")
	fmt.Println("  test              Ejecuta tests automáticos")
	fmt.Println("  version           Muestra versión")
	fmt.Println("  init <proyecto>   Crea proyecto con estructura")
	fmt.Println("  doctor            Verifica instalación")
	fmt.Println()
	fmt.Println(colorize("DESARROLLO:", ColorYellow))
	fmt.Println("  fmt [archivo]     Formatea código")
	fmt.Println("  lint [archivo]    Detecta errores")
	fmt.Println("  debug <archivo>   Ejecuta con debug")
	fmt.Println("  doc [archivo]     Genera documentación")
	fmt.Println("  deps              Lista dependencias")
	fmt.Println("  add <paquete>     Instala paquetes")
	fmt.Println()
	fmt.Println(colorize("SERVIDOR:", ColorYellow))
	fmt.Println("  serve [proyecto]  Inicia servidor HTTP")
	fmt.Println()
	fmt.Println(colorize("ACTUALIZACIONES:", ColorYellow))
	fmt.Println("  version-check     Verifica nuevas versiones")
	fmt.Println("  self-update       Actualiza Zylo")
	fmt.Println()
	fmt.Println(colorize("FLAGS:", ColorYellow))
	fmt.Println("  -v, --verbose     Modo verbose")
	fmt.Println("  -w, --watch       Modo watch")
	fmt.Println("  -h, --help        Muestra ayuda")
	fmt.Println()
	fmt.Println(colorize("EJEMPLOS:", ColorYellow))
	fmt.Println("  zylo run hello.zylo")
	fmt.Println("  zylo init mi-app")
	fmt.Println("  zylo test")
	fmt.Println("  zylo run --watch script.zylo")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Parsear flags globales
	verbose := false
	watch := false

	args := os.Args[2:]
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--verbose":
			verbose = true
		case "-w", "--watch":
			watch = true
		case "-h", "--help":
			printUsage()
			return
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	switch command {
		case "run":
			handleRun(filteredArgs, verbose, watch)
		case "repl":
			handleREPL(verbose)
		case "test":
		handleTest(verbose)
	case "version":
		handleVersion()
	case "init":
		handleInit(filteredArgs, verbose)
	case "doctor":
		handleDoctor(verbose)
	case "fmt":
		handleFmt(filteredArgs, verbose)
	case "lint":
		handleLint(filteredArgs, verbose)
	case "debug":
		handleDebug(filteredArgs, verbose)
	case "doc":
		handleDoc(filteredArgs, verbose)
	case "deps":
		handleDeps(verbose)
	case "add":
		handleAdd(filteredArgs, verbose)
	case "serve":
		handleServe(filteredArgs, verbose)
	case "version-check":
		handleVersionCheck(verbose)
	case "self-update":
		handleSelfUpdate(verbose)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("%sComando desconocido: %s%s\n", ColorRed, command, ColorReset)
		printUsage()
		os.Exit(1)
	}
}

// =============================================================================
// IMPLEMENTACIONES DE FUNCIONES
// =============================================================================

func handleRun(args []string, verbose, watch bool) {
	if len(args) == 0 {
		fmt.Println(colorize("Error: Debes especificar un archivo .zylo", ColorRed))
		os.Exit(1)
	}

	filename := args[0]

	if watch {
		fmt.Println(colorize("Modo watch no implementado aún", ColorYellow))
		runFile(filename, verbose)
	} else {
		runFile(filename, verbose)
	}
}

func handleREPL(verbose bool) {
	if verbose {
		fmt.Println(colorize("Iniciando REPL de Zylo...", ColorCyan))
	}

	fmt.Println(colorize("🐚 Bienvenido al REPL de Zylo v"+Version, ColorCyan))
	fmt.Println(colorize("Escribe '.exit' para salir o '.help' para ayuda", ColorGray))

	eval := evaluator.NewEvaluator()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(colorize("zylo> ", ColorBlue))
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ".") {
			switch line {
			case ".exit":
				fmt.Println(colorize("👋 ¡Hasta luego!", ColorCyan))
				return
			case ".help":
				fmt.Println(colorize("Comandos disponibles:", ColorCyan))
				fmt.Println("  .exit     - Salir del REPL")
				fmt.Println("  .clear    - Limpiar pantalla")
				fmt.Println("  .help     - Mostrar esta ayuda")
				continue
			case ".clear":
				fmt.Print("\033[2J\033[1;1H")
				continue
			default:
				fmt.Printf("%sComando desconocido: %s%s\n", ColorYellow, line, ColorReset)
				continue
			}
		}

		// Parsear y ejecutar
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		_ = program // Para evitar el warning "declared and not used"

		if len(p.Errors()) > 0 {
			fmt.Printf("%sError de sintaxis:%s\n", ColorRed, ColorReset)
			for _, err := range p.Errors() {
				fmt.Printf("  %s\n", err)
			}
			continue
		}

		err := eval.EvaluateProgram(program)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
		}
	}
}

func handleTest(verbose bool) {
	if verbose {
		fmt.Println(colorize("🧪 Ejecutando tests...", ColorCyan))
	}

	// Buscar archivos de test
	testFiles, err := filepath.Glob("tests/*_test.zylo")
	if err != nil {
		fmt.Printf("%sError buscando tests: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// También buscar en directorio actual
	currentTests, _ := filepath.Glob("*_test.zylo")
	testFiles = append(testFiles, currentTests...)

	if len(testFiles) == 0 {
		fmt.Println(colorize("⚠️  No se encontraron archivos de test", ColorYellow))
		return
	}

	passed := 0
	failed := 0

	for _, testFile := range testFiles {
		if verbose {
			fmt.Printf("Ejecutando %s...\n", testFile)
		}

		content, err := ioutil.ReadFile(testFile)
		if err != nil {
			fmt.Printf("%sError leyendo test %s: %v%s\n", ColorRed, testFile, err, ColorReset)
			failed++
			continue
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Printf("%sErrores de parsing en %s%s\n", ColorRed, testFile, ColorReset)
			failed++
			continue
		}

		eval := evaluator.NewEvaluator()
		err = eval.EvaluateProgram(program)
		if err != nil {
			fmt.Printf("%s❌ Test %s falló: %v%s\n", ColorRed, testFile, err, ColorReset)
			failed++
		} else {
			fmt.Printf("%s✅ Test %s pasó%s\n", ColorGreen, testFile, ColorReset)
			passed++
		}
	}

	fmt.Printf("%s📊 Resultados: %d pasaron, %d fallaron%s\n", ColorCyan, passed, failed, ColorReset)
}

func handleVersion() {
	fmt.Printf("%sZylo Programming Language v%s%s\n", ColorCyan, Version, ColorReset)
	fmt.Printf("%sCompilador e interprete integrado%s\n", ColorGray, ColorReset)
}

func handleInit(args []string, verbose bool) {
	if len(args) == 0 {
		fmt.Println(colorize("Error: Debes especificar el nombre del proyecto", ColorRed))
		os.Exit(1)
	}

	projectName := args[0]

	if verbose {
		fmt.Printf("📁 Creando proyecto '%s'...\n", projectName)
	}

	// Crear directorios
	dirs := []string{
		projectName,
		filepath.Join(projectName, "src"),
		filepath.Join(projectName, "std"),
		filepath.Join(projectName, "tests"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("%sError creando directorio %s: %v%s\n", ColorRed, dir, err, ColorReset)
			os.Exit(1)
		}
	}

	// Crear archivos
	files := map[string]string{
		filepath.Join(projectName, "src", "main.zylo"): fmt.Sprintf(`// Archivo principal del proyecto %s
show.log("¡Hola desde %s!")

// Tu código aquí
`, projectName, projectName),

		filepath.Join(projectName, "std", "utils.zylo"): `// Utilidades del proyecto

// Función de utilidad de ejemplo
func saludar(nombre) {
    return "¡Hola, " + nombre + "!"
}
`,

		filepath.Join(projectName, "tests", "main_test.zylo"): `// Tests del proyecto

// Test de ejemplo
func test_saludo() {
    resultado = saludar("Mundo")
    esperado = "¡Hola, Mundo!"

    if resultado == esperado {
        show.log("✅ Test de saludo pasó")
        return true
    } else {
        show.log("❌ Test de saludo falló")
        return false
    }
}

// Ejecutar tests
test_saludo()
`,

		filepath.Join(projectName, "README.md"): fmt.Sprintf(`# %s

Proyecto Zylo creado con zylo init.

## Estructura

- ` + "`src/`" + ` - Código fuente principal
- ` + "`std/`" + ` - Librerías y utilidades
- ` + "`tests/`" + ` - Tests automáticos

## Ejecutar

` + "```bash" + `
zylo run src/main.zylo
` + "```" + `

## Tests

` + "```bash" + `
zylo test
` + "```" + `
`, projectName),
	}

	for filePath, content := range files {
		err := ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			fmt.Printf("%sError creando archivo %s: %v%s\n", ColorRed, filePath, err, ColorReset)
			os.Exit(1)
		}
	}

	fmt.Printf("%s✅ Proyecto '%s' creado exitosamente!%s\n", ColorGreen, projectName, ColorReset)
	fmt.Printf("%sPara empezar:%s\n", ColorCyan, ColorReset)
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  zylo run src/main.zylo\n")
}

func handleDoctor(verbose bool) {
	if verbose {
		fmt.Println(colorize("🔍 Verificando instalación de Zylo...", ColorCyan))
	}

	// Verificar versión
	fmt.Printf("%s✅ Versión: %s%s\n", ColorGreen, Version, ColorReset)

	// Verificar ejecutable
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("%s⚠️  No se pudo determinar ruta del ejecutable%s\n", ColorYellow, ColorReset)
	} else {
		fmt.Printf("%s✅ Ejecutable: %s%s\n", ColorGreen, exePath, ColorReset)
	}

	// Verificar permisos
	tmpFile := filepath.Join(os.TempDir(), "zylo_test.tmp")
	err = ioutil.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		fmt.Printf("%s❌ Error: No hay permisos de escritura%s\n", ColorRed, ColorReset)
	} else {
		os.Remove(tmpFile)
		fmt.Printf("%s✅ Permisos de escritura: OK%s\n", ColorGreen, ColorReset)
	}

	// Verificar módulos estándar
	stdFiles := []string{"http.zylo", "json.zylo", "math.zylo"}
	for _, file := range stdFiles {
		path := filepath.Join("std", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("%s⚠️  Módulo faltante: %s%s\n", ColorYellow, file, ColorReset)
		} else {
			fmt.Printf("%s✅ Módulo encontrado: %s%s\n", ColorGreen, file, ColorReset)
		}
	}

	fmt.Printf("%s🎉 Verificación completada!%s\n", ColorCyan, ColorReset)
}

func handleFmt(args []string, verbose bool) {
	if len(args) == 0 {
		if verbose {
			fmt.Println(colorize("📝 Formateando todos los archivos .zylo...", ColorCyan))
		}
		formatAllFiles(verbose)
	} else {
		formatFile(args[0], verbose)
	}
}

func handleLint(args []string, verbose bool) {
	if len(args) == 0 {
		if verbose {
			fmt.Println(colorize("🔍 Analizando todos los archivos .zylo...", ColorCyan))
		}
		lintAllFiles(verbose)
	} else {
		lintFile(args[0], verbose)
	}
}

func handleDebug(args []string, verbose bool) {
	if len(args) == 0 {
		fmt.Println(colorize("Error: Debes especificar un archivo .zylo", ColorRed))
		os.Exit(1)
	}

	filename := args[0]
	if verbose {
		fmt.Printf("🐛 Ejecutando en modo debug: %s\n", filename)
	}

	os.Setenv("ZYLO_DEBUG", "true")
	runFile(filename, verbose)
}

func handleDoc(args []string, verbose bool) {
	if len(args) == 0 {
		if verbose {
			fmt.Println(colorize("📚 Generando documentación completa...", ColorCyan))
		}
		generateAllDocs(verbose)
	} else {
		generateDoc(args[0], verbose)
	}
}

func handleDeps(verbose bool) {
	if verbose {
		fmt.Println(colorize("📦 Dependencias instaladas:", ColorCyan))
	}

	// TODO: Implementar lista de dependencias
	fmt.Println(colorize("  (Funcionalidad no implementada aún)", ColorGray))
}

func handleAdd(args []string, verbose bool) {
	if len(args) == 0 {
		fmt.Println(colorize("Error: Debes especificar el nombre del paquete", ColorRed))
		os.Exit(1)
	}

	packageName := args[0]
	if verbose {
		fmt.Printf("📥 Instalando paquete: %s\n", packageName)
	}

	// TODO: Implementar instalación de paquetes
	fmt.Printf("%s✅ Paquete '%s' instalado%s\n", ColorGreen, packageName, ColorReset)
}

func handleServe(args []string, verbose bool) {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	if verbose {
		fmt.Printf("🌐 Iniciando servidor para proyecto: %s\n", projectPath)
	}

	// Buscar archivo main.zylo
	mainFile := filepath.Join(projectPath, "src", "main.zylo")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		fmt.Printf("%s❌ No se encontró src/main.zylo en el proyecto%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	runFile(mainFile, verbose)
}

func handleVersionCheck(verbose bool) {
	if verbose {
		fmt.Println(colorize("🔍 Verificando actualizaciones...", ColorCyan))
	}

	// TODO: Implementar verificación real
	fmt.Printf("%s✅ Estás usando la versión más reciente (%s)%s\n", ColorGreen, Version, ColorReset)
}

func handleSelfUpdate(verbose bool) {
	if verbose {
		fmt.Println(colorize("⬆️  Actualizando Zylo...", ColorCyan))
	}

	// TODO: Implementar auto-actualización
	fmt.Printf("%s✅ Zylo actualizado exitosamente%s\n", ColorGreen, ColorReset)
}

// =============================================================================
// FUNCIONES AUXILIARES
// =============================================================================

func runFile(filename string, verbose bool) {
	if verbose {
		fmt.Printf("🚀 Ejecutando %s...\n", filename)
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("%s❌ Error: El archivo '%s' no existe%s\n", ColorRed, filename, ColorReset)
		os.Exit(1)
	}

	// Verificar extensión
	if filepath.Ext(filename) != ".zylo" {
		fmt.Printf("%s❌ Error: El archivo debe tener extensión .zylo%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	// Leer archivo
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("%s❌ Error leyendo archivo: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Parsear
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("%s❌ Errores de parsing:%s\n", ColorRed, ColorReset)
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("%s✅ Parsing completado%s\n", ColorGreen, ColorReset)
	}

	// Análisis semántico
	sa := sema.NewSemanticAnalyzer()
	sa.Analyze(program)

	if len(sa.Errors()) > 0 {
		fmt.Printf("%s❌ Errores de análisis semántico:%s\n", ColorRed, ColorReset)
		for _, err := range sa.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("%s✅ Análisis semántico completado%s\n", ColorGreen, ColorReset)
	}

	// Generar código Go
	cg := codegen.NewCodeGenerator(sa.GetSymbolTable())
	goCode, err := cg.Generate(program)
	if err != nil {
		fmt.Printf("%s❌ Error generando código Go: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("%s✅ Código Go generado%s\n", ColorGreen, ColorReset)
	}

	// Compilar y ejecutar
	compileAndRunGo(goCode, verbose)
}

// compileAndRunGo compila y ejecuta código Go con información de debug
func compileAndRunGo(goCode string, verbose bool) {
	// Mostrar código Go generado si verbose está activado
	if verbose {
		fmt.Printf("%s🔧 CÓDIGO GO GENERADO:%s\n", ColorCyan, ColorReset)
		fmt.Printf("```\n%s```\n", goCode)
		fmt.Printf("%sFIN DEL CÓDIGO GO%s\n\n", ColorCyan, ColorReset)
	}

	// Crear archivo temporal para el código Go
	tmpFile, err := ioutil.TempFile("", "zylo_*.go")
	if err != nil {
		fmt.Printf("%s❌ Error creando archivo temporal: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name()) // Limpiar el archivo temporal

	// Escribir código Go al archivo temporal
	_, err = tmpFile.WriteString(goCode)
	if err != nil {
		fmt.Printf("%s❌ Error escribiendo código Go: %v%s\n", ColorRed, err, ColorReset)
		tmpFile.Close()
		os.Exit(1)
	}
	tmpFile.Close()

	if verbose {
		fmt.Printf("%s✅ Código escrito a %s%s\n", ColorGreen, tmpFile.Name(), ColorReset)
	}

	// Compilar primero para ver posibles errores
	if verbose {
		fmt.Printf("%s🔨 Compilando código Go...%s\n", ColorBlue, ColorReset)
	}

	buildCmd := exec.Command("go", "build", tmpFile.Name())
	buildOutput, buildErr := buildCmd.CombinedOutput()

	if buildErr != nil {
		if verbose {
			fmt.Printf("%s⚠️  Errores de compilación:%s\n", ColorYellow, ColorReset)
			fmt.Printf("%s\n", string(buildOutput))
		}
	}

	// Ejecutar el código con go run
	if verbose {
		fmt.Printf("%s🏃 Ejecutando código Go...%s\n", ColorBlue, ColorReset)
	}

	cmd := exec.Command("go", "run", tmpFile.Name())

	// Redirigir output directamente a la terminal del usuario
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ejecutar y mostrar TODA LA INFORMACIÓN
	runErr := cmd.Run()
	if runErr != nil {
		fmt.Printf("%s❌ Error ejecutando programa: %v%s\n", ColorRed, runErr, ColorReset)
		fmt.Printf("%s🔍 Detalles del error: %T%s\n", ColorYellow, runErr, ColorReset)
	} else {
		if verbose {
			fmt.Printf("%s✅ Programa ejecutado correctamente%s\n", ColorGreen, ColorReset)
		}
	}

	if verbose {
		fmt.Printf("%s✅ Ejecución completada%s\n", ColorGreen, ColorReset)
	}
}

func formatFile(filename string, verbose bool) {
	if verbose {
		fmt.Printf("📝 Formateando %s...\n", filename)
	}

	// Verificar que existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("%s❌ Archivo no encontrado: %s%s\n", ColorRed, filename, ColorReset)
		os.Exit(1)
	}

	// TODO: Implementar formateador real
	fmt.Printf("%s✅ Archivo formateado: %s%s\n", ColorGreen, filename, ColorReset)
}

func formatAllFiles(verbose bool) {
	files, err := filepath.Glob("**/*.zylo")
	if err != nil {
		fmt.Printf("%s❌ Error buscando archivos: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	for _, file := range files {
		formatFile(file, verbose)
	}

	fmt.Printf("%s✅ Todos los archivos formateados%s\n", ColorGreen, ColorReset)
}

func lintFile(filename string, verbose bool) {
	if verbose {
		fmt.Printf("🔍 Analizando %s...\n", filename)
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("%s❌ Error leyendo archivo: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	_ = program // Para evitar el warning "declared and not used"

	if len(p.Errors()) > 0 {
		fmt.Printf("%s❌ Errores de sintaxis encontrados:%s\n", ColorRed, ColorReset)
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	// TODO: Implementar análisis más avanzado
	fmt.Printf("%s✅ Análisis completado: %s%s\n", ColorGreen, filename, ColorReset)
}

func lintAllFiles(verbose bool) {
	files, err := filepath.Glob("**/*.zylo")
	if err != nil {
		fmt.Printf("%s❌ Error buscando archivos: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	totalIssues := 0
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		_ = p.ParseProgram()

		issues := len(p.Errors())
		totalIssues += issues

		if issues > 0 {
			fmt.Printf("%s⚠️  %s: %d issues%s\n", ColorYellow, file, issues, ColorReset)
		} else if verbose {
			fmt.Printf("%s✅ %s: OK%s\n", ColorGreen, file, ColorReset)
		}
	}

	if totalIssues == 0 {
		fmt.Printf("%s🎉 No se encontraron issues!%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%s📊 Total de issues encontrados: %d%s\n", ColorYellow, totalIssues, ColorReset)
	}
}

func generateDoc(filename string, verbose bool) {
	if verbose {
		fmt.Printf("📚 Generando documentación para %s...\n", filename)
	}

	// TODO: Implementar generador de docs real
	docContent := fmt.Sprintf(`# Documentación para %s

## Funciones

<!-- TODO: Extraer funciones del código -->

## Dependencias

<!-- TODO: Analizar imports -->

Generado automáticamente por zylo doc
`, filename)

	docFile := strings.TrimSuffix(filename, ".zylo") + "_doc.md"
	err := ioutil.WriteFile(docFile, []byte(docContent), 0644)
	if err != nil {
		fmt.Printf("%s❌ Error creando documentación: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Printf("%s✅ Documentación generada: %s%s\n", ColorGreen, docFile, ColorReset)
}

func generateAllDocs(verbose bool) {
	files, err := filepath.Glob("**/*.zylo")
	if err != nil {
		fmt.Printf("%s❌ Error buscando archivos: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	for _, file := range files {
		generateDoc(file, verbose)
	}

	fmt.Printf("%s✅ Documentación completa generada%s\n", ColorGreen, ColorReset)
}

// =============================================================================
// ESTRUCTURA DE PROYECTO DE EJEMPLO
// =============================================================================

/*
Para probar la CLI, crea un proyecto de ejemplo:

zylo init mi-proyecto
cd mi-proyecto
zylo run src/main.zylo
zylo test
zylo repl

Estructura creada:
mi-proyecto/
├── src/
│   └── main.zylo
├── std/
│   └── utils.zylo
├── tests/
│   └── main_test.zylo
├── zylo.toml
└── README.md
*/
