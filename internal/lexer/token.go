	package lexer

	import "fmt"

	// TokenType es un string que representa el tipo de un token.
	type TokenType string

	// Token representa una unidad léxica del lenguaje Zylo.
	type Token struct {
		Type      TokenType   // El tipo del token (e.g., IDENTIFIER, NUMBER).
		Lexeme    string      // El substring del código fuente que representa el token.
		Literal   interface{} // El valor literal del token, si aplica (e.g., 123, "hello").
		StartLine int         // La línea donde comienza el token.
		StartCol  int         // La columna donde comienza el token.
		EndLine   int         // La línea donde termina el token.
		EndCol    int         // La columna donde termina el token.
	}

	// String devuelve una representación legible del token, útil para debugging.
	func (t Token) String() string {
		return fmt.Sprintf("Token(Type: %s, Lexeme: '%s', Literal: %v, Pos: %d:%d-%d:%d)",
			t.Type, t.Lexeme, t.Literal, t.StartLine, t.StartCol, t.EndLine, t.EndCol)
	}

	// Constantes para los tipos de token.
	const (
		// Tokens de un solo carácter
		LEFT_PAREN    TokenType = "LEFT_PAREN"
		RIGHT_PAREN   TokenType = "RIGHT_PAREN"
		LEFT_BRACE    TokenType = "LEFT_BRACE"
		RIGHT_BRACE   TokenType = "RIGHT_BRACE"
		LEFT_BRACKET  TokenType = "LEFT_BRACKET"
		RIGHT_BRACKET TokenType = "RIGHT_BRACKET"
		COMMA         TokenType = "COMMA"
		DOT           TokenType = "DOT"
		RANGE         TokenType = "RANGE" // .. for ranges
		MINUS         TokenType = "MINUS"
		PLUS          TokenType = "PLUS"
		SEMICOLON     TokenType = "SEMICOLON"
		SLASH         TokenType = "SLASH"
		STAR          TokenType = "STAR"
		PERCENT       TokenType = "PERCENT"
		COLON         TokenType = "COLON"

		// Tokens de uno o dos caracteres
		BANG          TokenType = "BANG"
		BANG_EQUAL    TokenType = "BANG_EQUAL"
		EQUAL         TokenType = "EQUAL"
		EQUAL_EQUAL   TokenType = "EQUAL_EQUAL"
		GREATER       TokenType = "GREATER"
		GREATER_EQUAL TokenType = "GREATER_EQUAL"
		LESS          TokenType = "LESS"
		LESS_EQUAL    TokenType = "LESS_EQUAL"
		ARROW_FUNC    TokenType = "ARROW_FUNC"   // =>
		ARROW_RETURN  TokenType = "ARROW_RETURN" // ->

		// Literales
		IDENTIFIER      TokenType = "IDENTIFIER"
		STRING          TokenType = "STRING"
		NUMBER          TokenType = "NUMBER"
		TEMPLATE_STRING TokenType = "TEMPLATE_STRING" // Added for template strings

		// Marcadores de interpolación para template strings
		LEFT_DOLLAR_BRACE  TokenType = "LEFT_DOLLAR_BRACE"  // ${
		RIGHT_DOLLAR_BRACE TokenType = "RIGHT_DOLLAR_BRACE" // }

		// Tipos primitivos
		STRING_TYPE TokenType = "STRING_TYPE"
		FLOAT_TYPE  TokenType = "FLOAT_TYPE"
		INT_TYPE    TokenType = "INT_TYPE"
		BOOL_TYPE   TokenType = "BOOL_TYPE"
		LIST_TYPE   TokenType = "LIST_TYPE"
		MAP_TYPE    TokenType = "MAP_TYPE"
		ANY_TYPE    TokenType = "ANY_TYPE"

		// Palabras clave
		AND      TokenType = "AND"
		CLASS    TokenType = "CLASS"
		ELSE     TokenType = "ELSE"
		ELIF     TokenType = "ELIF"
		FALSE    TokenType = "FALSE"
		FOR      TokenType = "FOR"
		FUNC     TokenType = "FUNC"
		IF       TokenType = "IF"
		NIL      TokenType = "NIL"
		OR       TokenType = "OR"
		NOT      TokenType = "NOT"
		RETURN   TokenType = "RETURN"
		SUPER    TokenType = "SUPER"
		THIS     TokenType = "THIS"
		TRUE     TokenType = "TRUE"
		VAR      TokenType = "VAR"
		CONST    TokenType = "CONST"
		WHILE    TokenType = "WHILE"
		BREAK    TokenType = "BREAK"
		CONTINUE TokenType = "CONTINUE"
		SHOW     TokenType = "SHOW"
		LOG      TokenType = "LOG"
		IMPORT   TokenType = "IMPORT"
		FROM     TokenType = "FROM"
		TRY      TokenType = "TRY"
		CATCH    TokenType = "CATCH"
		THROW    TokenType = "THROW"
		FINALLY  TokenType = "FINALLY"
		ASYNC    TokenType = "ASYNC"
		AWAIT    TokenType = "AWAIT"
		SPAWN    TokenType = "SPAWN"
		IN       TokenType = "IN"
		SWITCH   TokenType = "SWITCH"  // Nueva palabra clave
		CASE     TokenType = "CASE"    // Nueva palabra clave
		DEFAULT  TokenType = "DEFAULT" // Nueva palabra clave
		EXTENDS  TokenType = "EXTENDS" // Nueva palabra clave para herencia
		EXPORT   TokenType = "EXPORT"  // Nueva palabra clave para módulos
		MATCH    TokenType = "MATCH"   // Nueva palabra clave para pattern matching
		AS       TokenType = "AS"      // Nueva palabra clave para conversión de tipos
		PUBLIC   TokenType = "PUBLIC"  // Nueva palabra clave para visibilidad
		PRIVATE  TokenType = "PRIVATE" // Nueva palabra clave para visibilidad
		VOID     TokenType = "VOID"    // Nueva palabra clave para funciones sin retorno

		// Operadores compuestos
		PLUS_EQUAL    TokenType = "PLUS_EQUAL"    // +=
		MINUS_EQUAL   TokenType = "MINUS_EQUAL"   // -=
		STAR_EQUAL    TokenType = "STAR_EQUAL"    // *=
		SLASH_EQUAL   TokenType = "SLASH_EQUAL"   // /=
		PERCENT_EQUAL TokenType = "PERCENT_EQUAL" // %=
		POWER         TokenType = "POWER"         // **
		FLOOR_DIVIDE  TokenType = "FLOOR_DIVIDE"  // //
		WALRUS_ASSIGN TokenType = "WALRUS_ASSIGN" // :=

		// Control
		NEWLINE TokenType = "NEWLINE"
		EOF     TokenType = "EOF"
		ERROR   TokenType = "ERROR"
	)

	// keywords es un mapa que asocia las palabras clave del lenguaje Zylo con sus tipos de token correspondientes.
		var keywords = map[string]TokenType{
			"and":      AND,
			"class":    CLASS,
			"else":     ELSE,
			"elif":     ELIF,
			"false":    FALSE,
			"for":      FOR,
			"func":     FUNC,
			"if":       IF,
			"nil":      NIL,
			"or":       OR,
			"not":      NOT, // AÑADIDO: soporte para 'not' como palabra clave
			"return":   RETURN,
			"super":    SUPER,
			"this":     THIS,
			"self":     THIS, // CORREGIDO: añadido 'self' como alias de 'this'
			"true":      TRUE,
			"var":       VAR,
			"const":     CONST,
			"while":    WHILE,
			"break":    BREAK,
			"continue": CONTINUE,
			"import":   IMPORT,
			"from":     FROM,
			"try":      TRY,
			"catch":    CATCH,
			"throw":    THROW,
			"finally":  FINALLY,
			"async":    ASYNC,
			"await":    AWAIT,
			"spawn":    SPAWN,
			"in":       IN,
			"switch":   SWITCH,
			"case":     CASE,
			"default":  DEFAULT,
			"extends":  EXTENDS,
			"export":   EXPORT,
			"match":    MATCH,
			"as":       AS,
			"public":   PUBLIC,
			"private":  PRIVATE,
			"void":     VOID,

			// Tipos primitivos Go agregados como palabras clave
			"int":      INT_TYPE,
			"float64":  FLOAT_TYPE,
			"string":   STRING_TYPE,
			"bool":     BOOL_TYPE,
			"list":     LIST_TYPE,
			"map":      MAP_TYPE,

			// Tipos alternativos
			"any": ANY_TYPE,
			"String": STRING_TYPE,
			"Float": FLOAT_TYPE,
			"Bool": BOOL_TYPE,

			// NOTA: "show" and "log" son tratados como identificadores regulares
			// para permitir el acceso a miembros como show.log()
		}
