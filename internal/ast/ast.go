package ast

import (
	"fmt"
	"strings"

	"github.com/zylo-lang/zylo/internal/lexer"
)

// Node es la interfaz base para todos los nodos del AST.
type Node interface {
	TokenLiteral() string // Devuelve el literal del token asociado al nodo.
	String() string       // Devuelve una representación en string del nodo para debugging.
}

// Statement es una interfaz para todos los nodos de sentencia.
type Statement interface {
	Node
	statementNode() // Método marcador para identificar nodos de sentencia.
}

// Expression es una interfaz para todos los nodos de expresión.
type Expression interface {
	Node
	expressionNode() // Método marcador para identificar nodos de expresión.
}

// Pattern es una interfaz para todos los nodos de patrón en pattern matching.
type Pattern interface {
	Node
	patternNode() // Método marcador para identificar nodos de patrón.
}

// Program es el nodo raíz de todo AST de un programa Zylo.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// ImportStatement representa una declaración de import (e.g., import zyloruntime).
type ImportStatement struct {
	Token           lexer.Token // El token 'import'.
	ModuleName      *Identifier // El nombre del módulo a importar (e.g., 'math' en 'import math').
	ModulePath      string      // La ruta del módulo si se importa con un string (e.g., "std/json").
	ImportedSymbols []*Identifier // Símbolos específicos importados (e.g., '{ sqrt, pow }' en 'import { sqrt, pow } from math').
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) expressionNode()      {} // También implementa la interfaz Expression
func (is *ImportStatement) TokenLiteral() string { return is.Token.Lexeme }
func (is *ImportStatement) String() string {
	var out string
	out += "import "
	if len(is.ImportedSymbols) > 0 {
		out += "{ "
		for i, sym := range is.ImportedSymbols {
			out += sym.String()
			if i < len(is.ImportedSymbols)-1 {
				out += ", "
			}
		}
		out += " }"
		if is.ModuleName != nil {
			out += " from " + is.ModuleName.String()
		} else if is.ModulePath != "" {
			out += fmt.Sprintf(" from %q", is.ModulePath)
		}
	} else if is.ModuleName != nil {
		out += is.ModuleName.String()
	} else if is.ModulePath != "" {
		out += fmt.Sprintf("%q", is.ModulePath)
	}
	out += ";"
	return out
}

// ExportStatement representa una declaración de exportación (e.g., export func myFunc()).
type ExportStatement struct {
	Token       lexer.Token // El token 'export'.
	Declaration Statement   // La declaración que se exporta (FuncStatement, ClassStatement, VarStatement).
}

func (es *ExportStatement) statementNode()       {}
func (es *ExportStatement) TokenLiteral() string { return es.Token.Lexeme }
func (es *ExportStatement) String() string {
	out := "export "
	if es.Declaration != nil {
		out += es.Declaration.String()
	}
	return out
}

// VarStatement representa una declaración de variable (e.g., x := 5;).
type VarStatement struct {
	Token               lexer.Token // El token del modificador o ':='.
	Name                *Identifier
	Value               Expression
	IsConstant          bool         // Indica si es una constante (nombre en mayúsculas)
	IsDestructuring     bool         // Indica si es una asignación por desestructuración
	DestructuringElements []Expression // Elementos para desestructuración (identificadores o patrones anidados)
	Visibility          string       // "public", "private", o vacío para package-private
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Lexeme }
func (vs *VarStatement) String() string {
	var out string
	if vs.Visibility != "" {
		out += vs.Visibility + " "
	}
	if vs.IsDestructuring {
		// TODO: Mejorar la representación de string para desestructuración
		out += fmt.Sprintf("destructuring(%s)", formatExpressions(vs.DestructuringElements))
	} else if vs.Name != nil {
		out += vs.Name.String()
	}
	if vs.Value != nil {
		out += " = " + vs.Value.String()
	}
	out += ";"
	return out
}

// Identifier representa un identificador en el código.
type Identifier struct {
	Token          lexer.Token // El token IDENTIFIER.
	Value          string
	TypeAnnotation string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Lexeme }
func (i *Identifier) String() string       { return i.Value }

// ExpressionStatement es una sentencia que consiste en una sola expresión.
type ExpressionStatement struct {
	Token      lexer.Token // El primer token de la expresión.
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Lexeme }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// FuncStatement representa una declaración de función.
type FuncStatement struct {
	Token       lexer.Token // El token del modificador o identificador.
	Name        *Identifier
	Parameters  []*Identifier
	ReturnType  string // Nuevo campo para el tipo de retorno
	Body        *BlockStatement
	IsAsync     bool   // Nuevo campo para indicar si la función es asíncrona
	Visibility  string // "public", "private", o vacío para package-private
	IsVoid      bool   // Nuevo campo para indicar si es una función void
}
func (fs *FuncStatement) statementNode()       {}
func (fs *FuncStatement) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *FuncStatement) String() string {
	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	returnType := ""
	if fs.ReturnType != "" {
		returnType = fmt.Sprintf(": %s", fs.ReturnType)
	}
	asyncPrefix := ""
	if fs.IsAsync {
		asyncPrefix = "async "
	}
	visibilityPrefix := ""
	if fs.Visibility != "" {
		visibilityPrefix = fs.Visibility + " "
	}
	voidPrefix := ""
	if fs.IsVoid {
		voidPrefix = "void "
	}
	return fmt.Sprintf("%s%s%s%s(%s)%s %s", asyncPrefix, visibilityPrefix, voidPrefix, fs.Name.String(), formatStrings(params), returnType, fs.Body.String())
}

// FunctionLiteral representa una función anónima como expresión (e.g., func() {}).
type FunctionLiteral struct {
	Token      lexer.Token // El token 'func'.
	Parameters []*Identifier
	ReturnType string
	Body       *BlockStatement
	IsAsync    bool
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Lexeme }
func (fl *FunctionLiteral) String() string {
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	returnType := ""
	if fl.ReturnType != "" {
		returnType = fmt.Sprintf(": %s", fl.ReturnType)
	}
	asyncPrefix := ""
	if fl.IsAsync {
		asyncPrefix = "async "
	}
	return fmt.Sprintf("%sfunc(%s)%s %s", asyncPrefix, formatStrings(params), returnType, fl.Body.String())
}

// ArrowFunctionExpression representa una función flecha (e.g., (x) => x * 2).
type ArrowFunctionExpression struct {
	Token      lexer.Token // El token '=>'.
	Parameters []*Identifier
	ReturnType string      // Nuevo campo para el tipo de retorno
	Body       *BlockStatement // Cuerpo de la función si es un bloque
	Expression Expression      // Expresión si es una expresión de una sola línea
	IsAsync    bool        // Nuevo campo para indicar si la función es asíncrona
}

func (afe *ArrowFunctionExpression) expressionNode()      {}
func (afe *ArrowFunctionExpression) TokenLiteral() string { return afe.Token.Lexeme }
func (afe *ArrowFunctionExpression) String() string {
	params := []string{}
	for _, p := range afe.Parameters {
		params = append(params, p.String())
	}
	returnType := ""
	if afe.ReturnType != "" {
		returnType = fmt.Sprintf(" -> %s", afe.ReturnType)
	}
	asyncPrefix := ""
	if afe.IsAsync {
		asyncPrefix = "async "
	}
	if afe.Body != nil {
		return fmt.Sprintf("%s(%s)%s => %s", asyncPrefix, formatStrings(params), returnType, afe.Body.String())
	}
	return fmt.Sprintf("%s(%s)%s => %s", asyncPrefix, formatStrings(params), returnType, afe.Expression.String())
}

// AwaitExpression representa una expresión 'await'.
type AwaitExpression struct {
	Token    lexer.Token // El token 'await'.
	Argument Expression
}

func (ae *AwaitExpression) expressionNode()      {}
func (ae *AwaitExpression) TokenLiteral() string { return ae.Token.Lexeme }
func (ae *AwaitExpression) String() string {
	if ae.Argument == nil {
		return "await INVALID"
	}
	return fmt.Sprintf("await %s", ae.Argument.String())
}

// ReturnStatement representa una sentencia de retorno.
type ReturnStatement struct {
	Token       lexer.Token // El token 'return'.
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Lexeme }
func (rs *ReturnStatement) String() string {
	var out string
	out += rs.TokenLiteral() + " "
	if rs.ReturnValue != nil {
		out += rs.ReturnValue.String()
	}
	out += ";"
	return out
}

// BlockStatement representa un bloque de código entre llaves.
type BlockStatement struct {
	Token      lexer.Token // El token '{'.
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Lexeme }
func (bs *BlockStatement) String() string {
	var out string
	out += "{\n"
	for _, s := range bs.Statements {
		out += "    " + s.String() + "\n"
	}
	out += "}"
	return out
}

// ForInStatement representa una sentencia 'for' con iteración sobre rangos o listas.
type ForInStatement struct {
	Token      lexer.Token // El token 'for'.
	Identifier *Identifier // El identificador de la variable de iteración (e.g., 'x' in 'for x in ...').
	Iterable   Expression  // La expresión que evalúa a la lista o rango sobre el que iterar.
	Body       *BlockStatement // El cuerpo del bucle.
}

func (fs *ForInStatement) statementNode()       {}
func (fs *ForInStatement) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *ForInStatement) String() string {
	out := "for "
	if fs.Identifier != nil {
		out += fs.Identifier.String()
	}
	out += " in "
	if fs.Iterable != nil {
		out += fs.Iterable.String()
	}
	out += " "
	if fs.Body != nil {
		out += fs.Body.String()
	}
	return out
}

// ForStatement representa una sentencia 'for' tradicional.
type ForStatement struct {
	Token     lexer.Token
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *ForStatement) String() string {
	out := "for "
	if fs.Init != nil {
		out += fs.Init.String()
	}
	out += "; "
	if fs.Condition != nil {
		out += fs.Condition.String()
	}
	out += "; "
	if fs.Post != nil {
		out += fs.Post.String()
	}
	out += " "
	if fs.Body != nil {
		out += fs.Body.String()
	}
	return out
}

// TryStatement representa una sentencia 'try-catch'.
type TryStatement struct {
	Token        lexer.Token // El token 'try'.
	TryBlock     *BlockStatement
	CatchClause  *CatchClause // Puede ser nil si solo hay finally.
	FinallyBlock *BlockStatement // Puede ser nil.
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Lexeme }
func (ts *TryStatement) String() string {
	out := "try "
	if ts.TryBlock != nil {
		out += ts.TryBlock.String()
	}
	if ts.CatchClause != nil {
		out += " " + ts.CatchClause.String()
	}
	if ts.FinallyBlock != nil {
		out += " finally " + ts.FinallyBlock.String()
	}
	return out
}

// CatchClause representa una cláusula 'catch'.
type CatchClause struct {
	Token      lexer.Token // El token 'catch'.
	Parameter  *Identifier // El identificador para la excepción capturada.
	CatchBlock *BlockStatement
}

func (cc *CatchClause) statementNode()       {} // CatchClause es parte de TryStatement, no una sentencia independiente.
func (cc *CatchClause) TokenLiteral() string { return cc.Token.Lexeme }
func (cc *CatchClause) String() string {
	if cc.Parameter == nil || cc.CatchBlock == nil {
		return "catch (invalid) { invalid }"
	}
	return fmt.Sprintf("catch (%s) %s", cc.Parameter.String(), cc.CatchBlock.String())
}

// ThrowStatement representa una sentencia 'throw'.
type ThrowStatement struct {
	Token     lexer.Token // El token 'throw'.
	Exception Expression
}

func (ths *ThrowStatement) statementNode()       {}
func (ths *ThrowStatement) TokenLiteral() string { return ths.Token.Lexeme }
func (ths *ThrowStatement) String() string {
	var out string
	out += ths.TokenLiteral() + " "
	if ths.Exception != nil {
		out += ths.Exception.String()
	}
	out += ";"
	return out
}

// NumberLiteral representa un literal numérico.
type NumberLiteral struct {
	Token lexer.Token
	Value interface{} // int64 or float64
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Lexeme }
func (nl *NumberLiteral) String() string       { return nl.Token.Lexeme }

// StringLiteral representa un literal de cadena.
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Lexeme }
func (sl *StringLiteral) String() string       { return sl.Token.Lexeme }

// TemplateStringLiteral representa un literal de cadena de plantilla (template string).
type TemplateStringLiteral struct {
	Token lexer.Token        // El token '`'.
	Value string             // El contenido de la plantilla (sin interpolación aún).
	Parts []interface{}      // Partes: strings y expresiones interpoladas.
	// Parts alterna: string, Expression, string, Expression, ...
}

func (tsl *TemplateStringLiteral) expressionNode()      {}
func (tsl *TemplateStringLiteral) TokenLiteral() string { return tsl.Token.Lexeme }
func (tsl *TemplateStringLiteral) String() string       { return fmt.Sprintf("`%s`", tsl.Value) }

// BooleanLiteral representa un literal booleano.
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Lexeme }
func (bl *BooleanLiteral) String() string       { return bl.Token.Lexeme }

// NullLiteral representa un literal null.
type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Lexeme }
func (nl *NullLiteral) String() string       { return "null" }

// PrefixExpression representa una expresión con un operador prefijo.
type PrefixExpression struct {
	Token    lexer.Token // El operador prefijo.
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Lexeme }
func (pe *PrefixExpression) String() string {
	if pe.Right == nil {
		return fmt.Sprintf("(%sINVALID)", pe.Operator)
	}
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// InfixExpression representa una expresión con un operador infijo.
type InfixExpression struct {
	Token    lexer.Token // El operador infijo.
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Lexeme }
func (ie *InfixExpression) String() string {
	if ie.Left == nil || ie.Right == nil {
		return fmt.Sprintf("(INVALID %s INVALID)", ie.Operator)
	}
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// CallExpression representa una llamada a función.
type CallExpression struct {
	Token     lexer.Token // El token '(' o el identificador de la función.
	Function  Expression  // La expresión que evalúa a la función.
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Lexeme }
func (ce *CallExpression) String() string {
	if ce.Function == nil {
		return "INVALID()"
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), formatExpressions(ce.Arguments))
}

// MethodCallExpression representa una llamada a método (e.g., obj.method(args)).
type MethodCallExpression struct {
	Token     lexer.Token   // El token '(' o el identificador del método.
	Object    Expression    // El objeto sobre el que se llama el método.
	Property  *Identifier   // El identificador del método.
	Arguments []Expression  // Los argumentos pasados al método.
}

func (mce *MethodCallExpression) expressionNode()      {}
func (mce *MethodCallExpression) TokenLiteral() string { return mce.Token.Lexeme }
func (mce *MethodCallExpression) String() string {
	if mce.Object == nil || mce.Property == nil {
		return "INVALID.METHOD()"
	}
	return fmt.Sprintf("%s.%s(%s)", mce.Object.String(), mce.Property.String(), formatExpressions(mce.Arguments))
}

// RangeExpression representa una expresión de rango (e.g., 1..10).
type RangeExpression struct {
	Token lexer.Token // El token '..'.
	Start Expression  // Expresión de inicio.
	End   Expression  // Expresión de fin.
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Lexeme }
func (re *RangeExpression) String() string {
	if re.Start == nil || re.End == nil {
		return "INVALID..INVALID"
	}
	return fmt.Sprintf("%s..%s", re.Start.String(), re.End.String())
}

// SliceExpression representa una expresión de slice (e.g., list[1:5]).
type SliceExpression struct {
	Token lexer.Token // El token ':'.
	Left  Expression  // El objeto a slicer (e.g., lista).
	Start Expression  // Índice de inicio (opcional).
	End   Expression  // Índice de fin (opcional).
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token.Lexeme }
func (se *SliceExpression) String() string {
	start := ""
	if se.Start != nil {
		start = se.Start.String()
	}
	end := ""
	if se.End != nil {
		end = se.End.String()
	}
	return fmt.Sprintf("%s[%s:%s]", se.Left.String(), start, end)
}

// IndexExpression representa el acceso a un índice (ej. array[index], array[start:end], array[-1]).
type IndexExpression struct {
	Token        lexer.Token // El token '['
	Left         Expression  // La expresión que evalúa al objeto indexable.
	Index        Expression  // La expresión que evalúa al índice.
	EndIndex     Expression  // Para slicing: array[start:end] (nil si no es slice)
	NegativeIndex bool       // Para negative indexing: array[-1]
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Lexeme }
func (ie *IndexExpression) String() string {
	if ie.Left == nil || ie.Index == nil {
		return "(INVALID[INVALID])"
	}
	return fmt.Sprintf("(%s[%s])", ie.Left.String(), ie.Index.String())
}

// MemberExpression representa el acceso a un miembro (ej. object.property).
type MemberExpression struct {
	Token    lexer.Token // El token del identificador de la propiedad.
	Object   Expression  // La expresión que evalúa al objeto.
	Property *Identifier // El identificador de la propiedad.
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Lexeme }
func (me *MemberExpression) String() string {
	if me.Object == nil || me.Property == nil {
		return "(INVALID.INVALID)"
	}
	return fmt.Sprintf("(%s.%s)", me.Object.String(), me.Property.String())
}

// BlockExpression representa un bloque de código como una expresión.
type BlockExpression struct {
	Token lexer.Token // El token '{'.
	Block *BlockStatement
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Lexeme }
func (be *BlockExpression) String() string {
	if be.Block == nil {
		return "{INVALID}"
	}
	return be.Block.String()
}

// IfStatement representa una sentencia 'if'.
type IfStatement struct {
	Token       lexer.Token     // El token 'if'.
	Condition   Expression      // La condición del if.
	Consequence *BlockStatement // El bloque del if.
	Alternative *BlockStatement // El bloque del else/elif (si existe).
	// Note: Elif se maneja como IfStatement dentro del Alternative.
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Lexeme }
func (is *IfStatement) String() string {
	out := "if "
	if is.Condition != nil {
		out += is.Condition.String()
	}
	out += " "
	if is.Consequence != nil {
		out += is.Consequence.String()
	}
	if is.Alternative != nil {
		out += " else " + is.Alternative.String()
	}
	return out
}

// IfExpression representa una expresión 'if' (e.g., if condition { true_exp } else { false_exp }).
type IfExpression struct {
	Token       lexer.Token     // El token 'if'.
	Condition   Expression      // La condición del if.
	Consequence *BlockStatement // El bloque del if.
	Alternative *BlockStatement // El bloque del else (si existe).
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Lexeme }
func (ie *IfExpression) String() string {
	out := "if "
	if ie.Condition != nil {
		out += ie.Condition.String()
	}
	out += " "
	if ie.Consequence != nil {
		out += ie.Consequence.String()
	}
	if ie.Alternative != nil {
		out += " else " + ie.Alternative.String()
	}
	return out
}

// BreakStatement representa una sentencia 'break'.
type BreakStatement struct {
	Token lexer.Token // El token 'break'.
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Lexeme }
func (bs *BreakStatement) String() string       { return bs.Token.Lexeme + ";" }

// ContinueStatement representa una sentencia 'continue'.
type ContinueStatement struct {
	Token lexer.Token // El token 'continue'.
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Lexeme }
func (cs *ContinueStatement) String() string       { return cs.Token.Lexeme + ";" }

// WhileStatement representa una sentencia 'while'.
type WhileStatement struct {
	Token     lexer.Token // El token 'while'.
	Condition Expression  // La condición del bucle.
	Body      *BlockStatement // El cuerpo del bucle.
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Lexeme }
func (ws *WhileStatement) String() string {
	out := "while "
	if ws.Condition != nil {
		out += ws.Condition.String()
	}
	out += " "
	if ws.Body != nil {
		out += ws.Body.String()
	}
	return out
}

// MethodStatement representa una declaración de método en una clase.
type MethodStatement struct {
	Token      lexer.Token // El token 'func'.
	Name       *Identifier
	Parameters []*Identifier
	ReturnType string // Tipo de retorno
	Body       *BlockStatement
	IsAsync    bool // Nuevo campo para indicar si el método es asíncrono
}

func (ms *MethodStatement) statementNode()       {}
func (ms *MethodStatement) TokenLiteral() string { return ms.Token.Lexeme }
func (ms *MethodStatement) String() string {
	params := []string{}
	for _, p := range ms.Parameters {
		params = append(params, p.String())
	}
	returnType := ""
	if ms.ReturnType != "" {
		returnType = fmt.Sprintf(": %s", ms.ReturnType)
	}
	asyncPrefix := ""
	if ms.IsAsync {
		asyncPrefix = "async "
	}
	return fmt.Sprintf("%sfunc %s(%s)%s %s", asyncPrefix, ms.Name.String(), formatStrings(params), returnType, ms.Body.String())
}

// ConstructorStatement representa una declaración de constructor (init) en una clase.
type ConstructorStatement struct {
	Token      lexer.Token // El token 'func'.
	Name       *Identifier // Debería ser 'init'
	Parameters []*Identifier
	Body       *BlockStatement
}

func (cs *ConstructorStatement) statementNode()       {}
func (cs *ConstructorStatement) TokenLiteral() string { return cs.Token.Lexeme }
func (cs *ConstructorStatement) String() string {
	params := []string{}
	for _, p := range cs.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("func %s(%s) %s", cs.Name.String(), formatStrings(params), cs.Body.String())
}

// ClassStatement representa una declaración de clase.
type ClassStatement struct {
	Token       lexer.Token // El token del modificador o 'class'.
	Name        *Identifier
	SuperClass  *Identifier               // Nuevo campo para la superclase
	TypeParams  []string                  // Generic type parameters
	Attributes  []*VarStatement           // Atributos de la clase
	Methods     []*MethodStatement        // Métodos de la clase
	InitMethod  *ConstructorStatement     // Método constructor (init)
	Visibility  string                    // "public", "private", o vacío para package-private
	IsVoid      bool                      // Nuevo campo para indicar si es una clase void
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Lexeme }
func (cs *ClassStatement) String() string {
	out := ""
	if cs.Visibility != "" {
		out += cs.Visibility + " "
	}
	if cs.IsVoid {
		out += "void "
	}
	out += "class "
	if cs.Name != nil {
		out += cs.Name.String()
	}
	if len(cs.TypeParams) > 0 {
		out += "<" + strings.Join(cs.TypeParams, ", ") + ">"
	}
	if cs.SuperClass != nil {
		out += " extends " + cs.SuperClass.String()
	}
	out += " {\n"
	for _, attr := range cs.Attributes {
		out += "    " + attr.String() + "\n"
	}
	for _, method := range cs.Methods {
		out += "    " + method.String() + "\n"
	}
	out += "}"
	return out
}

// ListLiteral representa un literal de lista (e.g., [1, 2, 3]).
type ListLiteral struct {
	Token    lexer.Token // El token '['.
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Lexeme }
func (ll *ListLiteral) String() string {
	if ll.Elements == nil {
		return "[]"
	}
	return fmt.Sprintf("[%s]", formatExpressions(ll.Elements))
}

// SetLiteral representa un literal de conjunto (e.g., {1, 2, 3}).
type SetLiteral struct {
	Token    lexer.Token // El token '{'.
	Elements []Expression
}

func (sl *SetLiteral) expressionNode()      {}
func (sl *SetLiteral) TokenLiteral() string { return sl.Token.Lexeme }
func (sl *SetLiteral) String() string {
	if sl.Elements == nil {
		return "{}"
	}
	return fmt.Sprintf("{%s}", formatExpressions(sl.Elements))
}

// MapLiteral representa un literal de mapa (e.g., {key: value}).
type MapLiteral struct {
	Token lexer.Token // El token '{'.
	Pairs map[string]Expression
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Lexeme }
func (ml *MapLiteral) String() string {
	if ml.Pairs == nil {
		return "{}"
	}
	var pairs []string
	for k, v := range ml.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v.String()))
	}
	return fmt.Sprintf("{%s}", formatStrings(pairs))
}

// ClassInstantiation representa la instanciación de una clase (e.g., Persona("Wilson", 25)).
type ClassInstantiation struct {
	Token     lexer.Token // El token de la clase.
	ClassName *Identifier
	Arguments []Expression
}

func (ci *ClassInstantiation) expressionNode()      {}
func (ci *ClassInstantiation) TokenLiteral() string { return ci.Token.Lexeme }
func (ci *ClassInstantiation) String() string {
	if ci.ClassName == nil {
		return "INVALID()"
	}
	return fmt.Sprintf("%s(%s)", ci.ClassName.String(), formatExpressions(ci.Arguments))
}

// ObjectLiteral representa un literal de objeto para clases (e.g., Result{value: 5}).
type ObjectLiteral struct {
	Token    lexer.Token              // El token '{'.
	ClassName *Identifier             // Nombre de la clase (opcional, para Result).
	Fields   map[*Identifier]Expression // Campos y sus valores.
}

func (ol *ObjectLiteral) expressionNode()      {}
func (ol *ObjectLiteral) TokenLiteral() string { return ol.Token.Lexeme }
func (ol *ObjectLiteral) String() string {
	if ol.Fields == nil {
		return "{}"
	}
	var pairs []string
	for key, value := range ol.Fields {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key.String(), value.String()))
	}
	className := ""
	if ol.ClassName != nil {
		className = ol.ClassName.String()
	}
	return fmt.Sprintf("%s{%s}", className, formatStrings(pairs))
}

// ThisExpression representa la expresión 'this'
type ThisExpression struct {
	Token lexer.Token // El token 'this'.
}

func (te *ThisExpression) expressionNode()      {}
func (te *ThisExpression) TokenLiteral() string { return te.Token.Lexeme }
func (te *ThisExpression) String() string       { return "this" }

// SuperExpression representa la expresión 'super'
type SuperExpression struct {
	Token lexer.Token // El token 'super'.
}

func (se *SuperExpression) expressionNode()      {}
func (se *SuperExpression) TokenLiteral() string { return se.Token.Lexeme }
func (se *SuperExpression) String() string       { return "super" }

// Helper para formatear listas de expresiones en strings.
func formatExpressions(exps []Expression) string {
	var parts []string
	for _, exp := range exps {
		parts = append(parts, exp.String())
	}
	return formatStrings(parts)
}

func formatStrings(strs []string) string {
	var result string
	for i, s := range strs {
		result += s
		if i < len(strs)-1 {
			result += ", "
		}
	}
	return result
}

// AssignmentExpression representa una asignación (e.g., x = 5, x += 1).
type AssignmentExpression struct {
	Token    lexer.Token
	Name     Expression // Cambiado de *Identifier a Expression
	Operator string
	Value    Expression
}

func (a *AssignmentExpression) expressionNode()      {}
func (a *AssignmentExpression) TokenLiteral() string { return a.Token.Lexeme }
func (a *AssignmentExpression) String() string {
	return fmt.Sprintf("%s %s %s", a.Name.String(), a.Operator, a.Value.String())
}

// DestructuringAssignmentExpression represents a destructuring assignment expression (e.g., [a, b] = [1, 2]).
type DestructuringAssignmentExpression struct {
	Token    lexer.Token  // El token '='.
	Targets  []Expression // Los identificadores o patrones a la izquierda.
	Value    Expression   // La expresión a la derecha.
	Operator string       // Operador, e.g., "="
}

func (da *DestructuringAssignmentExpression) expressionNode()      {}
func (da *DestructuringAssignmentExpression) TokenLiteral() string { return da.Token.Lexeme }
func (da *DestructuringAssignmentExpression) String() string {
	return fmt.Sprintf("%s %s %s", formatExpressions(da.Targets), da.Operator, da.Value.String())
}

// DotExpression representa el acceso a propiedad con punto (e.g., obj.prop).
type DotExpression struct {
	Token    lexer.Token
	Left     Expression
	Property *Identifier
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Lexeme }
func (de *DotExpression) String() string {
	return fmt.Sprintf("%s.%s", de.Left.String(), de.Property.String())
}

// SwitchStatement representa una sentencia 'switch-case'.
type SwitchStatement struct {
	Token      lexer.Token // El token 'switch'.
	Expression Expression
	Cases      []*CaseClause
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Lexeme }
func (ss *SwitchStatement) String() string {
	out := fmt.Sprintf("switch %s {\n", ss.Expression.String())
	for _, c := range ss.Cases {
		out += c.String() + "\n"
	}
	out += "}"
	return out
}

// CaseClause representa una cláusula 'case' o 'default' dentro de un switch.
type CaseClause struct {
	Token      lexer.Token // El token 'case' o 'default'.
	Expression Expression  // La expresión a comparar (nil para default).
	Body       *BlockStatement
}

func (cc *CaseClause) statementNode()       {} // No es una sentencia independiente
func (cc *CaseClause) TokenLiteral() string { return cc.Token.Lexeme }
func (cc *CaseClause) String() string {
	out := "case "
	if cc.Expression != nil {
		out += cc.Expression.String()
	}
	out += ": "
	if cc.Body != nil {
		out += cc.Body.String()
	}
	return out
}

// TypePattern representa un patrón de tipo, e.g., String(s), Int(n)
type TypePattern struct {
	Token    lexer.Token // El token del nombre del tipo
	TypeName string      // Nombre del tipo (e.g., "String")
	Variable *Identifier // Variable a la que asignar el valor (opcional)
}

func (tp *TypePattern) patternNode()       {}
func (tp *TypePattern) TokenLiteral() string { return tp.Token.Lexeme }
func (tp *TypePattern) String() string {
	if tp.Variable != nil {
		return fmt.Sprintf("%s(%s)", tp.TypeName, tp.Variable.String())
	}
	return tp.TypeName
}

// VariablePattern representa un patrón de variable, e.g., x
type VariablePattern struct {
	Token lexer.Token
	Name  *Identifier
}

func (vp *VariablePattern) patternNode()       {}
func (vp *VariablePattern) TokenLiteral() string { return vp.Token.Lexeme }
func (vp *VariablePattern) String() string {
	if vp.Name != nil {
		return vp.Name.String()
	}
	return "_"
}

// LiteralPattern representa un patrón literal, e.g., 5, "hello"
type LiteralPattern struct {
	Token lexer.Token
	Value Expression
}

func (lp *LiteralPattern) patternNode()       {}
func (lp *LiteralPattern) TokenLiteral() string { return lp.Token.Lexeme }
func (lp *LiteralPattern) String() string {
	if lp.Value != nil {
		return lp.Value.String()
	}
	return ""
}

// MatchStatement representa una sentencia 'match' para pattern matching.
type MatchStatement struct {
	Token      lexer.Token // El token 'match'.
	Expression Expression
	Cases      []*PatternCase
}

func (ms *MatchStatement) statementNode()       {}
func (ms *MatchStatement) TokenLiteral() string { return ms.Token.Lexeme }
func (ms *MatchStatement) String() string {
	out := fmt.Sprintf("match %s {\n", ms.Expression.String())
	for _, c := range ms.Cases {
		out += c.String() + "\n"
	}
	out += "}"
	return out
}

// PatternCase representa una cláusula 'case' dentro de un match.
type PatternCase struct {
	Token   lexer.Token // El token 'case' o 'default'.
	Pattern Pattern     // El patrón a comparar.
	Guard   Expression  // Condición opcional (if).
	Body    *BlockStatement
}

func (pc *PatternCase) statementNode()       {} // No es una sentencia independiente
func (pc *PatternCase) TokenLiteral() string { return pc.Token.Lexeme }
func (pc *PatternCase) String() string {
	out := "case "
	if pc.Pattern != nil {
		out += pc.Pattern.String()
	}
	if pc.Guard != nil {
		out += " if " + pc.Guard.String()
	}
	out += ": "
	if pc.Body != nil {
		out += pc.Body.String()
	}
	return out
}

// SpawnStatement representa una sentencia 'spawn' para ejecutar código concurrentemente.
type SpawnStatement struct {
	Token  lexer.Token
	Body   *BlockStatement
}

func (ss *SpawnStatement) statementNode()       {}
func (ss *SpawnStatement) TokenLiteral() string { return ss.Token.Lexeme }
func (ss *SpawnStatement) String() string {
	out := "spawn "
	if ss.Body != nil {
		out += ss.Body.String()
	}
	return out
}

// CollectionMethodCall representa una llamada a método en una colección (e.g., arr.push(element)).
type CollectionMethodCall struct {
	Token     lexer.Token   // El token '('.
	Object    Expression    // El objeto colección.
	Method    *Identifier   // El nombre del método.
	Arguments []Expression  // Los argumentos del método.
}

func (cmc *CollectionMethodCall) expressionNode()      {}
func (cmc *CollectionMethodCall) TokenLiteral() string { return cmc.Token.Lexeme }
func (cmc *CollectionMethodCall) String() string {
	if cmc.Object == nil || cmc.Method == nil {
		return "INVALID.METHOD()"
	}
	return fmt.Sprintf("%s.%s(%s)", cmc.Object.String(), cmc.Method.String(), formatExpressions(cmc.Arguments))
}

// AsExpression representa una expresión de conversión de tipo (e.g., value as Type).
type AsExpression struct {
	Token    lexer.Token // El token 'as'.
	Left     Expression  // La expresión a convertir.
	TypeName string      // El nombre del tipo al que convertir.
}

func (ae *AsExpression) expressionNode()      {}
func (ae *AsExpression) TokenLiteral() string { return ae.Token.Lexeme }
func (ae *AsExpression) String() string {
	if ae.Left == nil {
		return fmt.Sprintf("INVALID as %s", ae.TypeName)
	}
	return fmt.Sprintf("%s as %s", ae.Left.String(), ae.TypeName)
}
