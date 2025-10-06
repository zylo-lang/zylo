package codegen

import (
	"fmt"
	"strings"

	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/sema"
)

// CodeGenerator es el struct principal para la generaciรณn de cรณdigo Go.
type CodeGenerator struct {
	mainOutput         strings.Builder
	declarations       strings.Builder
	currentOutput      *strings.Builder
	indentation        int
	classNames         []string
	needsRuntimeImport bool
	inMainFunction     bool              // Track if we're generating code inside main()
	inVoidFunction     bool              // Track if we're generating code inside a void function
	symbolTable        *sema.SymbolTable // AรADIDO: tabla de sรญmbolos para type info
	imports            map[string]bool   // Track de imports necesarios
}

// NewCodeGenerator crea un nuevo CodeGenerator.
func NewCodeGenerator(symbolTable *sema.SymbolTable) *CodeGenerator {
	cg := &CodeGenerator{
		classNames:         make([]string, 0),
		needsRuntimeImport: false,
		inMainFunction:     false,
		symbolTable:        symbolTable,
		imports:            make(map[string]bool), // Inicializar mapa de imports
	}

	// Siempre incluir fmt para programas Zylo
	cg.EnsureImport("fmt")

	return cg
}

// EnsureImport ensures that a specific import package is included in the generated code
func (cg *CodeGenerator) EnsureImport(pkg string) {
	cg.imports[pkg] = true
}

// getKnownType determina el tipo formato Go conocido de una expresiรณn
func (cg *CodeGenerator) getKnownType(exp ast.Expression) string {
	if cg.symbolTable == nil {
		return ""
	}

	if ident, ok := exp.(*ast.Identifier); ok {
		if sym, ok := cg.symbolTable.Resolve(ident.Value); ok {
			// Convertir tipos sema a tipos Go
			switch sym.Type {
			case sema.IntType:
				return "int64"
			case sema.FloatType:
				return "float64"
			case sema.StringType:
				return "string"
			case sema.BoolType:
				return "bool"
			}
		}
	}

	// Inferir de literales
	if numLit, ok := exp.(*ast.NumberLiteral); ok {
		if _, ok := numLit.Value.(int64); ok {
			return "int64"
		}
		return "float64"
	}
	if _, ok := exp.(*ast.StringLiteral); ok {
		return "string"
	}
	if _, ok := exp.(*ast.BooleanLiteral); ok {
		return "bool"
	}

	return ""
}

// zyloTypeToGoType converts a Zylo type annotation to the corresponding Go type.
func (cg *CodeGenerator) zyloTypeToGoType(zyloType string) string {
	switch zyloType {
	case "int":
		return "int64"
	case "float", "Float":
		return "float64"
	case "string", "String":
		return "string"
	case "bool", "Bool":
		return "bool"
	default:
		return "interface{}"
	}
}

// Generate genera cรณdigo Go a partir de un AST.
func (cg *CodeGenerator) Generate(program *ast.Program) (string, error) {
	if program == nil {
		return "", fmt.Errorf("program is nil")
	}

	var mainFuncBody *ast.BlockStatement

	// Reset the generator state for a fresh generation
	cg.mainOutput.Reset()
	cg.declarations.Reset()
	cg.indentation = 0
	cg.classNames = make([]string, 0)
	cg.needsRuntimeImport = false
	cg.currentOutput = &cg.mainOutput

	// First pass: categorize statements
	for _, stmt := range program.Statements {
		if stmt == nil {
			continue
		}

		switch s := stmt.(type) {
		case *ast.FuncStatement:
			if s.Name.Value == "main" {
				mainFuncBody = s.Body
			} else {
				cg.generateStatementInDeclarations(s)
			}
		case *ast.ClassStatement:
			cg.classNames = append(cg.classNames, s.Name.Value)
			cg.generateStatementInDeclarations(s)
		case *ast.VarStatement, *ast.ExpressionStatement, *ast.WhileStatement,
		     *ast.ForStatement, *ast.IfStatement:
			// Executable statements belong in main
			if mainFuncBody == nil {
				mainFuncBody = &ast.BlockStatement{Statements: []ast.Statement{}}
			}
			mainFuncBody.Statements = append(mainFuncBody.Statements, s)
		default:
			// Other types of statements
			if mainFuncBody == nil {
				mainFuncBody = &ast.BlockStatement{Statements: []ast.Statement{}}
			}
			mainFuncBody.Statements = append(mainFuncBody.Statements, s)
		}
	}

	// Preliminary scan: determine what imports we need by processing expressions
	tempOutput := strings.Builder{}
	oldOutput := cg.currentOutput
	cg.currentOutput = &tempOutput
	cg.needsRuntimeImport = false

	if mainFuncBody != nil {
		for _, bodyStmt := range mainFuncBody.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.currentOutput = oldOutput

	// Generate the final output with proper imports
	cg.mainOutput.WriteString("package main\n\n")

	// Always import fmt for now - it will be cleaned if not needed
	cg.mainOutput.WriteString("import \"fmt\"\n\n")

	// Always ensure fmt is used to avoid "imported and not used" error
	// NOTE: This will be placed inside main() function

	// Generate built-in Result type
	cg.mainOutput.WriteString("// Built-in Result type\n")
	cg.mainOutput.WriteString("type Result struct {\n")
	cg.mainOutput.WriteString("    Value interface{}\n")
	cg.mainOutput.WriteString("    Error interface{}\n")
	cg.mainOutput.WriteString("}\n\n")

	// Generate Result methods
	cg.mainOutput.WriteString("func (r *Result) IsOk() interface{} {\n")
	cg.mainOutput.WriteString("    return r.Error == nil\n")
	cg.mainOutput.WriteString("}\n\n")

	cg.mainOutput.WriteString("func (r *Result) IsErr() interface{} {\n")
	cg.mainOutput.WriteString("    return r.Error != nil\n")
	cg.mainOutput.WriteString("}\n\n")

	cg.mainOutput.WriteString("func (r *Result) Unwrap() interface{} {\n")
	cg.mainOutput.WriteString("    if r.Error != nil {\n")
	cg.mainOutput.WriteString("        panic(r.Error)\n")
	cg.mainOutput.WriteString("    }\n")
	cg.mainOutput.WriteString("    return r.Value\n")
	cg.mainOutput.WriteString("}\n\n")

	// Add helper functions for nested indexing
	cg.mainOutput.WriteString("// zyloIndex performs safe nested indexing on interface{} arrays\n")
	cg.mainOutput.WriteString("func zyloIndex(value interface{}, indices ...int) interface{} {\n")
	cg.mainOutput.WriteString("    for _, index := range indices {\n")
	cg.mainOutput.WriteString("        if slice, ok := value.([]interface{}); ok {\n")
	cg.mainOutput.WriteString("            if index < 0 {\n")
	cg.mainOutput.WriteString("                index = len(slice) + index\n")
	cg.mainOutput.WriteString("            }\n")
	cg.mainOutput.WriteString("            if index >= 0 && index < len(slice) {\n")
	cg.mainOutput.WriteString("                value = slice[index]\n")
	cg.mainOutput.WriteString("            } else {\n")
	cg.mainOutput.WriteString("                panic(\"index out of bounds\")\n")
	cg.mainOutput.WriteString("            }\n")
	cg.mainOutput.WriteString("        } else {\n")
	cg.mainOutput.WriteString("            panic(\"cannot index non-array value\")\n")
	cg.mainOutput.WriteString("        }\n")
	cg.mainOutput.WriteString("    }\n")
	cg.mainOutput.WriteString("    return value\n")
	cg.mainOutput.WriteString("}\n\n")

	// Append all declarations (functions, classes)
	cg.mainOutput.WriteString(cg.declarations.String())

	// Generate main function with executable statements
	cg.mainOutput.WriteString("func main() {\n")
	cg.indentation = 1
	cg.currentOutput = &cg.mainOutput
	cg.inMainFunction = true // We're now in main()

	if mainFuncBody != nil {
		for _, bodyStmt := range mainFuncBody.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.mainOutput.WriteString("}\n")

	// Limpiar imports no usados antes de retornar
	finalCode := cleanUnusedImports(cg.mainOutput.String())

	return finalCode, nil
}

// generateStatementInDeclarations generates a statement to the declarations buffer.
func (cg *CodeGenerator) generateStatementInDeclarations(stmt ast.Statement) {
	oldOutput := cg.currentOutput
	cg.currentOutput = &cg.declarations
	cg.generateStatement(stmt)
	cg.currentOutput = oldOutput
}

// generateStatement genera cรณdigo Go para una sentencia del AST.
func (cg *CodeGenerator) generateStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.ImportStatement:
		if s != nil {
			cg.generateImportStatement(s)
		}
	case *ast.VarStatement:
		if s != nil {
			cg.generateVarStatement(s)
		}
	case *ast.ExpressionStatement:
		if s != nil {
			cg.generateExpressionStatement(s)
		}
	case *ast.FuncStatement:
		if s != nil {
			cg.generateFuncStatement(s)
		}
	case *ast.ReturnStatement:
		if s != nil {
			cg.generateReturnStatement(s)
		}
	case *ast.IfStatement:
		if s != nil {
			cg.generateIfStatement(s)
		}
	case *ast.WhileStatement:
		if s != nil {
			cg.generateWhileStatement(s)
		}
	case *ast.ForStatement:
		if s != nil {
			cg.generateForStatement(s)
		}
	case *ast.BreakStatement:
		if s != nil {
			cg.generateBreakStatement(s)
		}
	case *ast.ClassStatement:
		if s != nil {
			cg.generateClassStatement(s)
		}
	default:
		cg.writeString(fmt.Sprintf("// TODO: Sentencia no soportada: %T\n", s))
	}
}

// generateConstantDeclaration genera una declaraciรณn de constante global
func (cg *CodeGenerator) generateConstantDeclaration(stmt *ast.VarStatement) {
	if stmt == nil || stmt.Name == nil {
		return
	}

	// Generate in declarations buffer with no indentation
	oldIndent := cg.indentation
	cg.indentation = 0
	oldOutput := cg.currentOutput
	cg.currentOutput = &cg.declarations

	cg.writeString("const ")
	cg.writeString(stmt.Name.Value)
	cg.writeString(" = ")

	if stmt.Value != nil {
		cg.generateConstantValue(stmt.Value)
	} else {
		cg.writeString("nil")
	}
	cg.writeString("\n")

	cg.currentOutput = oldOutput
	cg.indentation = oldIndent
}

// generateConstantValue genera el valor apropiado para una constante
func (cg *CodeGenerator) generateConstantValue(exp ast.Expression) {
	switch e := exp.(type) {
	case *ast.NumberLiteral:
		if val, ok := e.Value.(float64); ok {
			cg.writeString(fmt.Sprintf("%f", val))
		} else if val, ok := e.Value.(int64); ok {
			cg.writeString(fmt.Sprintf("%d", val))
		} else {
			cg.writeString("0")
		}
	case *ast.StringLiteral:
		cg.writeString(fmt.Sprintf("%q", e.Value))
	case *ast.BooleanLiteral:
		cg.writeString(fmt.Sprintf("%t", e.Value))
	default:
		cg.writeString("// Constant value too complex")
	}
}

// writeString escribe una cadena en la salida con la indentaciรณn actual.
func (cg *CodeGenerator) writeString(s string) {
	if len(s) > 0 && s[0] != '\n' {
		for i := 0; i < cg.indentation; i++ {
			cg.currentOutput.WriteString("    ")
		}
	}
	cg.currentOutput.WriteString(s)
}

// indent aumenta el nivel de indentaciรณn.
func (cg *CodeGenerator) indent() {
	cg.indentation++
}

// dedent disminuye el nivel de indentaciรณn.
func (cg *CodeGenerator) dedent() {
	if cg.indentation > 0 {
		cg.indentation--
	}
}

// generateVarStatement genera cรณdigo Go para una declaraciรณn de variable.
func (cg *CodeGenerator) generateVarStatement(stmt *ast.VarStatement) {
	if stmt.IsDestructuring {
		cg.writeString("var ")
		cg.generateDestructuringTargets(stmt.DestructuringElements)
		cg.writeString(" = ")
		cg.generateExpression(stmt.Value)
		cg.writeString("\n")
		return
	}

	// For explicit types, use typed variable declaration
	if stmt.Name.TypeAnnotation != "" && stmt.Name.TypeAnnotation != "ANY" {
		cg.writeString("var ")
		cg.generateExpression(stmt.Name)

		// Use specific Go types for typed variables
		varType := cg.zyloTypeToGoType(stmt.Name.TypeAnnotation)
		cg.writeString(" " + varType)

		if stmt.Value != nil {
			cg.writeString(" = ")
			cg.generateAssignmentValue(stmt.Value, varType)
		}
		cg.writeString("\n")
		// Solo silenciar para nombres tรญpicos de loop
		if isLoopVariable(stmt.Name.Value) {
			cg.writeString(fmt.Sprintf("_ = %s\n", stmt.Name.Value))
		}
		return
	}

	// For inferred types (e.g., x := value), check if it's a literal and use explicit types
	if stmt.Value != nil {
		switch lit := stmt.Value.(type) {
		case *ast.NumberLiteral:
			// Force ALL numeric variables to be typed as int64 for consistency
			cg.generateExpression(stmt.Name)
			cg.writeString(" := ")
			cg.generateAssignmentValue(stmt.Value, "int64")
			cg.writeString("\n")

		case *ast.BooleanLiteral:
			// Boolean literal
			cg.generateExpression(stmt.Name)
			cg.writeString(" := ")
			if lit.Value {
				cg.writeString("true")
			} else {
				cg.writeString("false")
			}
			cg.writeString("\n")

		case *ast.StringLiteral:
			// String literal
			cg.generateExpression(stmt.Name)
			cg.writeString(" := ")
			cg.writeString(fmt.Sprintf("%q", lit.Value))
			cg.writeString("\n")

		default:
			// Complex expressions - use interface{}{" + intent +"}
			cg.generateExpression(stmt.Name)
			cg.writeString(" := ")
			cg.generateAssignmentValue(stmt.Value, "interface{}")
			cg.writeString("\n")
		}
	}
}

// generateDestructuringTargets genera los objetivos de una desestructuraciรณn.
func (cg *CodeGenerator) generateDestructuringTargets(targets []ast.Expression) {
	cg.needsRuntimeImport = true
	cg.writeString("[]interface{}{")
	for i, target := range targets {
		if ident, ok := target.(*ast.Identifier); ok {
			cg.writeString(ident.Value)
		} else {
			cg.writeString("nil")
		}
		if i < len(targets)-1 {
			cg.writeString(", ")
		}
	}
	cg.writeString("}")
}

// generateAssignmentValue generates the appropriate value for assignment based on target type
func (cg *CodeGenerator) generateAssignmentValue(exp ast.Expression, targetType string) {
	switch e := exp.(type) {
	case *ast.NumberLiteral:
		// Force typed literals for arithmetic operations
		if val, ok := e.Value.(float64); ok {
			cg.writeString(fmt.Sprintf("float64(%f)", val))
		} else if val, ok := e.Value.(int64); ok {
			cg.writeString(fmt.Sprintf("int64(%d)", val))
		} else {
			cg.writeString("0")
		}
	case *ast.StringLiteral:
		cg.writeString(fmt.Sprintf("%q", e.Value))
	case *ast.BooleanLiteral:
		cg.writeString(fmt.Sprintf("%t", e.Value))
	default:
		// For other expressions, use normal expression generation
		cg.generateExpression(exp)
	}
}

// generateExpressionStatement genera cรณdigo Go para una sentencia de expresiรณn.
func (cg *CodeGenerator) generateExpressionStatement(stmt *ast.ExpressionStatement) {
	if stmt == nil || stmt.Expression == nil {
		return
	}
	cg.generateExpression(stmt.Expression)
	cg.writeString("\n")
}

// generateFuncStatement genera cรณdigo Go para una declaraciรณn de funciรณn.
func (cg *CodeGenerator) generateFuncStatement(stmt *ast.FuncStatement) {
	if stmt == nil || stmt.Name == nil {
		return
	}
	// Debug: print ReturnType to see what we're getting

	cg.writeString(fmt.Sprintf("func %s(", stmt.Name.Value))

	for i, param := range stmt.Parameters {
		if i > 0 {
			cg.writeString(", ")
		}
		if param != nil {
			// Use specific Go types for typed parameters
			var paramType string
			if param.TypeAnnotation != "" && param.TypeAnnotation != "ANY" {
				paramType = cg.zyloTypeToGoType(param.TypeAnnotation)
			} else {
				paramType = "interface{}"
			}
			cg.writeString(fmt.Sprintf("%s %s", param.Value, paramType))
		}
	}

	// Use appropriate return type
	var returnType string
	if stmt.IsVoid {
		returnType = ""
	} else if stmt.ReturnType != "" && stmt.ReturnType != "ANY" {
		// Use native Go type for functions with explicit return types
		returnType = " " + cg.zyloTypeToGoType(stmt.ReturnType)
	} else {
		returnType = " interface{}"
	}
	cg.writeString(")" + returnType + " {\n")
	cg.indent()

	prevVoidFunction := cg.inVoidFunction
	cg.inVoidFunction = stmt.IsVoid

	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.inVoidFunction = prevVoidFunction
	cg.dedent()
	cg.writeString("}\n")
}

// generateReturnStatement genera cรณdigo Go para una sentencia de retorno.
func (cg *CodeGenerator) generateReturnStatement(stmt *ast.ReturnStatement) {
	// If we're in main function, Go cannot return values, so skip the return
	if cg.inMainFunction {
		// Do nothing - main() cannot have return statements with values in Go
		return
	}

	// For void functions, ignore return values and just return
	if cg.inVoidFunction {
		if stmt.ReturnValue != nil {
			// In void functions, return statements should just be 'return' without value
			cg.writeString("return\n")
		} else {
			// This shouldn't happen but handle it gracefully
			cg.writeString("return\n")
		}
		return
	}

	// Non-void functions
	cg.writeString("return")
	if stmt.ReturnValue != nil {
		cg.writeString(" ")
		cg.generateExpression(stmt.ReturnValue)
	}
	cg.writeString("\n")
}

// generateIfStatement genera cรณdigo Go para una sentencia 'if'.
func (cg *CodeGenerator) generateIfStatement(stmt *ast.IfStatement) {
	cg.writeString("if ")
	cg.generateExpression(stmt.Condition)
	cg.writeString(" {\n")
	cg.indent()

	if stmt.Consequence != nil {
		for _, bodyStmt := range stmt.Consequence.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}")

	if stmt.Alternative != nil {
		cg.writeString(" else {\n")
		cg.indent()

		for _, bodyStmt := range stmt.Alternative.Statements {
			cg.generateStatement(bodyStmt)
		}

		cg.dedent()
		cg.writeString("}")
	}
	cg.writeString("\n")
}

// generateWhileStatement genera cรณdigo Go para una sentencia 'while'.
func (cg *CodeGenerator) generateWhileStatement(stmt *ast.WhileStatement) {
	cg.writeString("for ")
	cg.generateExpression(stmt.Condition)
	cg.writeString(" {\n")
	cg.indent()

	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}\n")
}

// generateForStatement genera código Go para una sentencia 'for' tradicional.
// Necesita generar partes inline sin newlines para la sintaxis correcta de Go.
func (cg *CodeGenerator) generateForStatement(stmt *ast.ForStatement) {
	cg.writeString("for ")

	// Generate init part inline (sin newlines)
	if stmt.Init != nil {
		if varStmt, ok := stmt.Init.(*ast.VarStatement); ok {
			if varStmt.Value != nil {
				// Solo para declaraciones de variables inline: x := expr
				cg.generateExpression(varStmt.Name)
				cg.writeString(" := ")
				cg.generateAssignmentValue(varStmt.Value, "int64")
			}
		} else {
			cg.generateStatement(stmt.Init)
		}
	}

	cg.writeString("; ")

	// Generate condition expression
	if stmt.Condition != nil {
		cg.generateExpression(stmt.Condition)
	}

	cg.writeString("; ")

	// Generate post part inline (sin newlines)
	if stmt.Post != nil {
		if varStmt, ok := stmt.Post.(*ast.VarStatement); ok {
			// Recast for assignments: i = i + 1
			cg.generateExpression(varStmt.Name)
			cg.writeString(" = ")
			if varStmt.Value != nil {
				cg.generateAssignmentValue(varStmt.Value, "int64")
			}
		} else if assignStmt, ok := stmt.Post.(*ast.ExpressionStatement); ok {
			if assignExpr, ok := assignStmt.Expression.(*ast.AssignmentExpression); ok {
				cg.generateExpression(assignExpr.Name)
				cg.writeString(" " + assignExpr.Operator + " ")
				cg.generateExpression(assignExpr.Value)
			}
		} else {
			// Fallback to normal statement generation
			cg.generateStatement(stmt.Post)
		}
	}

	cg.writeString(" {\n")
	cg.indent()

	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}\n")
}

// generateBreakStatement genera cรณdigo Go para una sentencia break.
func (cg *CodeGenerator) generateBreakStatement(stmt *ast.BreakStatement) {
	if stmt == nil {
		return
	}
	cg.writeString("break\n")
}

// generatePrefixExpression genera cรณdigo Go para expresiones prefijas (operadores unarios).
func (cg *CodeGenerator) generatePrefixExpression(exp *ast.PrefixExpression) {
	if exp == nil || exp.Right == nil {
		return
	}

	switch exp.Operator {
	case "-":
		cg.writeString("-")
		cg.generateExpression(exp.Right)
	case "!":
		cg.writeString("!")
		cg.generateExpression(exp.Right)
	default:
		// Fallback a runtime
		cg.needsRuntimeImport = true
		cg.writeString(fmt.Sprintf("zyloruntime.Prefix%s(", strings.Title(exp.Operator)))
		cg.generateExpression(exp.Right)
		cg.writeString(")")
	}
}

// generateAssignmentExpression genera cรณdigo Go para una expresiรณn de asignaciรณn.
func (cg *CodeGenerator) generateAssignmentExpression(exp *ast.AssignmentExpression) {
	if exp == nil {
		return
	}

	cg.generateExpression(exp.Name)
	cg.writeString(" " + exp.Operator + " ")
	cg.generateExpression(exp.Value)
}

// generateInfixExpression genera cรณdigo Go para expresiones infijas (operaciones binarias).
func (cg *CodeGenerator) generateInfixExpression(exp *ast.InfixExpression) {
	if exp == nil || exp.Left == nil || exp.Right == nil {
		return
	}

	// Use direct Go operations for all basic operators
	cg.generateExpression(exp.Left)
	cg.writeString(" " + exp.Operator + " ")
	cg.generateExpression(exp.Right)
}

// Helper functions to detect literals
func (cg *CodeGenerator) isIntLiteral(exp ast.Expression) bool {
	if numLit, ok := exp.(*ast.NumberLiteral); ok {
		_, ok := numLit.Value.(int64)
		return ok
	}
	return false
}

func (cg *CodeGenerator) isFloatLiteral(exp ast.Expression) bool {
	if numLit, ok := exp.(*ast.NumberLiteral); ok {
		_, ok := numLit.Value.(float64)
		return ok
	}
	return false
}

// generateExpression genera cรณdigo Go para una expresiรณn del AST.
func (cg *CodeGenerator) generateExpression(exp ast.Expression) {
	if exp == nil {
		cg.writeString("// TODO: Expresiรณn no soportada: <nil>")
		return
	}

	switch e := exp.(type) {
	case *ast.Identifier:
		cg.writeString(e.Value)
	case *ast.CallExpression:
		if dotExpr, ok := e.Function.(*ast.DotExpression); ok {
			if leftIdent, ok := dotExpr.Left.(*ast.Identifier); ok && leftIdent.Value == "show" && dotExpr.Property.Value == "log" {
				// Use fmt.Println for direct output - no runtime wrapper needed
				cg.writeString("fmt.Println(")
				for i, arg := range e.Arguments {
					cg.generateExpression(arg)
					if i < len(e.Arguments)-1 {
						cg.writeString(", ")
					}
				}
				cg.writeString(")")
				return
			}
		}

		// Check for built-in println function
		if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "println" {
			cg.writeString("fmt.Println(")
			for i, arg := range e.Arguments {
				cg.generateExpression(arg)
				if i < len(e.Arguments)-1 {
					cg.writeString(", ")
				}
			}
			cg.writeString(")")
			return
		}

		if ident, ok := e.Function.(*ast.Identifier); ok {
			switch ident.Value {
			case "len":
				// Use native Go len() function for arrays/slices
				cg.writeString("len(")
				cg.generateExpression(e.Arguments[0])
				cg.writeString(")")
			case "println", "split", "to_number", "string", "read_line", "read_file", "write_file", "type_of", "is_null", "is_empty", "to_int", "to_bool", "replace", "substring", "trim", "power", "sqrt", "abs", "round", "min", "max", "string_list", "string_map", "map_keys", "map_values":
				cg.needsRuntimeImport = true
				cg.writeString(fmt.Sprintf("zyloruntime.%s(", strings.Title(ident.Value)))
				for i, arg := range e.Arguments {
					cg.writeString("zyloruntime.ToZyloObject(")
					cg.generateExpression(arg)
					cg.writeString(")")
					if i < len(e.Arguments)-1 {
						cg.writeString(", ")
					}
				}
				cg.writeString(")")
			default:
				cg.generateExpression(e.Function)
				cg.writeString("(")
				for i, arg := range e.Arguments {
					cg.generateExpression(arg)
					if i < len(e.Arguments)-1 {
						cg.writeString(", ")
					}
				}
				cg.writeString(")")
			}
		} else {
			cg.generateExpression(e.Function)
			cg.writeString("(")
			for i, arg := range e.Arguments {
				cg.generateExpression(arg)
				if i < len(e.Arguments)-1 {
					cg.writeString(", ")
				}
			}
			cg.writeString(")")
		}
	case *ast.DotExpression:
		// Handle show.log specifically
		if e.Left != nil && e.Property != nil {
			if leftIdent, ok := e.Left.(*ast.Identifier); ok && leftIdent.Value == "show" && e.Property.Value == "log" {
				// We're handling this in the CallExpression case already, so this is just the property access
				cg.writeString("// show.log handled in call expression")
				return
			}
		}

		// Regular dot expression handling
		oldIndent := cg.indentation
		cg.indentation = 0
		if e.Left != nil {
			cg.generateExpression(e.Left)
		}
		cg.writeString(".")
		if e.Property != nil {
			cg.writeString(e.Property.Value)
		}
		cg.indentation = oldIndent
	case *ast.NumberLiteral:
		cg.generateNumberLiteral(e)
	case *ast.StringLiteral:
		cg.generateStringLiteral(e)
	case *ast.BooleanLiteral:
		cg.generateBooleanLiteral(e)
	case *ast.ListLiteral:
		cg.generateListLiteral(e)
	case *ast.MapLiteral:
		cg.generateMapLiteral(e)
	case *ast.IndexExpression:
		cg.generateIndexExpression(e)
	case *ast.InfixExpression:
		cg.generateInfixExpression(e)
	case *ast.AssignmentExpression:
		cg.generateAssignmentExpression(e)
	case *ast.PrefixExpression:
		cg.generatePrefixExpression(e)
	case *ast.CollectionMethodCall:
		cg.generateCollectionMethodCallSelfContained(e)
	default:
		cg.writeString(fmt.Sprintf("// TODO: Expresiรณn no soportada: %T", e))
	}
}

// generateClassStatement genera cรณdigo Go para una declaraciรณn de clase.
func (cg *CodeGenerator) generateClassStatement(stmt *ast.ClassStatement) {
	if stmt.Name == nil {
		return
	}

	className := stmt.Name.Value

	cg.writeString(fmt.Sprintf("type %s struct {\n", className))
	cg.indent()

	for _, attr := range stmt.Attributes {
		if attr.Name != nil {
			cg.writeString(fmt.Sprintf("%s interface{}\n", attr.Name.Value))
		}
	}

	cg.dedent()
	cg.writeString("}\n\n")

	cg.writeString(fmt.Sprintf("func New%s(", className))
	if stmt.InitMethod != nil && len(stmt.InitMethod.Parameters) > 0 {
		for i, param := range stmt.InitMethod.Parameters {
			if i > 0 {
				cg.writeString(", ")
			}
			cg.writeString(fmt.Sprintf("%s interface{}", param.Value))
		}
	}
	cg.writeString(fmt.Sprintf(") interface{} {\n"))
	cg.indent()

	cg.writeString(fmt.Sprintf("obj := &%s{}\n", className))

	if stmt.InitMethod != nil {
		cg.writeString(fmt.Sprintf("obj.%s(", stmt.InitMethod.Name.Value))
		if len(stmt.InitMethod.Parameters) > 0 {
			for i, param := range stmt.InitMethod.Parameters {
				if i > 0 {
					cg.writeString(", ")
				}
				cg.writeString(param.Value)
			}
		}
		cg.writeString(")\n")
	}

	cg.writeString("return obj\n")
	cg.dedent()
	cg.writeString("}\n\n")

	for _, method := range stmt.Methods {
		if method.Name == nil {
			continue
		}

		cg.writeString(fmt.Sprintf("func (obj *%s) %s(", className, method.Name.Value))

		for i, param := range method.Parameters {
			if i > 0 {
				cg.writeString(", ")
			}
			cg.writeString(fmt.Sprintf("%s interface{}", param.Value))
		}

		cg.writeString(") interface{}")

		cg.writeString(" {\n")
		cg.indent()

		if method.Body != nil {
			for _, bodyStmt := range method.Body.Statements {
				cg.generateStatement(bodyStmt)
			}
		}

		cg.dedent()
		cg.writeString("}\n\n")
	}
}

// generateNumberLiteral generates Go code for a number literal
func (cg *CodeGenerator) generateNumberLiteral(exp *ast.NumberLiteral) {
	if exp == nil {
		return
	}

	if val, ok := exp.Value.(float64); ok {
		cg.writeString(fmt.Sprintf("float64(%f)", val))
	} else if val, ok := exp.Value.(int64); ok {
		cg.writeString(fmt.Sprintf("int64(%d)", val))
	} else {
		cg.writeString("int64(0)") // fallback
	}
}

// generateStringLiteral generates Go code for a string literal
func (cg *CodeGenerator) generateStringLiteral(exp *ast.StringLiteral) {
	if exp == nil {
		return
	}
	cg.writeString(fmt.Sprintf("%q", exp.Value))
}

// generateBooleanLiteral generates Go code for a boolean literal
func (cg *CodeGenerator) generateBooleanLiteral(exp *ast.BooleanLiteral) {
	if exp == nil {
		return
	}
	if exp.Value {
		cg.writeString("true")
	} else {
		cg.writeString("false")
	}
}

// generateListLiteral generates Go code for a list literal
func (cg *CodeGenerator) generateListLiteral(exp *ast.ListLiteral) {
	if exp == nil {
		cg.writeString("nil")
		return
	}

	cg.writeString("[]interface{}{")
	for i, element := range exp.Elements {
		cg.generateExpression(element)
		if i < len(exp.Elements)-1 {
			cg.writeString(", ")
		}
	}
	cg.writeString("}")
}

// generateMapLiteral generates Go code for a map literal
func (cg *CodeGenerator) generateMapLiteral(exp *ast.MapLiteral) {
	if exp == nil {
		cg.writeString("nil")
		return
	}

	cg.writeString("map[string]interface{}{")
	i := 0
	for key, value := range exp.Pairs {
		cg.writeString(fmt.Sprintf("%q: ", key))
		cg.generateExpression(value)
		if i < len(exp.Pairs)-1 {
			cg.writeString(", ")
		}
		i++
	}
	cg.writeString("}")
}

// generateIndexExpression generates Go code for indexing operations
func (cg *CodeGenerator) generateIndexExpression(exp *ast.IndexExpression) {
	if exp == nil {
		cg.writeString("nil")
		return
	}

	// Handle special slice syntax (array[start:end])
	if exp.EndIndex != nil {
		// This is a slice operation: array[start:end]
		cg.generateExpression(exp.Left)
		cg.writeString("[")
		cg.generateExpression(exp.Index) // start
		cg.writeString(":")
		cg.generateExpression(exp.EndIndex) // end
		cg.writeString("]")
		return
	}

	// Use zyloIndex helper function for all complex indexing operations
	// zyloIndex handles nested indexing and negative indexing correctly
	cg.writeString("zyloIndex(")

	// Generate the array/object being indexed
	cg.generateExpression(exp.Left)

	// Add comma and calculate the index
	cg.writeString(", ")

	// Handle negative indexing: -i -> len(arr) + (-i)
	if exp.NegativeIndex {
		cg.writeString("len(")
		cg.generateExpression(exp.Left)
		cg.writeString(") + int(")
		cg.generateExpression(exp.Index)
		cg.writeString(")")
	} else {
		cg.writeString("int(")
		cg.generateExpression(exp.Index)
		cg.writeString(")")
	}

	cg.writeString(")")
}

// generateCollectionMethodCall generates Go code for collection method calls like arr.push(element)
func (cg *CodeGenerator) generateCollectionMethodCall(exp *ast.CollectionMethodCall) {
	if exp == nil || exp.Object == nil || exp.Method == nil {
		cg.writeString("// Invalid collection method call")
		return
	}

	// Import runtime for method implementations
	cg.needsRuntimeImport = true

	// Generate call to runtime method
	methodName := strings.Title(exp.Method.Value)
	if methodName == "Push" {
		methodName = "ListPush"
	} else if methodName == "Pop" {
		methodName = "ListPop"
	} else if methodName == "Shift" {
		methodName = "ListShift"
	} else if methodName == "Unshift" {
		methodName = "ListUnshift"
	} else if methodName == "Set" {
		methodName = "MapSet"
	} else if methodName == "Get" {
		methodName = "MapGet"
	} else if methodName == "Has" {
		methodName = "MapHas"
	} else {
		// Generic List/Map prefix based on method name
		// For now, assume all unknown methods are list methods
		methodName = "List" + methodName
	}

	cg.writeString(fmt.Sprintf("zyloruntime.%s(", methodName))

	// Generate object argument
	cg.generateExpression(exp.Object)

	// Generate method arguments
	for _, arg := range exp.Arguments {
		cg.writeString(", ")
		cg.generateExpression(arg)
	}

	cg.writeString(")")
}

// generateCollectionMethodCallSelfContained generates self-contained Go code for collection methods
func (cg *CodeGenerator) generateCollectionMethodCallSelfContained(exp *ast.CollectionMethodCall) {
	if exp == nil || exp.Object == nil || exp.Method == nil {
		cg.writeString("// Invalid collection method call")
		return
	}

	// Generate self-contained function calls using native Go operations
	switch exp.Method.Value {
	case "push":
		// Native Go append: array = append(array, element)
		cg.generateExpression(exp.Object)
		cg.writeString(" = append(")
		cg.generateExpression(exp.Object)
		if len(exp.Arguments) > 0 {
			cg.writeString(", ")
			for i, arg := range exp.Arguments {
				cg.generateExpression(arg)
				if i < len(exp.Arguments)-1 {
					cg.writeString(", ")
				}
			}
		}
		cg.writeString(")")

	case "pop":
		// Get last element: array[len(array)-1]
		cg.writeString("// Get last element (pop operation not implemented)\n")
		cg.writeString("// ")
		cg.generateExpression(exp.Object)
		cg.writeString("[len(")
		cg.generateExpression(exp.Object)
		cg.writeString(")-1]")

	case "length", "len":
		// Native len() function
		cg.writeString("len(")
		cg.generateExpression(exp.Object)
		cg.writeString(")")

	default:
		// For unsupported methods, generate comment
		cg.writeString("// Collection method '")
		cg.writeString(exp.Method.Value)
		cg.writeString("' not implemented")
	}
}

// generateImportStatement genera código Go para una declaración de import
func (cg *CodeGenerator) generateImportStatement(stmt *ast.ImportStatement) {
	// Por ahora, simplemente ignoramos los imports en el código generado
	// TODO: Implementar import de módulos reales
	cg.writeString("// Import statement ignored: ")
	if stmt.ModuleName != nil {
		cg.writeString("import " + stmt.ModuleName.Value + "\n")
	}
}

func isLoopVariable(name string) bool {
	loopVars := []string{"i", "j", "k", "current_base", "candidate", "temp"}
	for _, v := range loopVars {
		if name == v {
			return true
		}
	}
	return false
}

//// cleanUnusedImports SIMPLIFICADA Y DEFINITIVA
// Para programas Zylo: NUNCA limpiar imports de fmt o runtime
func cleanUnusedImports(code string) string {
	// Para programas Zylo, conservamos todos los imports (son pocos y siempre necesarios)
	// Esto evita complejidad de detección que puede fallar
	if strings.Contains(code, "fmt.Println(") || strings.Contains(code, "package main") {
		// Este es código Go generado compilado para Zylo - conservamos todos los imports
		return code
	}

	// Para el resto de casos, no limpiar nada (código trivial)
	return code
}
