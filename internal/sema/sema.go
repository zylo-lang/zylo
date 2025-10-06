package sema

import (
	"fmt"
	"strings"

	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/lexer"
)

// ZYLO ERRORS - Sistema profesional de errores de tipo
const (
	ZYLO_ERR_001_PARSER_ERROR      = "ZYLO_ERR_001: Error de sintaxis"
	ZYLO_ERR_002_VAR_UNDEFINED     = "ZYLO_ERR_002: Variable no definida"
	ZYLO_ERR_003_INCOMPATIBLE_TYPE = "ZYLO_ERR_003: Tipo incompatible"
	ZYLO_ERR_004_INVALID_INDEX     = "ZYLO_ERR_004: Índice de lista inválido"
	ZYLO_ERR_005_INVALID_MAP_KEY   = "ZYLO_ERR_005: Clave de mapa inválida"
	ZYLO_ERR_006_INVALID_ASSIGNMENT = "ZYLO_ERR_006: Asignación inválida"
	ZYLO_ERR_007_FUNCTION_ARGS     = "ZYLO_ERR_007: Parámetros de función inválidos"
	ZYLO_ERR_008_RETURN_TYPE       = "ZYLO_ERR_008: Tipo de retorno inválido"
	ZYLO_ERR_009_UNKNOWN_TYPE      = "ZYLO_ERR_009: Tipo desconocido"
	ZYLO_ERR_010_INVALID_OPERATION = "ZYLO_ERR_010: Operación inválida"
	ZYLO_ERR_011_TYPE_CASE         = "ZYLO_ERR_011: Tipos deben estar en minúscula"
	ZYLO_ERR_012_DUPLICATE_VAR     = "ZYLO_ERR_012: Variable ya declarada"
	ZYLO_ERR_013_FUNCTION_NOT_FOUND = "ZYLO_ERR_013: Función no encontrada"
	ZYLO_ERR_014_ACCESS_DENIED     = "ZYLO_ERR_014: Acceso denegado"
)

// ZyloError representa un error profesional con metadata completa
type ZyloError struct {
	Code          string
	Message       string
	Line          int
	Column       int
	Filename     string
	Expected     string
	Received     string
	Suggestion   string
	Severity     string // "error", "warning", "info"
	Context      string // additional context information
}

// Error implementa la interfaz error
func (z *ZyloError) Error() string {
	return fmt.Sprintf("%s - %s:%d:%d - %s", z.Code, z.Filename, z.Line, z.Column, z.Message)
}

// FullError retorna el error completo con sugerencia
func (z *ZyloError) FullError() string {
	result := z.Error()
	if z.Expected != "" || z.Received != "" {
		result += fmt.Sprintf("\n  Esperado: %s, Recibido: %s", z.Expected, z.Received)
	}
	if z.Suggestion != "" {
		result += fmt.Sprintf("\n  Sugerencia: %s", z.Suggestion)
	}
	if z.Context != "" {
		result += fmt.Sprintf("\n  Contexto: %s", z.Context)
	}
	return result
}

// ErrorBuilder crea errores profesionales
type ErrorBuilder struct {
	filename string
}

// NewErrorBuilder crea un nuevo builder de errores
func NewErrorBuilder(filename string) *ErrorBuilder {
	return &ErrorBuilder{filename: filename}
}

// SyntaxError crea error ZYLO_ERR_001
func (eb *ErrorBuilder) SyntaxError(token lexer.Token, message string) *ZyloError {
	return &ZyloError{
		Code:      ZYLO_ERR_001_PARSER_ERROR,
		Message:   message,
		Line:      token.StartLine,
		Column:   token.StartCol,
		Filename: eb.filename,
		Suggestion: "Revise la sintaxis según docs/syntax.md",
		Severity: "error",
	}
}

// UndefinedVarError crea error ZYLO_ERR_002
func (eb *ErrorBuilder) UndefinedVarError(token lexer.Token, varName string) *ZyloError {
	return &ZyloError{
		Code:        ZYLO_ERR_002_VAR_UNDEFINED,
		Message:     fmt.Sprintf("Variable '%s' no está definida", varName),
		Line:        token.StartLine,
		Column:     token.StartCol,
		Filename:   eb.filename,
		Suggestion: "Declare la variable antes de usarla o verifica si hay un error ortográfico",
		Severity:  "error",
	}
}

// IncompatibleTypeError crea error ZYLO_ERR_003
func (eb *ErrorBuilder) IncompatibleTypeError(token lexer.Token, expected, received string) *ZyloError {
	return &ZyloError{
		Code:      ZYLO_ERR_003_INCOMPATIBLE_TYPE,
		Message:   "Asignación de tipo incompatible",
		Line:      token.StartLine,
		Column:   token.StartCol,
		Filename: eb.filename,
		Expected: expected,
		Received: received,
		Suggestion: "Convierta el tipo explícitamente o cambie el tipo de la variable",
		Severity: "error",
	}
}

// TypeCaseError crea error ZYLO_ERR_011
func (eb *ErrorBuilder) TypeCaseError(token lexer.Token, wrongType string) *ZyloError {
	return &ZyloError{
		Code:      ZYLO_ERR_011_TYPE_CASE,
		Message:   fmt.Sprintf("Tipo '%s' debe estar en minúscula", wrongType),
		Line:      token.StartLine,
		Column:   token.StartCol,
		Filename: eb.filename,
		Expected: strings.ToLower(wrongType),
		Received: wrongType,
		Suggestion: "Use tipos en minúscula: int, float, string, bool",
		Severity: "error",
	}
}

// Type representa un tipo en el sistema de tipos de Zylo
type Type interface {
	String() string
	Equals(Type) bool
}

// PrimitiveType representa tipos primitivos
type PrimitiveType struct{ Name string }

func (t *PrimitiveType) String() string        { return t.Name }
func (t *PrimitiveType) Equals(other Type) bool {
	if o, ok := other.(*PrimitiveType); ok {
		return t.Name == o.Name
	}
	return false
}

// ListType representa tipos de lista
type ListType struct{ ElementType Type }

func (t *ListType) String() string { return fmt.Sprintf("List<%s>", t.ElementType.String()) }
func (t *ListType) Equals(other Type) bool {
	if o, ok := other.(*ListType); ok {
		return t.ElementType.Equals(o.ElementType)
	}
	return false
}

// MapType representa tipos de mapa
type MapType struct {
	KeyType   Type
	ValueType Type
}

func (t *MapType) String() string {
	return fmt.Sprintf("Map<%s, %s>", t.KeyType.String(), t.ValueType.String())
}
func (t *MapType) Equals(other Type) bool {
	if o, ok := other.(*MapType); ok {
		return t.KeyType.Equals(o.KeyType) && t.ValueType.Equals(o.ValueType)
	}
	return false
}

// FunctionType representa tipos de función
type FunctionType struct {
	ParamTypes []Type
	ReturnType Type
}

func (t *FunctionType) String() string { return "func" }
func (t *FunctionType) Equals(other Type) bool {
	if o, ok := other.(*FunctionType); ok {
		if len(t.ParamTypes) != len(o.ParamTypes) {
			return false
		}
		for i := range t.ParamTypes {
			if !t.ParamTypes[i].Equals(o.ParamTypes[i]) {
				return false
			}
		}
		return t.ReturnType.Equals(o.ReturnType)
	}
	return false
}

// ClassType representa tipos de clase
type ClassType struct {
	Name       string
	SuperClass *ClassType
	Methods    map[string]*FunctionType
	Fields     map[string]Type
	TypeParams []string
}

func (t *ClassType) String() string { return t.Name }
func (t *ClassType) Equals(other Type) bool {
	if o, ok := other.(*ClassType); ok {
		return t.Name == o.Name
	}
	return false
}

// AnyType representa el tipo any (top type)
type AnyType struct{}

func (t *AnyType) String() string        { return "any" }
func (t *AnyType) Equals(other Type) bool { _, ok := other.(*AnyType); return ok }

// Tipos primitivos globales
var (
	IntType    = &PrimitiveType{Name: "int"}
	FloatType  = &PrimitiveType{Name: "float"}
	StringType = &PrimitiveType{Name: "string"}
	BoolType   = &PrimitiveType{Name: "bool"}
	NullType   = &PrimitiveType{Name: "nil"}
	Any        = &AnyType{}
)

// Symbol representa una entrada en la tabla de símbolos
type Symbol struct {
	Name  string
	Type  Type
	Scope string
}

// SymbolTable representa una tabla de símbolos
type SymbolTable struct {
	parent       *SymbolTable
	symbols      map[string]*Symbol
	scopeName    string
	scopeLevel   int
	isFunction   bool
	capturedVars map[string]*Symbol
}

// NewSymbolTable crea una nueva tabla de símbolos
func NewSymbolTable(scopeName string, level int, parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		parent:       parent,
		symbols:      make(map[string]*Symbol),
		scopeName:    scopeName,
		scopeLevel:   level,
		isFunction:   false,
		capturedVars: make(map[string]*Symbol),
	}
}

// NewFunctionSymbolTable crea una tabla de símbolos para función
func NewFunctionSymbolTable(scopeName string, level int, parent *SymbolTable) *SymbolTable {
	st := NewSymbolTable(scopeName, level, parent)
	st.isFunction = true
	return st
}

// Define añade un símbolo
func (st *SymbolTable) Define(name string, t Type) *Symbol {
	symbol := &Symbol{
		Name:  name,
		Type:  t,
		Scope: fmt.Sprintf("%s (Level %d)", st.scopeName, st.scopeLevel),
	}
	st.symbols[name] = symbol
	return symbol
}

// Resolve busca un símbolo
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	if sym, ok := st.symbols[name]; ok {
		return sym, true
	}
	if sym, ok := st.capturedVars[name]; ok {
		return sym, true
	}
	if st.parent != nil {
		if sym, ok := st.parent.Resolve(name); ok {
			if st.isFunction && sym.Scope != st.scopeName {
				st.capturedVars[name] = sym
			}
			return sym, true
		}
	}
	return nil, false
}

// SemanticAnalyzer realiza análisis semántico
type SemanticAnalyzer struct {
	symbolTable     *SymbolTable
	zyloErrors      []*ZyloError
	currentFunction *FunctionType
	inAsyncContext  bool
	inLoop          bool
	errorBuilder    *ErrorBuilder
}

// NewSemanticAnalyzer crea un analizador semántico
func NewSemanticAnalyzer() *SemanticAnalyzer {
	globalScope := NewSymbolTable("global", 0, nil)

	// Built-in functions
	globalScope.Define("show.log", &FunctionType{
		ParamTypes: []Type{Any}, // Variadic - accepts any number of arguments
		ReturnType: NullType,
	})
	// Crear módulo "show" que contiene funciones de logging
	showModule := &ClassType{
		Name: "show",
		Methods: map[string]*FunctionType{
			"log": {ParamTypes: []Type{Any}, ReturnType: NullType}, // Variadic
		},
		Fields:  make(map[string]Type),
	}
	globalScope.Define("show", showModule)
	globalScope.Define("print", &FunctionType{
		ParamTypes: []Type{Any},
		ReturnType: NullType,
	})
	globalScope.Define("read.line", &FunctionType{
		ParamTypes: []Type{},
		ReturnType: StringType,
	})
	globalScope.Define("read.int", &FunctionType{
		ParamTypes: []Type{},
		ReturnType: IntType,
	})
	globalScope.Define("string", &FunctionType{
		ParamTypes: []Type{Any},
		ReturnType: StringType,
	})
	globalScope.Define("println", &FunctionType{
		ParamTypes: []Type{Any}, // Variadic
		ReturnType: NullType,
	})
	globalScope.Define("len", &FunctionType{
		ParamTypes: []Type{Any},
		ReturnType: IntType,
	})
	globalScope.Define("split", &FunctionType{
		ParamTypes: []Type{StringType, StringType},
		ReturnType: &ListType{ElementType: StringType},
	})
	globalScope.Define("to_number", &FunctionType{
		ParamTypes: []Type{StringType},
		ReturnType: FloatType,
	})

	return &SemanticAnalyzer{
		symbolTable:     globalScope,
		zyloErrors:      []*ZyloError{},
		inAsyncContext:  false,
		inLoop:          false,
		errorBuilder:    NewErrorBuilder("analysis"),
	}
}

// Analyze ejecuta el análisis semántico
func (sa *SemanticAnalyzer) Analyze(node ast.Node) Type {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			sa.Analyze(stmt)
		}
		return nil

	case *ast.VarStatement:
		return sa.analyzeVarStatement(n)

	case *ast.ImportStatement:
		return sa.analyzeImportStatement(n)
	case *ast.FuncStatement:
		return sa.analyzeFuncStatement(n)

	case *ast.ReturnStatement:
		return sa.analyzeReturnStatement(n)

	case *ast.IfStatement:
		return sa.analyzeIfStatement(n)

	case *ast.WhileStatement:
		return sa.analyzeWhileStatement(n)

	case *ast.ForStatement:
		return sa.analyzeForStatement(n)
	case *ast.CollectionMethodCall:
		return sa.analyzeCollectionMethodCall(n)
	case *ast.ForInStatement:
		return sa.analyzeForInStatement(n)

	case *ast.BreakStatement:
		if !sa.inLoop {
			sa.addError(n.Token, "break solo puede usarse dentro de un bucle")
		}
		return nil

	case *ast.ContinueStatement:
		if !sa.inLoop {
			sa.addError(n.Token, "continue solo puede usarse dentro de un bucle")
		}
		return nil

	case *ast.ClassStatement:
		return sa.analyzeClassStatement(n)

	case *ast.ExpressionStatement:
		if n.Expression != nil {
			return sa.Analyze(n.Expression)
		}
		return nil

	case *ast.BlockStatement:
		sa.enterScope("block")
		for _, stmt := range n.Statements {
			sa.Analyze(stmt)
		}
		sa.exitScope()
		return nil

	// Expressions
	case *ast.Identifier:
		return sa.analyzeIdentifier(n)

	case *ast.NumberLiteral:
		if _, ok := n.Value.(int64); ok {
			return IntType
		}
		return FloatType

	case *ast.StringLiteral:
		return StringType

	case *ast.BooleanLiteral:
		return BoolType

	case *ast.NullLiteral:
		return NullType

	case *ast.ListLiteral:
		return sa.analyzeListLiteral(n)

	case *ast.MapLiteral:
		return sa.analyzeMapLiteral(n)

	case *ast.CallExpression:
		return sa.analyzeCallExpression(n)

	case *ast.DotExpression:
		return sa.analyzeDotExpression(n)

	case *ast.IndexExpression:
		return sa.analyzeIndexExpression(n)

	case *ast.InfixExpression:
		return sa.analyzeInfixExpression(n)

	case *ast.PrefixExpression:
		return sa.analyzePrefixExpression(n)

	case *ast.AssignmentExpression:
		return sa.analyzeAssignmentExpression(n)

	default:
		return Any
	}
}

// analyzeVarStatement analiza declaración de variable
func (sa *SemanticAnalyzer) analyzeVarStatement(stmt *ast.VarStatement) Type {
	var expectedType Type = Any

	if stmt.Name.TypeAnnotation != "" {
		expectedType = sa.stringToType(stmt.Name.Token, stmt.Name.TypeAnnotation)
	}

	var valueType Type = NullType
	if stmt.Value != nil {
		valueType = sa.Analyze(stmt.Value)
	}

	if expectedType == Any {
		expectedType = valueType
	}

	if !sa.isAssignable(expectedType, valueType) {
		sa.addError(stmt.Token, fmt.Sprintf("no se puede asignar %s a variable de tipo %s", valueType, expectedType))
	}

	sa.symbolTable.Define(stmt.Name.Value, expectedType)
	return nil
}

// analyzeFuncStatement analiza declaración de función
func (sa *SemanticAnalyzer) analyzeFuncStatement(stmt *ast.FuncStatement) Type {
	paramTypes := make([]Type, len(stmt.Parameters))
	for i, p := range stmt.Parameters {
		if p.TypeAnnotation != "" {
			paramTypes[i] = sa.stringToType(p.Token, p.TypeAnnotation)
		} else {
			paramTypes[i] = Any
		}
	}

	var returnType Type = Any
	if stmt.ReturnType != "" {
		returnType = sa.stringToType(stmt.Token, stmt.ReturnType)
	}

	funcType := &FunctionType{ParamTypes: paramTypes, ReturnType: returnType}
	sa.symbolTable.Define(stmt.Name.Value, funcType)

	sa.enterFunctionScope(stmt.Name.Value)
	previousFunction := sa.currentFunction
	sa.currentFunction = funcType

	for i, p := range stmt.Parameters {
		sa.symbolTable.Define(p.Value, paramTypes[i])
	}

	sa.Analyze(stmt.Body)

	sa.currentFunction = previousFunction
	sa.exitFunctionScope()
	return nil
}

// analyzeReturnStatement analiza return
func (sa *SemanticAnalyzer) analyzeReturnStatement(stmt *ast.ReturnStatement) Type {
	if sa.currentFunction == nil {
		sa.addError(stmt.Token, "return fuera de función")
		return nil
	}

	if stmt.ReturnValue != nil {
		valueType := sa.Analyze(stmt.ReturnValue)
		if !sa.isAssignable(sa.currentFunction.ReturnType, valueType) {
			sa.addError(stmt.Token, fmt.Sprintf("tipo de retorno incorrecto: esperado %s, obtenido %s", sa.currentFunction.ReturnType, valueType))
		}
	} else {
		if sa.currentFunction.ReturnType != NullType && sa.currentFunction.ReturnType != Any {
			sa.addError(stmt.Token, fmt.Sprintf("función espera retorno de tipo %s", sa.currentFunction.ReturnType))
		}
	}
	return nil
}

// analyzeIfStatement analiza if
func (sa *SemanticAnalyzer) analyzeIfStatement(stmt *ast.IfStatement) Type {
	condType := sa.Analyze(stmt.Condition)
	if condType != BoolType && condType != Any {
		sa.addError(stmt.Token, "condición debe ser booleana")
	}

	sa.Analyze(stmt.Consequence)
	if stmt.Alternative != nil {
		sa.Analyze(stmt.Alternative)
	}
	return nil
}

// analyzeWhileStatement analiza while
func (sa *SemanticAnalyzer) analyzeWhileStatement(stmt *ast.WhileStatement) Type {
	condType := sa.Analyze(stmt.Condition)
	if condType != BoolType && condType != Any {
		sa.addError(stmt.Token, "condición debe ser booleana")
	}

	wasInLoop := sa.inLoop
	sa.inLoop = true
	sa.Analyze(stmt.Body)
	sa.inLoop = wasInLoop
	return nil
}

// analyzeForStatement analiza bucle for tradicional
func (sa *SemanticAnalyzer) analyzeForStatement(stmt *ast.ForStatement) Type {
	// Analizar la inicialización
	if stmt.Init != nil {
		sa.Analyze(stmt.Init)
	}

	// Analizar condición
	if stmt.Condition != nil {
		condType := sa.Analyze(stmt.Condition)
		if condType != BoolType && condType != Any {
			sa.addError(stmt.Token, "condición del for debe ser booleana")
		}
	}

	// Analizar el post statement (incremento)
	if stmt.Post != nil {
		sa.Analyze(stmt.Post)
	}

	// Analizar cuerpo del bucle
	wasInLoop := sa.inLoop
	sa.inLoop = true
	sa.Analyze(stmt.Body)
	sa.inLoop = wasInLoop

	return nil
}

// analyzeForInStatement analiza for-in
func (sa *SemanticAnalyzer) analyzeForInStatement(stmt *ast.ForInStatement) Type {
	iterableType := sa.Analyze(stmt.Iterable)

	var elementType Type = Any
	if listType, ok := iterableType.(*ListType); ok {
		elementType = listType.ElementType
	} else if iterableType == StringType {
		elementType = StringType
	} else if iterableType != Any {
		sa.addError(stmt.Token, "for-in requiere lista o string")
	}

	sa.enterScope("for-in")
	sa.symbolTable.Define(stmt.Identifier.Value, elementType)

	wasInLoop := sa.inLoop
	sa.inLoop = true
	sa.Analyze(stmt.Body)
	sa.inLoop = wasInLoop

	sa.exitScope()
	return nil
}

// analyzeClassStatement analiza clase
func (sa *SemanticAnalyzer) analyzeClassStatement(stmt *ast.ClassStatement) Type {
	classType := &ClassType{
		Name:    stmt.Name.Value,
		Methods: make(map[string]*FunctionType),
		Fields:  make(map[string]Type),
	}

	if stmt.SuperClass != nil {
		if superSym, ok := sa.symbolTable.Resolve(stmt.SuperClass.Value); ok {
			if superClass, ok := superSym.Type.(*ClassType); ok {
				classType.SuperClass = superClass
				for name, method := range superClass.Methods {
					classType.Methods[name] = method
				}
				for name, field := range superClass.Fields {
					classType.Fields[name] = field
				}
			}
		}
	}

	sa.enterScope(stmt.Name.Value)

	for _, attr := range stmt.Attributes {
		var attrType Type = Any
		if attr.Name.TypeAnnotation != "" {
			attrType = sa.stringToType(attr.Name.Token, attr.Name.TypeAnnotation)
		} else if attr.Value != nil {
			attrType = sa.Analyze(attr.Value)
		}
		classType.Fields[attr.Name.Value] = attrType
		sa.symbolTable.Define(attr.Name.Value, attrType)
	}

	for _, method := range stmt.Methods {
		paramTypes := make([]Type, len(method.Parameters))
		for i, p := range method.Parameters {
			if p.TypeAnnotation != "" {
				paramTypes[i] = sa.stringToType(p.Token, p.TypeAnnotation)
			} else {
				paramTypes[i] = Any
			}
		}

		var returnType Type = Any
		if method.ReturnType != "" {
			returnType = sa.stringToType(method.Token, method.ReturnType)
		}

		funcType := &FunctionType{ParamTypes: paramTypes, ReturnType: returnType}
		classType.Methods[method.Name.Value] = funcType
	}

	sa.exitScope()
	sa.symbolTable.Define(stmt.Name.Value, classType)
	return nil
}

// analyzeIdentifier analiza identificador
func (sa *SemanticAnalyzer) analyzeIdentifier(exp *ast.Identifier) Type {
	if sym, ok := sa.symbolTable.Resolve(exp.Value); ok {
		return sym.Type
	}
	sa.addError(exp.Token, fmt.Sprintf("variable no definida: %s", exp.Value))
	return Any
}

// analyzeListLiteral analiza literal de lista
func (sa *SemanticAnalyzer) analyzeListLiteral(exp *ast.ListLiteral) Type {
	if len(exp.Elements) == 0 {
		return &ListType{ElementType: Any}
	}

	firstType := sa.Analyze(exp.Elements[0])
	for _, elem := range exp.Elements[1:] {
		elemType := sa.Analyze(elem)
		if !firstType.Equals(elemType) && elemType != Any && firstType != Any {
			return &ListType{ElementType: Any}
		}
	}
	return &ListType{ElementType: firstType}
}

// analyzeMapLiteral analiza literal de mapa
func (sa *SemanticAnalyzer) analyzeMapLiteral(exp *ast.MapLiteral) Type {
	if len(exp.Pairs) == 0 {
		return &MapType{KeyType: Any, ValueType: Any}
	}

	var keyType, valueType Type = StringType, Any
	for _, v := range exp.Pairs {
		vType := sa.Analyze(v)
		if valueType == Any {
			valueType = vType
		}
	}
	return &MapType{KeyType: keyType, ValueType: valueType}
}

// analyzeCallExpression analiza llamada a función
func (sa *SemanticAnalyzer) analyzeCallExpression(exp *ast.CallExpression) Type {
	funcType := sa.Analyze(exp.Function)

	if ft, ok := funcType.(*FunctionType); ok {
		// Handle variadic functions (show.log accepts any number of Any arguments)
		if len(ft.ParamTypes) == 1 && ft.ParamTypes[0] == Any {
			// Variadic function - all arguments are accepted as Any
			for _, arg := range exp.Arguments {
				sa.Analyze(arg) // Just analyze for side effects
			}
		} else {
			// Regular function - check argument count and types
			if len(exp.Arguments) != len(ft.ParamTypes) {
				sa.addError(exp.Token, fmt.Sprintf("esperados %d argumentos, recibidos %d", len(ft.ParamTypes), len(exp.Arguments)))
			} else {
				for i, arg := range exp.Arguments {
					argType := sa.Analyze(arg)
					if !sa.isAssignable(ft.ParamTypes[i], argType) {
						sa.addError(exp.Token, fmt.Sprintf("argumento %d: esperado %s, obtenido %s", i+1, ft.ParamTypes[i], argType))
					}
				}
			}
		}
		return ft.ReturnType
	}

	if classType, ok := funcType.(*ClassType); ok {
		return classType
	}

	return Any
}

// analyzeDotExpression analiza expresión de punto
func (sa *SemanticAnalyzer) analyzeDotExpression(exp *ast.DotExpression) Type {
	objType := sa.Analyze(exp.Left)

	if classType, ok := objType.(*ClassType); ok {
		// Check if this is an imported module (e.g., math.sqrt)
		if _, exists := classType.Methods[exp.Property.Value]; exists {
			return classType.Methods[exp.Property.Value]
		}
		if _, exists := classType.Fields[exp.Property.Value]; exists {
			return classType.Fields[exp.Property.Value]
		}

		// For modules like 'math', we don't have specific function types defined yet
		// So we return a generic function type for math functions
		if objIdent, ok := exp.Left.(*ast.Identifier); ok {
			if sym, exists := sa.symbolTable.Resolve(objIdent.Value); exists {
				if _, isModule := sym.Type.(*ClassType); isModule {
					// This is a module function call, return a generic function type
					return &FunctionType{
						ParamTypes: []Type{Any}, // Generic parameter
						ReturnType: Any,         // Generic return type
					}
				}
			}
		}
	}

	return Any
}

// analyzeIndexExpression analiza indexación
func (sa *SemanticAnalyzer) analyzeIndexExpression(exp *ast.IndexExpression) Type {
	leftType := sa.Analyze(exp.Left)
	indexType := sa.Analyze(exp.Index)

	if indexType != IntType && indexType != Any {
		sa.addError(exp.Token, "índice debe ser entero")
	}

	if listType, ok := leftType.(*ListType); ok {
		return listType.ElementType
	}
	if mapType, ok := leftType.(*MapType); ok {
		return mapType.ValueType
	}
	if leftType == StringType {
		return StringType
	}

	return Any
}

// analyzeInfixExpression analiza expresión infija
func (sa *SemanticAnalyzer) analyzeInfixExpression(exp *ast.InfixExpression) Type {
	leftType := sa.Analyze(exp.Left)
	rightType := sa.Analyze(exp.Right)

	if !sa.areTypesCompatible(leftType, rightType, exp.Operator) {
		sa.addError(exp.Token, fmt.Sprintf("operador '%s' no válido para %s y %s", exp.Operator, leftType, rightType))
	}

	return sa.inferInfixReturnType(leftType, rightType, exp.Operator)
}

// analyzePrefixExpression analiza expresión prefija
func (sa *SemanticAnalyzer) analyzePrefixExpression(exp *ast.PrefixExpression) Type {
	rightType := sa.Analyze(exp.Right)

	switch exp.Operator {
	case "!":
		return BoolType
	case "-":
		if rightType == IntType || rightType == FloatType || rightType == Any {
			return rightType
		}
		sa.addError(exp.Token, "operador '-' requiere número")
		return Any
	}

	return Any
}

// analyzeAssignmentExpression analiza asignación
func (sa *SemanticAnalyzer) analyzeAssignmentExpression(exp *ast.AssignmentExpression) Type {
	targetType := sa.Analyze(exp.Name)
	valueType := sa.Analyze(exp.Value)

	if !sa.isAssignable(targetType, valueType) {
		sa.addError(exp.Token, fmt.Sprintf("no se puede asignar %s a %s", valueType, targetType))
	}

	return targetType
}

// Helper functions

func (sa *SemanticAnalyzer) stringToType(token lexer.Token, typeStr string) Type {
	// Handle List<T>
	if strings.HasPrefix(typeStr, "List<") && strings.HasSuffix(typeStr, ">") {
		innerType := typeStr[5 : len(typeStr)-1]
		return &ListType{ElementType: sa.stringToType(token, innerType)}
	}

	// Handle Map<K, V>
	if strings.HasPrefix(typeStr, "Map<") && strings.HasSuffix(typeStr, ">") {
		inner := typeStr[4 : len(typeStr)-1]
		parts := strings.SplitN(inner, ",", 2)
		if len(parts) == 2 {
			keyType := sa.stringToType(token, strings.TrimSpace(parts[0]))
			valueType := sa.stringToType(token, strings.TrimSpace(parts[1]))
			return &MapType{KeyType: keyType, ValueType: valueType}
		}
	}

	switch typeStr {
	case "int", "Int":
		return IntType
	case "float", "Float":
		return FloatType
	case "string", "String":
		return StringType
	case "bool", "Bool":
		return BoolType
	case "nil":
		return NullType
	case "ANY":
		return Any
	case "":
		return Any
	default:
		if sym, ok := sa.symbolTable.Resolve(typeStr); ok {
			return sym.Type
		}
		sa.addError(token, fmt.Sprintf("tipo desconocido: %s", typeStr))
		return Any
	}
}

func (sa *SemanticAnalyzer) isAssignable(target, value Type) bool {
	if target == Any || value == Any {
		return true
	}
	if target.Equals(value) {
		return true
	}

	// Int puede asignarse a Float
	if target == FloatType && value == IntType {
		return true
	}

	return false
}

func (sa *SemanticAnalyzer) areTypesCompatible(left, right Type, op string) bool {
	if left == Any || right == Any {
		return true
	}

	switch op {
	case "+":
		if left == StringType || right == StringType {
			return true
		}
		return sa.isNumericType(left) && sa.isNumericType(right)
	case "-", "*", "/", "%", "**", "//":
		return sa.isNumericType(left) && sa.isNumericType(right)
	case "==", "!=":
		return true
	case "<", "<=", ">", ">=":
		return sa.isNumericType(left) && sa.isNumericType(right)
	case "and", "or", "&&", "||":
		return true
	}

	return left.Equals(right)
}

func (sa *SemanticAnalyzer) isNumericType(t Type) bool {
	return t == IntType || t == FloatType
}

func (sa *SemanticAnalyzer) inferInfixReturnType(left, right Type, op string) Type {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=", "and", "or", "&&", "||":
		return BoolType
	case "+":
		if left == StringType || right == StringType {
			return StringType
		}
		if left == FloatType || right == FloatType {
			return FloatType
		}
		return IntType
	case "-", "*", "/", "%", "**", "//":
		if left == FloatType || right == FloatType {
			return FloatType
		}
		return IntType
	}
	return Any
}

func (sa *SemanticAnalyzer) enterScope(name string) {
	newScope := NewSymbolTable(name, sa.symbolTable.scopeLevel+1, sa.symbolTable)
	sa.symbolTable = newScope
}

func (sa *SemanticAnalyzer) exitScope() {
	if sa.symbolTable.parent != nil {
		sa.symbolTable = sa.symbolTable.parent
	}
}

func (sa *SemanticAnalyzer) enterFunctionScope(name string) {
	newScope := NewFunctionSymbolTable(name, sa.symbolTable.scopeLevel+1, sa.symbolTable)
	sa.symbolTable = newScope
}

func (sa *SemanticAnalyzer) exitFunctionScope() {
	sa.exitScope()
}

// GetSymbolTable retorna la tabla de símbolos
func (sa *SemanticAnalyzer) GetSymbolTable() *SymbolTable {
	return sa.symbolTable
}

// ZyloErrors retorna los errores ZyloError
func (sa *SemanticAnalyzer) ZyloErrors() []*ZyloError {
	return sa.zyloErrors
}

// Errors retorna los errores como strings (para compatibilidad)
func (sa *SemanticAnalyzer) Errors() []string {
	strings := make([]string, len(sa.zyloErrors))
	for i, zyloErr := range sa.zyloErrors {
		strings[i] = zyloErr.FullError()
	}
	return strings
}

// addError agrega un ZyloError
func (sa *SemanticAnalyzer) addError(token lexer.Token, msg string) {
	error := sa.errorBuilder.IncompatibleTypeError(token, "esperado", "recibido")
	error.Message = msg
	sa.zyloErrors = append(sa.zyloErrors, error)
}

// addZyloError agrega un ZyloError directo
func (sa *SemanticAnalyzer) addZyloError(error *ZyloError) {
	sa.zyloErrors = append(sa.zyloErrors, error)
}

// analyzeImportStatement analiza declaración de import con resolución avanzada de módulos
func (sa *SemanticAnalyzer) analyzeImportStatement(stmt *ast.ImportStatement) Type {
	var moduleType *ClassType

	if stmt.ModuleName != nil {
		// Import simple de módulo (e.g., import math)
		moduleType = &ClassType{
			Name:    stmt.ModuleName.Value,
			Methods: make(map[string]*FunctionType),
			Fields:  make(map[string]Type),
		}

		// Resolver módulo de la stdlib si existe
		if resolved := sa.resolveStdLibModule(stmt.ModuleName.Value); resolved != nil {
			// Copiar métodos y campos del módulo resuelto
			for k, v := range resolved.Methods {
				moduleType.Methods[k] = v
			}
			for k, v := range resolved.Fields {
				moduleType.Fields[k] = v
			}
		}

		sa.symbolTable.Define(stmt.ModuleName.Value, moduleType)
	} else if stmt.ModulePath != "" {
		// Import de path (e.g., import "std/math" or "./local/module")
		// Intentar resolver tanto stdlib como local paths
		if resolved := sa.resolveModulePath(stmt.ModulePath); resolved != nil {
			moduleType = resolved
			// Para paths, usar el nombre del archivo como nombre del módulo
			parts := strings.Split(stmt.ModulePath, "/")
			if len(parts) > 0 {
				moduleName := strings.TrimSuffix(parts[len(parts)-1], ".zylo")
				if moduleName == "" {
					moduleName = parts[len(parts)-1]
				}
				sa.symbolTable.Define(moduleName, moduleType)
			}
		} else {
			sa.addError(stmt.Token, fmt.Sprintf("Módulo no encontrado: %s", stmt.ModulePath))
			return Any
		}
	}

	// Return the resolved module type for statement semantics
	return moduleType
}

// resolveStdLibModule resuelve un módulo de la biblioteca estándar
func (sa *SemanticAnalyzer) resolveStdLibModule(moduleName string) *ClassType {
	switch moduleName {
	case "math":
		return &ClassType{
			Name: "math",
			Methods: map[string]*FunctionType{
				"sqrt":    {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"power":   {ParamTypes: []Type{FloatType, FloatType}, ReturnType: FloatType},
				"abs":     {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"floor":   {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"ceil":    {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"round":   {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"sin":     {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"cos":     {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"tan":     {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"factorial": {ParamTypes: []Type{IntType}, ReturnType: IntType},
				"gcd":       {ParamTypes: []Type{IntType, IntType}, ReturnType: IntType},
				"lcm":       {ParamTypes: []Type{IntType, IntType}, ReturnType: IntType},
				"is_prime":  {ParamTypes: []Type{IntType}, ReturnType: BoolType},
				"fibonacci_iterative": {ParamTypes: []Type{IntType}, ReturnType: IntType},
				"degrees_to_radians":  {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"radians_to_degrees":  {ParamTypes: []Type{FloatType}, ReturnType: FloatType},
				"clamp":    {ParamTypes: []Type{FloatType, FloatType, FloatType}, ReturnType: FloatType},
				"lerp":     {ParamTypes: []Type{FloatType, FloatType, FloatType}, ReturnType: FloatType},
				"map_range": {ParamTypes: []Type{FloatType, FloatType, FloatType, FloatType, FloatType}, ReturnType: FloatType},
				"add":      {ParamTypes: []Type{FloatType, FloatType}, ReturnType: FloatType},
				"subtract": {ParamTypes: []Type{FloatType, FloatType}, ReturnType: FloatType},
				"multiply": {ParamTypes: []Type{FloatType, FloatType}, ReturnType: FloatType},
				"divide":   {ParamTypes: []Type{FloatType, FloatType}, ReturnType: FloatType},
			},
			Fields: map[string]Type{
				"PI":  FloatType,
				"E":   FloatType,
				"TAU": FloatType,
				"PHI": FloatType,
			},
		}
	case "string":
		return &ClassType{
			Name: "string",
			Methods: map[string]*FunctionType{
				"split":     {ParamTypes: []Type{StringType, StringType}, ReturnType: &ListType{ElementType: StringType}},
				"join":      {ParamTypes: []Type{&ListType{ElementType: StringType}, StringType}, ReturnType: StringType},
				"substring": {ParamTypes: []Type{StringType, IntType, IntType}, ReturnType: StringType},
				"replace":   {ParamTypes: []Type{StringType, StringType, StringType}, ReturnType: StringType},
				"trim":      {ParamTypes: []Type{StringType}, ReturnType: StringType},
				"to_upper":  {ParamTypes: []Type{StringType}, ReturnType: StringType},
				"to_lower":  {ParamTypes: []Type{StringType}, ReturnType: StringType},
				"contains":  {ParamTypes: []Type{StringType, StringType}, ReturnType: BoolType},
				"starts_with": {ParamTypes: []Type{StringType, StringType}, ReturnType: BoolType},
				"ends_with": {ParamTypes: []Type{StringType, StringType}, ReturnType: BoolType},
			},
			Fields: make(map[string]Type),
		}
	case "json":
		return &ClassType{
			Name: "json",
			Methods: map[string]*FunctionType{
				"parse": {ParamTypes: []Type{StringType}, ReturnType: Any},
				"stringify": {ParamTypes: []Type{Any}, ReturnType: StringType},
			},
			Fields: make(map[string]Type),
		}
	case "io":
		return &ClassType{
			Name: "io",
			Methods: map[string]*FunctionType{
				"read_file": {ParamTypes: []Type{StringType}, ReturnType: StringType},
				"write_file": {ParamTypes: []Type{StringType, StringType}, ReturnType: BoolType},
				"read_line": {ParamTypes: []Type{}, ReturnType: StringType},
			},
			Fields: make(map[string]Type),
		}
	case "time":
		return &ClassType{
			Name: "time",
			Methods: map[string]*FunctionType{
				"now":       {ParamTypes: []Type{}, ReturnType: StringType},
				"parse":     {ParamTypes: []Type{StringType, StringType}, ReturnType: StringType},
				"format":    {ParamTypes: []Type{StringType, StringType}, ReturnType: StringType},
				"add_days":  {ParamTypes: []Type{StringType, IntType}, ReturnType: StringType},
				"add_hours": {ParamTypes: []Type{StringType, IntType}, ReturnType: StringType},
				"diff_days": {ParamTypes: []Type{StringType, StringType}, ReturnType: IntType},
			},
			Fields: make(map[string]Type),
		}
	case "list":
		return &ClassType{
			Name: "list",
			Methods: map[string]*FunctionType{
				"push":     {ParamTypes: []Type{&ListType{ElementType: Any}, Any}, ReturnType: &ListType{ElementType: Any}},
				"pop":      {ParamTypes: []Type{&ListType{ElementType: Any}}, ReturnType: Any},
				"shift":    {ParamTypes: []Type{&ListType{ElementType: Any}}, ReturnType: Any},
				"unshift":  {ParamTypes: []Type{&ListType{ElementType: Any}, Any}, ReturnType: &ListType{ElementType: Any}},
				"slice":    {ParamTypes: []Type{&ListType{ElementType: Any}, IntType, IntType}, ReturnType: &ListType{ElementType: Any}},
				"sort":     {ParamTypes: []Type{&ListType{ElementType: Any}}, ReturnType: &ListType{ElementType: Any}},
				"reverse":  {ParamTypes: []Type{&ListType{ElementType: Any}}, ReturnType: &ListType{ElementType: Any}},
				"concat":   {ParamTypes: []Type{&ListType{ElementType: Any}, &ListType{ElementType: Any}}, ReturnType: &ListType{ElementType: Any}},
				"includes": {ParamTypes: []Type{&ListType{ElementType: Any}, Any}, ReturnType: BoolType},
				"index_of": {ParamTypes: []Type{&ListType{ElementType: Any}, Any}, ReturnType: IntType},
			},
			Fields: make(map[string]Type),
		}
	case "map":
		return &ClassType{
			Name: "map",
			Methods: map[string]*FunctionType{
				"set":     {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}, StringType, Any}, ReturnType: &MapType{KeyType: StringType, ValueType: Any}},
				"get":     {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}, StringType}, ReturnType: Any},
				"has":     {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}, StringType}, ReturnType: BoolType},
				"delete":  {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}, StringType}, ReturnType: &MapType{KeyType: StringType, ValueType: Any}},
				"clear":   {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}}, ReturnType: &MapType{KeyType: StringType, ValueType: Any}},
				"keys":    {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}}, ReturnType: &ListType{ElementType: StringType}},
				"values":  {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}}, ReturnType: &ListType{ElementType: Any}},
				"entries": {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}}, ReturnType: &ListType{ElementType: Any}},
				"size":    {ParamTypes: []Type{&MapType{KeyType: StringType, ValueType: Any}}, ReturnType: IntType},
			},
			Fields: make(map[string]Type),
		}
	default:
		return nil // Module not found in stdlib
	}
}

// resolveModulePath resuelve un módulo desde una ruta de archivo
func (sa *SemanticAnalyzer) resolveModulePath(modulePath string) *ClassType {
	// TODO: Implement file system resolution for local modules
	// For now, support basic std/ path resolution
	if strings.HasPrefix(modulePath, "std/") {
		stdModuleName := strings.TrimPrefix(modulePath, "std/")
		stdModuleName = strings.TrimSuffix(stdModuleName, ".zylo")
		return sa.resolveStdLibModule(stdModuleName)
	}
	// For other paths, return nil to indicate not found
	return nil
}

// analyzeCollectionMethodCall analiza llamada a método de colección o función de módulo
func (sa *SemanticAnalyzer) analyzeCollectionMethodCall(exp *ast.CollectionMethodCall) Type {
	// First check if this is a module function call (e.g., math.sqrt(4))
	objType := sa.Analyze(exp.Object)

	if _, ok := objType.(*ClassType); ok {
		// This is a module function call (e.g., math.sqrt(x))
		// For now, accept any function call on modules
		// TODO: Add proper validation for specific module functions

		// Analyze arguments
		for _, arg := range exp.Arguments {
			sa.Analyze(arg)
		}

		// Return appropriate type based on method name
		switch exp.Method.Value {
		case "sqrt", "abs", "floor", "ceil", "round", "sin", "cos", "tan":
			// Math functions typically return float
			return FloatType
		case "add", "subtract", "multiply", "divide", "power":
			// Arithmetic operations return float
			return FloatType
		case "factorial", "gcd", "lcm":
			// Integer math functions
			return IntType
		case "is_prime", "is_even", "is_odd":
			// Boolean functions
			return BoolType
		default:
			// Default to any for unknown module functions
			return Any
		}
	}

	// This is a collection method call (e.g., arr.push(element))
	var methods map[string]bool

	// Definir métodos válidos para cada tipo de colección
	if _, isList := objType.(*ListType); isList || objType == Any {
		// Métodos disponibles para listas
		methods = map[string]bool{
			"push": true, "pop": true, "shift": true, "unshift": true,
			"splice": true, "forEach": true, "map": true, "filter": true,
			"find": true, "some": true, "every": true, "indexOf": true,
			"includes": true, "join": true, "slice": true, "reverse": true,
			"sort": true, "concat": true, "length": true,
		}
	} else if _, isMap := objType.(*MapType); isMap || objType == Any {
		// Métodos disponibles para mapas
		methods = map[string]bool{
			"set": true, "get": true, "has": true, "delete": true,
			"clear": true, "keys": true, "values": true, "entries": true,
			"forEach": true, "size": true,
		}
	} else {
		sa.addError(exp.Token, fmt.Sprintf("El objeto no es una colección válida para método '%s'", exp.Method.Value))
		return Any
	}

	// Verificar que el método existe
	if !methods[exp.Method.Value] {
		sa.addError(exp.Token, fmt.Sprintf("Método '%s' no existe en este tipo de colección", exp.Method.Value))
		return Any
	}

	// Analizar argumentos para validación de tipo básica
	for _, arg := range exp.Arguments {
		sa.Analyze(arg)
	}

	// Retornar tipo basado en el método (simplificado)
	switch exp.Method.Value {
	case "pop", "shift", "get":
		if listType, ok := objType.(*ListType); ok {
			return listType.ElementType
		}
		if mapType, ok := objType.(*MapType); ok {
			return mapType.ValueType
		}
		return Any
	case "push", "unshift", "splice", "reverse", "sort", "set", "delete", "clear":
		// Estos métodos modifican la colección y pueden retornar la colección o void
		return objType
	case "indexOf", "size", "length":
		return IntType
	case "includes", "has", "some", "every":
		return BoolType
	case "slice", "filter", "map", "concat", "keys", "values", "entries", "join":
		// Estos retornan una nueva colección
		return objType
	case "find", "forEach":
		return Any
	default:
		return Any
	}
}
