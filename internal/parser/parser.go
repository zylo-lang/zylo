package parser

import (
	"fmt"
	"strings"
	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/lexer"
)

type Parser struct {
	l              *lexer.Lexer
	curToken       lexer.Token
	peekToken      lexer.Token
	errors         []string
	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	ASSIGN
	ANDOR
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	POWER_PREC
	PREFIX
	CALL
	INDEX
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:  make(map[lexer.TokenType]infixParseFn),
	}

	p.nextToken()
	p.nextToken()

	// Prefix parsers
	p.registerPrefix(lexer.IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(lexer.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TEMPLATE_STRING, p.parseTemplateStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.NIL, p.parseNullLiteral)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LEFT_PAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LEFT_BRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.LEFT_BRACE, p.parseBlockOrCollectionLiteral) // Modificado para manejar bloques, mapas y sets
	p.registerPrefix(lexer.THIS, p.parseThisExpression)

	// Nuevos prefix parsers requeridos por la tarea
	p.registerPrefix(lexer.ASYNC, p.parseAsyncExpression)
	p.registerPrefix(lexer.AWAIT, p.parseAwaitExpression)
	p.registerPrefix(lexer.IF, p.parseIfExpression) // Para expresiones if
	p.registerPrefix(lexer.VAR, p.parseVarExpression) // Stub para evitar errores si 'var' aparece en contexto de expresión
	p.registerPrefix(lexer.RETURN, p.parseReturnExpression) // Stub para evitar errores si 'return' aparece en contexto de expresión
	p.registerPrefix(lexer.NOT, p.parseNotExpression)       // Añadido para la palabra clave 'not'
	p.registerPrefix(lexer.FUNC, p.parseFunctionLiteralPrefix) // Para funciones anónimas como expresiones

	// Stubs temporales para tokens que no deberían ser prefijos pero causan errores
	p.registerPrefix(lexer.COMMA, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.COLON, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.RIGHT_BRACKET, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.RIGHT_PAREN, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.RIGHT_BRACE, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.ERROR, p.parseErrorToken) // Handle lexer error tokens

	// Handle modifiers in expression context (should not happen, but handle gracefully)
	p.registerPrefix(lexer.PUBLIC, p.parseModifierInExpression)
	p.registerPrefix(lexer.PRIVATE, p.parseModifierInExpression)
	p.registerPrefix(lexer.VOID, p.parseModifierInExpression)


	// Handle keywords that shouldn't be prefix but might appear
	p.registerPrefix(lexer.SUPER, p.parseSuperExpression)
	p.registerPrefix(lexer.ELIF, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.ELSE, p.parseUnexpectedPrefix)

	// Handle type tokens in expression context (should not happen)
	p.registerPrefix(lexer.INT_TYPE, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.STRING_TYPE, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.FLOAT_TYPE, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.BOOL_TYPE, p.parseUnexpectedPrefix)
	p.registerPrefix(lexer.WALRUS_ASSIGN, p.parseWalrusAssignInExpression)

	// Infix parsers - operadores de comparación y matemáticos
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.STAR, p.parseInfixExpression)
	p.registerInfix(lexer.PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.POWER, p.parseInfixExpression)
	p.registerInfix(lexer.FLOOR_DIVIDE, p.parseInfixExpression)
	p.registerInfix(lexer.EQUAL_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.BANG_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.LESS, p.parseInfixExpression)
	p.registerInfix(lexer.LESS_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER, p.parseInfixExpression)
	p.registerInfix(lexer.GREATER_EQUAL, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.PLUS_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.MINUS_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.STAR_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.SLASH_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.PERCENT_EQUAL, p.parseAssignmentExpression)
	p.registerInfix(lexer.DOT, p.parseDotExpression)
	p.registerInfix(lexer.LEFT_PAREN, p.parseCallExpression)
	p.registerInfix(lexer.LEFT_BRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.RANGE, p.parseRangeExpression)
	p.registerInfix(lexer.IN, p.parseInExpression)
	p.registerInfix(lexer.ARROW_RETURN, p.parseArrowFunctionExpressionInfix)
	p.registerInfix(lexer.AS, p.parseAsExpression)

	// Comentarios explicativos
	// The prefix parsers for comparison operators are not needed since they work as infix operators
	// This allows expressions like: a <= b, a > c, etc.

	// Register EQUAL as prefix to handle invalid assignments like 5 = 10
	p.registerPrefix(lexer.EQUAL, p.parseUnexpectedPrefix)

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string    { return p.errors }
func (p *Parser) addError(msg string) { p.errors = append(p.errors, msg) }

func (p *Parser) registerPrefix(tt lexer.TokenType, fn prefixParseFn) { p.prefixParseFns[tt] = fn }
func (p *Parser) registerInfix(tt lexer.TokenType, fn infixParseFn)   { p.infixParseFns[tt] = fn }

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Statement{}}

	for p.curToken.Type != lexer.EOF {
		p.skipNewlines()
		if p.curToken.Type == lexer.EOF {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// ✅ Solo avanzar si NO estamos en EOF
		if !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	p.skipNewlines()

	switch p.curToken.Type {
	case lexer.SEMICOLON, lexer.NEWLINE:
		p.nextToken()
		return nil
		case lexer.IDENTIFIER:
		if p.peekTokenIs(lexer.WALRUS_ASSIGN) {
			return p.parseWalrusStatement()
		}
		if p.peekTokenIs(lexer.FOR) {
			// This is a for loop: identifier for condition { ... }
			return p.parseForInLoop()
		}
		if p.isTypeToken(p.peekToken) {
			// This is a typed variable declaration: identifier type := value
			return p.parseTypedVariableDeclaration()
		}
		return p.parseExpressionStatement()
	case lexer.PUBLIC, lexer.PRIVATE, lexer.VOID:
		// Modifier found, parse declaration
		return p.parseDeclaration()
	case lexer.VAR, lexer.CONST:
		return p.parseVarStatement()
	case lexer.FUNC:
		return p.parseFunctionStatement()
	case lexer.ASYNC: // async func statement
		if p.peekTokenIs(lexer.FUNC) {
			p.nextToken() // Consume ASYNC
			return p.parseFunctionStatementWithAsync(true)
		}
		// If not async func, it's an async expression statement
		return p.parseExpressionStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.THROW:
		return p.parseThrowStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.EXPORT:
		return p.parseExportStatement()
	case lexer.SWITCH:
		return p.parseSwitchStatement()
	case lexer.MATCH:
		return p.parseMatchStatement()
	case lexer.SPAWN:
		return p.parseSpawnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// Statements

// parseVarStatement parses a variable declaration (e.g., var x = 10; or public x = 10;).
func (p *Parser) parseVarStatement() ast.Statement {
	token := p.curToken
	var visibility string

	// Consume 'var' keyword if present
	if p.curTokenIs(lexer.VAR) {
		p.nextToken()
	}

	// Check for visibility modifier
	if p.curTokenIs(lexer.PUBLIC) {
		visibility = "public"
		p.nextToken()
	} else if p.curTokenIs(lexer.PRIVATE) {
		visibility = "private"
		p.nextToken()
	}

	stmt := &ast.VarStatement{Token: token, Visibility: visibility}

	// At this point, curToken should be the variable name (IDENTIFIER)
	if !p.curTokenIs(lexer.IDENTIFIER) {
		p.addError(fmt.Sprintf("expected variable name, got %s", p.curToken.Type))
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // Consume COLON
		p.nextToken() // Advance to type identifier
		if p.curTokenIs(lexer.IDENTIFIER) || p.curTokenIs(lexer.ANY_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) {
			stmt.Name.TypeAnnotation = p.curToken.Lexeme
		} else {
			stmt.Name.TypeAnnotation = "ANY"
		}
	}

	if p.peekTokenIs(lexer.EQUAL) {
		p.nextToken() // Consume EQUAL
		p.nextToken() // Advance to expression
		stmt.Value = p.parseExpression(LOWEST)
	} else if p.peekTokenIs(lexer.WALRUS_ASSIGN) {
		p.nextToken() // Consume WALRUS_ASSIGN
		p.nextToken() // Advance to expression
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

// parseVarWithModifier parses a variable declaration where the modifier has already been consumed.
func (p *Parser) parseVarWithModifier(modifier lexer.Token) ast.Statement {
	var visibility string

	if modifier.Type == lexer.PUBLIC {
		visibility = "public"
	} else if modifier.Type == lexer.PRIVATE {
		visibility = "private"
	}

	stmt := &ast.VarStatement{Token: modifier, Visibility: visibility}

	// curToken debe ser el nombre de la variable (IDENTIFIER)
	if !p.curTokenIs(lexer.IDENTIFIER) {
		p.addError(fmt.Sprintf("expected variable name after modifier, got %s", p.curToken.Type))
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	// Verificar tipo de anotación
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // Consume COLON
		p.nextToken() // Advance to type identifier
		if p.curTokenIs(lexer.IDENTIFIER) || p.curTokenIs(lexer.ANY_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) {
			stmt.Name.TypeAnnotation = p.curToken.Lexeme
		} else {
			stmt.Name.TypeAnnotation = "ANY"
		}
	}

	// Verificar asignación
	if p.peekTokenIs(lexer.EQUAL) {
		p.nextToken() // Consume EQUAL
		p.nextToken() // Advance to expression
		stmt.Value = p.parseExpression(LOWEST)
	} else if p.peekTokenIs(lexer.WALRUS_ASSIGN) {
		p.nextToken() // Consume WALRUS_ASSIGN
		p.nextToken() // Advance to expression
		stmt.Value = p.parseExpression(LOWEST)
		// Set type annotation for walrus assignment
		if stmt.Name.TypeAnnotation == "" {
			stmt.Name.TypeAnnotation = "ANY"
		}
	}

	return stmt
}

// parseWalrusStatement parses a walrus assignment declaration (e.g., x := 10; or NOMBRE STRING := "Wilson").
func (p *Parser) parseWalrusStatement() ast.Statement {
	stmt := &ast.VarStatement{Token: p.curToken} // curToken is IDENTIFIER

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	// Check if constant (all uppercase)
	if strings.ToUpper(stmt.Name.Value) == stmt.Name.Value {
		stmt.IsConstant = true
	}

	p.nextToken() // Consume IDENTIFIER

	// Optional type annotation
	if p.curTokenIs(lexer.IDENTIFIER) || p.curTokenIs(lexer.ANY_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) {
		stmt.Name.TypeAnnotation = p.curToken.Lexeme
		p.nextToken() // Consume type
	} else {
		// If no type specified, assign ANY
		stmt.Name.TypeAnnotation = "ANY"
	}

	if !p.curTokenIs(lexer.WALRUS_ASSIGN) {
		// Not a walrus statement, backtrack
		return nil
	}
	p.nextToken() // Consume WALRUS_ASSIGN

	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseDeclaration parses declarations that start with modifiers (public, private, void).
func (p *Parser) parseDeclaration() ast.Statement {
	modifier := p.curToken

	// Consume the modifier
	p.nextToken()

	if p.curTokenIs(lexer.IDENTIFIER) {
		// Could be function or variable
		if p.peekTokenIs(lexer.LEFT_PAREN) {
			// It's a function
			return p.parseFunctionWithModifier(modifier, false)
		} else {
			// It's a variable
			return p.parseVarWithModifier(modifier)
		}
	} else if p.curTokenIs(lexer.FUNC) {
		// It's a function with explicit func keyword
		return p.parseFunctionWithModifier(modifier, false)
	} else if p.curTokenIs(lexer.CLASS) {
		// It's a class
		return p.parseClassWithModifier(modifier)
	}

	// Invalid
	return p.parseExpressionStatement()
}

// parseFunctionStatement parses a named function declaration (e.g., func myFunc() {} or myFunc() {} or public myFunc() {}).
func (p *Parser) parseFunctionStatement() ast.Statement {
	return p.parseFunctionStatementWithAsync(false)
}

// parseFunctionStatementWithAsync parses a named function declaration, handling the async keyword and modifiers.
func (p *Parser) parseFunctionStatementWithAsync(isAsync bool) ast.Statement {
	token := p.curToken // Puede ser modifier, void, func, o identifier
	var visibility string
	var isVoid bool

	// Check for visibility modifiers
	if p.curTokenIs(lexer.PUBLIC) {
		visibility = "public"
		p.nextToken()
	} else if p.curTokenIs(lexer.PRIVATE) {
		visibility = "private"
		p.nextToken()
	}

	// Check for void modifier
	if p.curTokenIs(lexer.VOID) {
		isVoid = true
		p.nextToken()
	}

	// Check for legacy 'func' keyword (opcional)
	if p.curTokenIs(lexer.FUNC) {
		p.nextToken() // consume 'func'
	}

	// At this point, curToken MUST be the function name (IDENTIFIER)
	if !p.curTokenIs(lexer.IDENTIFIER) {
		p.addError(fmt.Sprintf("expected function name, got %s", p.curToken.Type))
		return nil
	}

	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
	// NO avanzar aquí, parseFunctionLiteralBody lo hará

	funcLit, err := p.parseFunctionLiteralBody(isAsync)
	if err != nil {
		p.addError(err.Error())
		return nil
	}

	return &ast.FuncStatement{
		Token:      token,
		Name:       name,
		Parameters: funcLit.Parameters,
		ReturnType: funcLit.ReturnType,
		Body:       funcLit.Body,
		IsAsync:    isAsync,
		Visibility: visibility,
		IsVoid:     isVoid,
	}
}

// parseFunctionWithModifier parses a function declaration where the modifier has already been consumed.
func (p *Parser) parseFunctionWithModifier(modifier lexer.Token, isAsync bool) ast.Statement {
	var visibility string
	var isVoid bool

	if modifier.Type == lexer.PUBLIC {
		visibility = "public"
	} else if modifier.Type == lexer.PRIVATE {
		visibility = "private"
	} else if modifier.Type == lexer.VOID {
		isVoid = true
	}

	// Check for 'func' keyword
	if p.curTokenIs(lexer.FUNC) {
		p.nextToken() // consume 'func'
	}

	// curToken debe ser el nombre de la función (IDENTIFIER)
	if !p.curTokenIs(lexer.IDENTIFIER) {
		p.addError(fmt.Sprintf("expected function name after modifier, got %s", p.curToken.Type))
		return nil
	}

	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	funcLit, err := p.parseFunctionLiteralBody(isAsync)
	if err != nil {
		p.addError(err.Error())
		return nil
	}

	return &ast.FuncStatement{
		Token:      modifier,
		Name:       name,
		Parameters: funcLit.Parameters,
		ReturnType: funcLit.ReturnType,
		Body:       funcLit.Body,
		IsAsync:    isAsync,
		Visibility: visibility,
		IsVoid:     isVoid,
	}
}

// parseFunctionLiteralBody parses the common parts of a function (parameters, return type, body).
// It assumes the function name (if any) has already been consumed, and expects LEFT_PAREN next.
func (p *Parser) parseFunctionLiteralBody(isAsync bool) (*ast.FunctionLiteral, error) {
	lit := &ast.FunctionLiteral{Token: p.curToken, IsAsync: isAsync}

	// curToken es el nombre de la función, peekToken debe ser LEFT_PAREN
	if !p.peekTokenIs(lexer.LEFT_PAREN) {
		return nil, fmt.Errorf("expected '(' after function name, got %s", p.peekToken.Type)
	}

	p.nextToken() // Ahora curToken es LEFT_PAREN

	lit.Parameters = p.parseFunctionParameters()
	if lit.Parameters == nil {
		return nil, fmt.Errorf("failed to parse function parameters")
	}

	// Después de parseFunctionParameters, curToken es RIGHT_PAREN

	// Verificar si hay tipo de retorno
	if p.peekTokenIs(lexer.COLON) || p.peekTokenIs(lexer.ARROW_RETURN) {
		p.nextToken() // Consume COLON o ARROW_RETURN
		p.nextToken() // Avanzar al tipo
		if p.curTokenIs(lexer.IDENTIFIER) || p.curTokenIs(lexer.ANY_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) {
			lit.ReturnType = p.curToken.Lexeme
		} else {
			return nil, fmt.Errorf("expected return type identifier, got %s", p.curToken.Type)
		}
	} else {
		lit.ReturnType = "ANY"
	}

	// Saltar newlines y avanzar hasta LEFT_BRACE
	p.nextToken()
	p.skipNewlines()

	// Ahora curToken debe ser LEFT_BRACE
	if !p.curTokenIs(lexer.LEFT_BRACE) {
		return nil, fmt.Errorf("expected '{' for function body, got %s (token: %s)", p.curToken.Type, p.curToken.Lexeme)
	}

	lit.Body = p.parseBlockStatement()
	if lit.Body == nil {
		return nil, fmt.Errorf("failed to parse function body")
	}
	return lit, nil
}

// parseFunctionParameters parses the parameters list of a function (e.g., (a int, b string)).
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(lexer.RIGHT_PAREN) {
		p.nextToken() // Consume RIGHT_PAREN
		return identifiers
	}

	p.nextToken() // Advance to first parameter identifier
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	// Check for type after identifier (new syntax: name type)
	if p.peekTokenIs(lexer.IDENTIFIER) || p.peekTokenIs(lexer.ANY_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) {
		p.nextToken() // Consume type token
		ident.TypeAnnotation = p.curToken.Lexeme
	} else if p.peekTokenIs(lexer.COLON) {
		// Legacy support for : type syntax
		p.nextToken() // Consume COLON
		p.nextToken() // Advance to type identifier
		if p.curTokenIs(lexer.IDENTIFIER) || p.curTokenIs(lexer.ANY_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) {
			ident.TypeAnnotation = p.curToken.Lexeme
		} else {
			ident.TypeAnnotation = "ANY"
		}
	} else {
		ident.TypeAnnotation = "ANY"
	}

	identifiers = append(identifiers, ident)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // Consume COMMA
		p.nextToken() // Advance to next parameter identifier
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

		// Check for type after identifier (new syntax: name type)
		if p.peekTokenIs(lexer.IDENTIFIER) || p.peekTokenIs(lexer.ANY_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) {
			p.nextToken() // Consume type token
			ident.TypeAnnotation = p.curToken.Lexeme
		} else if p.peekTokenIs(lexer.COLON) {
			// Legacy support for : type syntax
			p.nextToken() // Consume COLON
			p.nextToken() // Advance to type identifier
			if p.curTokenIs(lexer.IDENTIFIER) || p.curTokenIs(lexer.ANY_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) {
				ident.TypeAnnotation = p.curToken.Lexeme
			} else {
				ident.TypeAnnotation = "ANY"
			}
		} else {
			ident.TypeAnnotation = "ANY"
		}

		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(lexer.RIGHT_PAREN) {
		return nil
	}

	return identifiers
}

// isTypeToken checks if a token is a valid type token.
func (p *Parser) isTypeToken(token lexer.Token) bool {
	switch token.Type {
	case lexer.IDENTIFIER, lexer.ANY_TYPE, lexer.INT_TYPE, lexer.STRING_TYPE, lexer.FLOAT_TYPE, lexer.BOOL_TYPE:
		return true
	default:
		return false
	}
}

// parseTypedVariableDeclaration parses a typed variable declaration: identifier type := value
func (p *Parser) parseTypedVariableDeclaration() ast.Statement {
	stmt := &ast.VarStatement{Token: p.curToken}

	// curToken is the variable name (IDENTIFIER)
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	// Check if constant (all uppercase)
	if strings.ToUpper(stmt.Name.Value) == stmt.Name.Value {
		stmt.IsConstant = true
	}

	p.nextToken() // Consume variable name (IDENTIFIER)

	// Next token must be a type
	if !p.isTypeToken(p.curToken) {
		p.addError(fmt.Sprintf("expected type after variable name, got %s", p.curToken.Type))
		return nil
	}

	stmt.Name.TypeAnnotation = p.curToken.Lexeme
	p.nextToken() // Consume type

	// Next token must be WALRUS_ASSIGN
	if !p.curTokenIs(lexer.WALRUS_ASSIGN) {
		p.addError(fmt.Sprintf("expected ':=' after type, got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // Consume WALRUS_ASSIGN

	// Parse the value expression
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseReturnStatement parses a return statement (e.g., return x + 1;).
func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() // Consume RETURN

	if !p.curTokenIs(lexer.SEMICOLON) && !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.RIGHT_BRACE) && !p.curTokenIs(lexer.EOF) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	p.skipNewlines()
	return stmt
}

// parseExpressionStatement parses a statement that is just an expression (e.g., x + 1;).
func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	p.skipNewlines()
	return stmt
}

// parseIfStatement parses an if-else if-else statement.
func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.curToken}
	p.nextToken() // Consume IF
	stmt.Condition = p.parseExpression(LOWEST)

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.ELIF) {
		p.nextToken() // Consume ELIF
		elseIfStmt := p.parseIfStatement()
		stmt.Alternative = &ast.BlockStatement{
			Statements: []ast.Statement{elseIfStmt},
		}
	} else if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // Consume ELSE
		p.skipNewlines()

		if p.peekTokenIs(lexer.IF) || p.peekTokenIs(lexer.ELIF) {
			p.nextToken() // Consume IF or ELIF for else if
			elseIfStmt := p.parseIfStatement()
			stmt.Alternative = &ast.BlockStatement{
				Statements: []ast.Statement{elseIfStmt},
			}
		} else if p.expectPeek(lexer.LEFT_BRACE) {
			stmt.Alternative = p.parseBlockStatement()
		} else {
			p.addError("expected 'if', 'elif' or '{' after 'else'")
			return nil
		}
	}

	return stmt
}

// parseWhileStatement parses a while loop.
func (p *Parser) parseWhileStatement() ast.Statement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	p.nextToken() // Consume WHILE
	stmt.Condition = p.parseExpression(LOWEST)

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

// parseForStatement parses a for loop, including for-in and traditional for loops.
func (p *Parser) parseForStatement() ast.Statement {
	token := p.curToken
	p.nextToken() // Consume FOR

	// Check if it's a for-in loop: for identifier in ...
	if p.curTokenIs(lexer.IDENTIFIER) && p.peekTokenIs(lexer.IN) {
		stmt := &ast.ForInStatement{Token: token}
		stmt.Identifier = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		p.nextToken() // Consume IDENTIFIER
		p.nextToken() // Consume IN
		stmt.Iterable = p.parseExpression(LOWEST)

		p.skipNewlines()
		if !p.expectPeek(lexer.LEFT_BRACE) {
			return nil
		}

		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	// Traditional for loop: for [init]; [condition]; [post] { body }
	stmt := &ast.ForStatement{Token: token}

	// Parse init statement (optional)
	if !p.curTokenIs(lexer.SEMICOLON) {
		stmt.Init = p.parseStatement()
		p.skipNewlines()
	}

	// Expect first semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}
	p.nextToken() // Consume SEMICOLON

	// Parse condition expression (optional)
	if !p.curTokenIs(lexer.SEMICOLON) {
		stmt.Condition = p.parseExpression(LOWEST)
		p.skipNewlines()
	}

	// Expect second semicolon
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}
	p.nextToken() // Consume SEMICOLON

	// Parse post statement (optional)
	if !p.curTokenIs(lexer.LEFT_BRACE) {
		stmt.Post = p.parseStatement()
		p.skipNewlines()
	}

	// Parse body
	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

// parseBlockStatement parses a block of statements enclosed in curly braces.
// It assumes the LEFT_BRACE is the current token.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}

	// curToken debe ser LEFT_BRACE
	if !p.curTokenIs(lexer.LEFT_BRACE) {
		p.addError(fmt.Sprintf("expected LEFT_BRACE at start of block, got %s", p.curToken.Type))
		return nil
	}

	p.nextToken() // Consumir LEFT_BRACE, avanzar al primer statement

	for !p.curTokenIs(lexer.RIGHT_BRACE) && !p.curTokenIs(lexer.EOF) {
		p.skipNewlines()
		if p.curTokenIs(lexer.RIGHT_BRACE) || p.curTokenIs(lexer.EOF) {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		// parseStatement deja curToken en el último token del statement
		if !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	// Al salir del loop, curToken debe ser RIGHT_BRACE
	if !p.curTokenIs(lexer.RIGHT_BRACE) {
		p.addError(fmt.Sprintf("expected '}', got %s", p.curToken.Type))
		return nil
	}

	// IMPORTANTE: NO hacer nextToken() aquí
	// Dejar curToken en RIGHT_BRACE para que el caller lo maneje
	return block
}

// parseBreakStatement parses a break statement.
func (p *Parser) parseBreakStatement() ast.Statement {
	stmt := &ast.BreakStatement{Token: p.curToken}
	p.skipNewlines()
	return stmt
}

// parseContinueStatement parses a continue statement.
func (p *Parser) parseContinueStatement() ast.Statement {
	stmt := &ast.ContinueStatement{Token: p.curToken}
	p.skipNewlines()
	return stmt
}

// parseClassStatement parses a class declaration.
func (p *Parser) parseClassStatement() ast.Statement {
	token := p.curToken
	var visibility string
	var isVoid bool

	// Check for visibility modifier
	if p.curTokenIs(lexer.PUBLIC) {
		visibility = "public"
		p.nextToken()
	} else if p.curTokenIs(lexer.PRIVATE) {
		visibility = "private"
		p.nextToken()
	}

	// Check for void modifier
	if p.curTokenIs(lexer.VOID) {
		isVoid = true
		p.nextToken()
	}

	// Now expect CLASS keyword
	if !p.curTokenIs(lexer.CLASS) {
		p.addError(fmt.Sprintf("expected 'class', got %s", p.curToken.Type))
		return nil
	}

	p.nextToken() // consume CLASS

	stmt := &ast.ClassStatement{Token: token, Visibility: visibility, IsVoid: isVoid}

	// Ahora curToken debe ser el nombre de la clase
	if !p.curTokenIs(lexer.IDENTIFIER) {
		p.addError(fmt.Sprintf("expected class name, got %s", p.curToken.Type))
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
	p.nextToken() // Avanzar después del nombre

	if p.curTokenIs(lexer.EXTENDS) {
		p.nextToken() // Consume EXTENDS
		if !p.curTokenIs(lexer.IDENTIFIER) {
			p.addError(fmt.Sprintf("expected superclass name, got %s", p.curToken.Type))
			return nil
		}
		stmt.SuperClass = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		p.nextToken() // Avanzar después de superclass
	}

	p.skipNewlines()

	if !p.curTokenIs(lexer.LEFT_BRACE) {
		p.addError(fmt.Sprintf("expected '{' for class body, got %s", p.curToken.Type))
		return nil
	}

	block := p.parseBlockStatement()

	for _, s := range block.Statements {
		switch node := s.(type) {
		case *ast.VarStatement:
			stmt.Attributes = append(stmt.Attributes, node)
		case *ast.FuncStatement:
			method := &ast.MethodStatement{
				Token:      node.Token,
				Name:       node.Name,
				Parameters: node.Parameters,
				ReturnType: node.ReturnType,
				Body:       node.Body,
				IsAsync:    node.IsAsync,
			}
			if node.Name.Value == "init" {
				stmt.InitMethod = &ast.ConstructorStatement{
					Token:      node.Token,
					Name:       node.Name,
					Parameters: node.Parameters,
					Body:       node.Body,
				}
			} else {
				stmt.Methods = append(stmt.Methods, method)
			}
		}
	}

	return stmt
}

// parseClassWithModifier parses a class declaration where the modifier has already been consumed.
func (p *Parser) parseClassWithModifier(modifier lexer.Token) ast.Statement {
	var visibility string
	var isVoid bool

	if modifier.Type == lexer.PUBLIC {
		visibility = "public"
	} else if modifier.Type == lexer.PRIVATE {
		visibility = "private"
	} else if modifier.Type == lexer.VOID {
		isVoid = true
	}

	// curToken debe ser CLASS
	if !p.curTokenIs(lexer.CLASS) {
		p.addError(fmt.Sprintf("expected 'class' after modifier, got %s", p.curToken.Type))
		return nil
	}

	p.nextToken() // consume CLASS

	stmt := &ast.ClassStatement{Token: modifier, Visibility: visibility, IsVoid: isVoid}

	// Ahora curToken debe ser el nombre
	if !p.curTokenIs(lexer.IDENTIFIER) {
		p.addError(fmt.Sprintf("expected class name, got %s", p.curToken.Type))
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
	p.nextToken()

	if p.curTokenIs(lexer.EXTENDS) {
		p.nextToken()
		if !p.curTokenIs(lexer.IDENTIFIER) {
			p.addError(fmt.Sprintf("expected superclass name, got %s", p.curToken.Type))
			return nil
		}
		stmt.SuperClass = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		p.nextToken()
	}

	p.skipNewlines()

	if !p.curTokenIs(lexer.LEFT_BRACE) {
		p.addError(fmt.Sprintf("expected '{' for class body, got %s", p.curToken.Type))
		return nil
	}

	block := p.parseBlockStatement()

	for _, s := range block.Statements {
		switch node := s.(type) {
		case *ast.VarStatement:
			stmt.Attributes = append(stmt.Attributes, node)
		case *ast.FuncStatement:
			method := &ast.MethodStatement{
				Token:      node.Token,
				Name:       node.Name,
				Parameters: node.Parameters,
				ReturnType: node.ReturnType,
				Body:       node.Body,
				IsAsync:    node.IsAsync,
			}
			if node.Name.Value == "init" {
				stmt.InitMethod = &ast.ConstructorStatement{
					Token:      node.Token,
					Name:       node.Name,
					Parameters: node.Parameters,
					Body:       node.Body,
				}
			} else {
				stmt.Methods = append(stmt.Methods, method)
			}
		}
	}

	return stmt
}

// parseTryStatement parses a try-catch-finally block.
func (p *Parser) parseTryStatement() ast.Statement {
	stmt := &ast.TryStatement{Token: p.curToken}

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	stmt.TryBlock = p.parseBlockStatement()

	if p.peekTokenIs(lexer.CATCH) {
		p.nextToken() // Consume CATCH
		var catchParam *ast.Identifier

		if p.peekTokenIs(lexer.LEFT_PAREN) {
			p.nextToken() // Consume LEFT_PAREN
			if p.expectPeek(lexer.IDENTIFIER) {
				catchParam = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
			}
			p.expectPeek(lexer.RIGHT_PAREN)
		}

		p.skipNewlines()
		if !p.expectPeek(lexer.LEFT_BRACE) {
			return nil
		}

		stmt.CatchClause = &ast.CatchClause{
			Token:      p.curToken,
			Parameter:  catchParam,
			CatchBlock: p.parseBlockStatement(),
		}
	}

	if p.peekTokenIs(lexer.FINALLY) {
		p.nextToken() // Consume FINALLY
		p.skipNewlines()
		if p.expectPeek(lexer.LEFT_BRACE) {
			stmt.FinallyBlock = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseThrowStatement parses a throw statement.
func (p *Parser) parseThrowStatement() ast.Statement {
	stmt := &ast.ThrowStatement{Token: p.curToken}
	p.nextToken() // Consume THROW
	stmt.Exception = p.parseExpression(LOWEST)
	return stmt
}

// parseImportStatement parses an import statement.
// Supports both: import "module/path" and import moduleName
func (p *Parser) parseImportStatement() ast.Statement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	// Peek ahead to see if it's a string literal or identifier
	if p.peekTokenIs(lexer.STRING) {
		p.nextToken() // consume STRING
		// For string imports like import "std/math"
		stmt.ModulePath = strings.Trim(p.curToken.Lexeme, `"`)
		return stmt
	} else if p.peekTokenIs(lexer.IDENTIFIER) {
		p.nextToken() // consume IDENTIFIER
		// For identifier imports like import math
		stmt.ModuleName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		return stmt
	} else {
		p.addError(fmt.Sprintf("expected string literal or identifier after 'import', got %s", p.peekToken.Type))
		return nil
	}
}


// parseExportStatement parses an export statement.
func (p *Parser) parseExportStatement() ast.Statement {
	stmt := &ast.ExportStatement{Token: p.curToken}
	p.nextToken() // Consume EXPORT
	stmt.Declaration = p.parseStatement()
	return stmt
}

// parseForInLoop parses a for-in loop: variable for condition { ... }
func (p *Parser) parseForInLoop() ast.Statement {
	// curToken is the variable, peekToken is FOR
	varStmt := p.parseWalrusStatement() // Parse variable assignment
	if varStmt == nil {
		p.addError("expected variable declaration in for loop")
		return nil
	}

	p.nextToken() // Consume FOR token

	// Parse condition (expression that evaluates to boolean)
	condition := p.parseExpression(LOWEST)
	if condition == nil {
		p.addError("expected condition in for loop")
		return nil
	}

	// Parse block
	p.skipNewlines()
	block := p.parseBlockStatement()
	if block == nil {
		p.addError("expected block statement in for loop")
		return nil
	}

	// Create a while loop equivalent for for-in syntax
	return &ast.WhileStatement{
		Token:     p.curToken,
		Condition: condition,
		Body:      block,
	}
}

// parseSwitchStatement parses a switch statement.
func (p *Parser) parseSwitchStatement() ast.Statement {
	stmt := &ast.SwitchStatement{Token: p.curToken}
	p.nextToken() // Consume SWITCH
	stmt.Expression = p.parseExpression(LOWEST)

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	p.nextToken() // Consume LEFT_BRACE
	p.skipNewlines()

	for p.curTokenIs(lexer.CASE) || p.curTokenIs(lexer.DEFAULT) {
		caseClause := &ast.CaseClause{Token: p.curToken}

		if p.curTokenIs(lexer.CASE) {
			p.nextToken() // Consume CASE
			caseClause.Expression = p.parseExpression(LOWEST)
		} else {
			p.nextToken() // Consume DEFAULT
		}

		p.skipNewlines()
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // Consume COLON
		block := &ast.BlockStatement{Statements: []ast.Statement{}}

		for !p.curTokenIs(lexer.CASE) && !p.curTokenIs(lexer.DEFAULT) && !p.curTokenIs(lexer.RIGHT_BRACE) && !p.curTokenIs(lexer.EOF) {
			p.skipNewlines()
			if p.curTokenIs(lexer.CASE) || p.curTokenIs(lexer.DEFAULT) || p.curTokenIs(lexer.RIGHT_BRACE) || p.curTokenIs(lexer.EOF) {
				break
			}
			stmt := p.parseStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}
			p.nextToken()
			p.skipNewlines()
		}

		caseClause.Body = block
		stmt.Cases = append(stmt.Cases, caseClause)
	}

	return stmt
}

// parseMatchStatement parses a match statement.
func (p *Parser) parseMatchStatement() ast.Statement {
	stmt := &ast.MatchStatement{Token: p.curToken}
	p.nextToken() // Consume MATCH
	stmt.Expression = p.parseExpression(LOWEST)

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	p.nextToken() // Consume LEFT_BRACE

	for p.curTokenIs(lexer.CASE) || p.curTokenIs(lexer.DEFAULT) {
		patternCase := &ast.PatternCase{Token: p.curToken}

		if p.curTokenIs(lexer.CASE) {
			p.nextToken() // Consume CASE
			patternCase.Pattern = p.parsePattern()
		}

		p.skipNewlines()
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // Consume COLON
		block := &ast.BlockStatement{Statements: []ast.Statement{}}

		for !p.curTokenIs(lexer.CASE) && !p.curTokenIs(lexer.DEFAULT) && !p.curTokenIs(lexer.RIGHT_BRACE) && !p.curTokenIs(lexer.EOF) {
			p.skipNewlines()
			if p.curTokenIs(lexer.CASE) || p.curTokenIs(lexer.DEFAULT) || p.curTokenIs(lexer.RIGHT_BRACE) || p.curTokenIs(lexer.EOF) {
				break
			}
			s := p.parseStatement()
			if s != nil {
				block.Statements = append(block.Statements, s)
			}
			p.nextToken()
		}

		patternCase.Body = block
		stmt.Cases = append(stmt.Cases, patternCase)
	}

	return stmt
}

// parsePattern parses a pattern for match statements.
func (p *Parser) parsePattern() ast.Pattern {
	if p.curTokenIs(lexer.IDENTIFIER) {
		return &ast.VariablePattern{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme},
		}
	}
	return &ast.LiteralPattern{
		Token: p.curToken,
		Value: p.parseExpression(LOWEST),
	}
}

// parseSpawnStatement parses a spawn statement.
func (p *Parser) parseSpawnStatement() ast.Statement {
	stmt := &ast.SpawnStatement{Token: p.curToken}

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

// Expressions

// parseExpression is the main entry point for parsing expressions with precedence.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	p.skipNewlines()

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.addError(fmt.Sprintf("no prefix parse function for %s found", p.curToken.Type))
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && !p.peekTokenIs(lexer.NEWLINE) && !p.peekTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier parses an identifier expression.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
}

// parseNumberLiteral parses a number literal.
func (p *Parser) parseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Token: p.curToken, Value: p.curToken.Literal}
	return lit
}

// parseStringLiteral parses a string literal.
func (p *Parser) parseStringLiteral() ast.Expression {
	value := ""
	if p.curToken.Literal != nil {
		if str, ok := p.curToken.Literal.(string); ok {
			value = str
		}
	}
	return &ast.StringLiteral{Token: p.curToken, Value: value}
}

// parseTemplateStringLiteral parses a template string literal.
func (p *Parser) parseTemplateStringLiteral() ast.Expression {
	value := ""
	if p.curToken.Literal != nil {
		if str, ok := p.curToken.Literal.(string); ok {
			value = str
		}
	}
	return &ast.TemplateStringLiteral{Token: p.curToken, Value: value}
}

// parseBoolean parses a boolean literal (true/false).
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

// parseNullLiteral parses a null literal.
func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

// parseThisExpression parses a 'this' expression.
func (p *Parser) parseThisExpression() ast.Expression {
	return &ast.ThisExpression{Token: p.curToken}
}

// parseSuperExpression parses a 'super' expression.
func (p *Parser) parseSuperExpression() ast.Expression {
	return &ast.SuperExpression{Token: p.curToken}
}

// parseGroupedExpression parses an expression enclosed in parentheses.
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // Consume LEFT_PAREN
	p.skipNewlines()

	if p.curTokenIs(lexer.RIGHT_PAREN) {
		// Empty grouped expression: just return nil to indicate there's no inner expression
		// This is a placeholder - in proper syntax this shouldn't happen but we handle it gracefully
		return nil
	}

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RIGHT_PAREN) {
		return nil
	}

	return exp
}

// parsePrefixExpression parses a prefix expression (e.g., !x, -y).
func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lexeme,
	}
	p.nextToken() // Consume operator
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

// parseInfixExpression parses an infix expression (e.g., x + y).
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lexeme,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken() // Consume operator
	expr.Right = p.parseExpression(precedence)
	return expr
}

// parseAssignmentExpression parses an assignment expression (e.g., x = 10, y += 5).
func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	// The left side of an assignment must be an identifier or an index/dot expression.
	var name ast.Expression
	switch node := left.(type) {
	case *ast.Identifier:
		name = node
	case *ast.IndexExpression:
		name = node
	case *ast.DotExpression:
		name = node
	default:
		p.addError(fmt.Sprintf("left side of assignment must be assignable, got %T", left))
		return nil
	}

	expr := &ast.AssignmentExpression{
		Token:    p.curToken,
		Name:     name,
		Operator: p.curToken.Lexeme,
	}
	p.nextToken() // Consume operator
	expr.Value = p.parseExpression(LOWEST)
	return expr
}

// parseDotExpression parses a dot access expression (e.g., obj.property).
func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	expr := &ast.DotExpression{Token: p.curToken, Left: left}

	if !p.expectPeek(lexer.IDENTIFIER) {
		return nil
	}

	expr.Property = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
	return expr
}

// parseCallExpression parses a function call expression (e.g., func(arg1, arg2)).
// For collection method calls like arr.push(element), it returns CollectionMethodCall instead.
// For module function calls like show.log(x), it returns CallExpression.
func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	// Special handling for show.log calls - treat as regular CallExpression
	if dotExpr, ok := fn.(*ast.DotExpression); ok {
		if leftIdent, ok := dotExpr.Left.(*ast.Identifier); ok && leftIdent.Value == "show" {
			// show.log is special - treat as regular CallExpression
			exp := &ast.CallExpression{Token: p.curToken, Function: fn}
			exp.Arguments = p.parseExpressionList(lexer.RIGHT_PAREN)
			return exp
		}
		// For other dot expressions, treat as collection method calls
		// The semantic analyzer will handle the distinction
		exp := &ast.CollectionMethodCall{
			Token:     p.curToken,
			Object:    dotExpr.Left,
			Method:    dotExpr.Property,
			Arguments: p.parseExpressionList(lexer.RIGHT_PAREN),
		}
		return exp
	}

	// Regular function call
	exp := &ast.CallExpression{Token: p.curToken, Function: fn}
	exp.Arguments = p.parseExpressionList(lexer.RIGHT_PAREN)
	return exp
}

// parseIndexExpression parses an index or slice access expression (e.g., arr[0], arr[1:3], arr[-1]).
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken() // Consume LEFT_BRACKET

	// Parse start index
	exp.Index = p.parseExpression(LOWEST)

	// Check if this is a slice operation (arr[start:end])
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // Consume COLON
		p.nextToken() // Move to end expression
		exp.EndIndex = p.parseExpression(LOWEST)
	}

	// Check for negative indexing: [-something]
	if p.curTokenIs(lexer.MINUS) {
		exp.NegativeIndex = true
	}

	if !p.expectPeek(lexer.RIGHT_BRACKET) {
		return nil
	}

	return exp
}

// parseRangeExpression parses a range expression (e.g., 1..10).
func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expr := &ast.RangeExpression{Token: p.curToken, Start: left}
	p.nextToken() // Consume RANGE
	expr.End = p.parseExpression(SUM)
	return expr
}

// parseListLiteral parses a list literal (e.g., [1, 2, 3]).
func (p *Parser) parseListLiteral() ast.Expression {
	list := &ast.ListLiteral{Token: p.curToken}
	list.Elements = p.parseExpressionList(lexer.RIGHT_BRACKET)
	return list
}

// parseBlockOrCollectionLiteral handles the logic to distinguish between BlockStatement, MapLiteral, and SetLiteral.
// It assumes the LEFT_BRACE has already been consumed.
func (p *Parser) parseBlockOrCollectionLiteral() ast.Expression {
	token := p.curToken // The '{' token (LEFT_BRACE)
	p.nextToken()       // Consume LEFT_BRACE
	p.skipNewlines()

	// If the next token is '}', it's an empty block, map, or set.
	if p.curTokenIs(lexer.RIGHT_BRACE) {
		p.nextToken() // Consume RIGHT_BRACE
		// Default to an empty block for now, as it's the most common.
		// A more robust parser might need to infer context or use type hints.
		return &ast.BlockExpression{Token: token, Block: &ast.BlockStatement{Token: token, Statements: []ast.Statement{}}}
	}

	// Try to parse the first element/key.
	// We need to peek ahead to distinguish between map and set.
	// This requires a more advanced peek mechanism or backtracking.
	// For simplicity, let's try to parse the first element/key.
	// If it's followed by a COLON, it's a map.
	// If it's followed by a COMMA or RIGHT_BRACE, it's a set.
	// Otherwise, it's a block statement.

	// Save current token to potentially backtrack
	curTokenBackup := p.curToken

	// Try to parse the first element/key.
	// We need to peek ahead to distinguish between map and set.
	// For simplicity, let's try to parse the first expression.
	firstExp := p.parseExpression(LOWEST)
	if firstExp == nil {
		// If we couldn't parse an expression, it's likely a block statement starting with a statement.
		// Rewind tokens and parse as a block.
		p.curToken = token // Rewind to LEFT_BRACE
		p.peekToken = curTokenBackup
		block := p.parseBlockStatement()
		if block == nil {
			return nil
		}
		return &ast.BlockExpression{Token: token, Block: block}
	}

	if p.curTokenIs(lexer.COLON) {
		// It's a MapLiteral
		// Rewind tokens to before firstExp and parse as map
		p.curToken = token // Rewind to LEFT_BRACE
		p.peekToken = curTokenBackup
		p.nextToken() // Consume LEFT_BRACE again
		return p.parseMapLiteral()
	} else if p.curTokenIs(lexer.COMMA) || p.curTokenIs(lexer.RIGHT_BRACE) {
		// It's a SetLiteral
		// Rewind tokens to before firstExp and parse as set
		p.curToken = token // Rewind to LEFT_BRACE
		p.peekToken = curTokenBackup
		p.nextToken() // Consume LEFT_BRACE again
		return p.parseSetLiteral()
	} else {
		// If it's not a map or set, it must be a block statement.
		// The firstExp was actually the first expression statement in the block.
		// Rewind tokens and parse as a block.
		p.curToken = token // Rewind to LEFT_BRACE
		p.peekToken = curTokenBackup
		block := p.parseBlockStatement()
		if block == nil {
			return nil
		}
		return &ast.BlockExpression{Token: token, Block: block}
	}
}

// parseMapLiteral parses a map literal (e.g., {key: value, another: 1}).
// It assumes the LEFT_BRACE has already been consumed.
func (p *Parser) parseMapLiteral() ast.Expression {
	m := &ast.MapLiteral{Token: p.curToken, Pairs: make(map[string]ast.Expression)}

	p.skipNewlines()
	if p.peekTokenIs(lexer.RIGHT_BRACE) {
		p.nextToken() // Consume RIGHT_BRACE
		return m
	}

	for !p.peekTokenIs(lexer.RIGHT_BRACE) && !p.peekTokenIs(lexer.EOF) {
		p.skipNewlines()

		// ✅ Verificar si llegamos al final después de una coma trailing
		if p.curTokenIs(lexer.RIGHT_BRACE) {
			break
		}

		key := p.parseExpression(LOWEST)
		if key == nil {
			return nil
		}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // Advance to value
		value := p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}

		// Check if key is a string literal or identifier
		var keyStr string
		if sl, ok := key.(*ast.StringLiteral); ok {
			keyStr = sl.Value
		} else if id, ok := key.(*ast.Identifier); ok {
			keyStr = id.Value
		} else {
			p.addError("map key must be a string literal or identifier")
			return nil
		}
		m.Pairs[keyStr] = value

		p.skipNewlines()
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // Consume COMMA
			p.skipNewlines()
			p.nextToken() // Advance to next key
			// ✅ Continuar el loop - si viene }, el loop lo detectará
		} else if !p.peekTokenIs(lexer.RIGHT_BRACE) {
			p.addError(fmt.Sprintf("expected ',' or '}', got %s", p.peekToken.Type))
			return nil
		}
	}

	if !p.expectPeek(lexer.RIGHT_BRACE) {
		return nil
	}

	return m
}

// parseSetLiteral parses a set literal (e.g., {1, 2, 3}).
// It assumes the LEFT_BRACE has already been consumed.
func (p *Parser) parseSetLiteral() ast.Expression {
	s := &ast.SetLiteral{Token: p.curToken, Elements: []ast.Expression{}} // Token is LEFT_BRACE

	p.skipNewlines()
	if p.peekTokenIs(lexer.RIGHT_BRACE) {
		p.nextToken() // Consume RIGHT_BRACE
		return s
	}

	for !p.peekTokenIs(lexer.RIGHT_BRACE) && !p.peekTokenIs(lexer.EOF) {
		p.nextToken() // Advance to element
		element := p.parseExpression(LOWEST)
		if element == nil {
			return nil
		}
		s.Elements = append(s.Elements, element)

		p.skipNewlines()
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // Consume COMMA
			p.skipNewlines()
		} else if !p.peekTokenIs(lexer.RIGHT_BRACE) {
			p.addError(fmt.Sprintf("expected ',' or '}', got %s", p.peekToken.Type))
			return nil
		}
	}

	if !p.expectPeek(lexer.RIGHT_BRACE) {
		return nil
	}

	return s
}

// parseFunctionLiteralPrefix parses an anonymous function literal used as a prefix expression (e.g., func() {}).
func (p *Parser) parseFunctionLiteralPrefix() ast.Expression {
	p.nextToken()       // Consume FUNC

	funcLit, err := p.parseFunctionLiteralBody(false) // Not async
	if err != nil {
		p.addError(err.Error())
		return nil
	}
	return funcLit
}

// parseAsyncExpression parses an 'async' keyword, which can precede a function declaration
// (async func) or an arrow function expression (async (params) => body).
func (p *Parser) parseAsyncExpression() ast.Expression {
	p.nextToken()       // Advance past 'async'

	p.skipNewlines()

	if p.curTokenIs(lexer.FUNC) {
		p.nextToken()           // Consume 'func'

		funcLit, err := p.parseFunctionLiteralBody(true) // Pass true for isAsync
		if err != nil {
			p.addError(err.Error())
			return nil
		}
		return funcLit // Return the FunctionLiteral as an expression
	} else if p.curTokenIs(lexer.LEFT_PAREN) || p.curTokenIs(lexer.IDENTIFIER) {
		// It's an async arrow function expression (e.g., async (a) => a + 1)
		return p.parseArrowFunctionExpression(true) // Pass true for isAsync
	} else {
		p.addError(fmt.Sprintf("expected 'func', '(' or identifier after 'async', got %s", p.curToken.Type))
		return nil
	}
}

// parseAwaitExpression parses an 'await' expression (e.g., await someAsyncCall()).
func (p *Parser) parseAwaitExpression() ast.Expression {
	token := p.curToken // The 'await' token
	p.nextToken()       // Advance to the expression to await
	exp := p.parseExpression(PREFIX)
	if exp == nil {
		p.addError("expected expression after 'await'")
		return nil
	}
	return &ast.AwaitExpression{Token: token, Argument: exp}
}

// parseIfExpression parses an if expression (e.g., if condition { true_exp } else { false_exp }).
func (p *Parser) parseIfExpression() ast.Expression {
	token := p.curToken // The 'if' token
	p.nextToken()       // Consume 'if'

	condition := p.parseExpression(LOWEST)
	if condition == nil {
		p.addError("expected condition after 'if'")
		return nil
	}

	p.skipNewlines()
	if !p.expectPeek(lexer.LEFT_BRACE) {
		return nil
	}

	consequence := p.parseBlockStatement()
	if consequence == nil {
		p.addError("expected block statement for 'if' consequence")
		return nil
	}

	var alternative *ast.BlockStatement
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // Consume 'else'
		p.skipNewlines()

		if p.peekTokenIs(lexer.IF) {
			p.nextToken() // Consume 'if' for an 'else if'
			// Recursively parse another IfExpression for the 'else if'
			elseIfExp := p.parseIfExpression()
			if elseIfExp != nil {
				// Wrap the IfExpression in a BlockStatement for the Alternative
				alternative = &ast.BlockStatement{Statements: []ast.Statement{&ast.ExpressionStatement{Expression: elseIfExp}}}
			}
		} else if p.expectPeek(lexer.LEFT_BRACE) {
			alternative = p.parseBlockStatement()
		} else {
			p.addError("expected 'if' or '{' after 'else'")
			return nil
		}
	}

	return &ast.IfExpression{
		Token:       token,
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternative,
	}
}

// parseVarExpression is a stub for when 'var' appears in an expression context.
func (p *Parser) parseVarExpression() ast.Expression {
	p.addError(fmt.Sprintf("VAR token is not expected in expression context at %s", p.curToken.String()))
	p.nextToken() // Advance the token to avoid infinite loops in case of error
	return &ast.Identifier{Token: p.curToken, Value: "INVALID_VAR_EXPRESSION"}
}

// parseReturnExpression is a stub for when 'return' appears in an expression context.
func (p *Parser) parseReturnExpression() ast.Expression {
	p.addError(fmt.Sprintf("RETURN token is not expected in expression context at %s", p.curToken.String()))
	p.nextToken() // Advance the token to avoid infinite loops in case of error
	return &ast.Identifier{Token: p.curToken, Value: "INVALID_RETURN_EXPRESSION"}
}

// parseUnexpectedPrefix is a temporary stub for tokens that should not be prefixes.
func (p *Parser) parseUnexpectedPrefix() ast.Expression {
	if p.curToken.Type == lexer.COMMA || p.curToken.Type == lexer.COLON ||
		p.curToken.Type == lexer.ELIF || p.curToken.Type == lexer.ELSE ||
		p.curToken.Type == lexer.RIGHT_BRACE {
		// Ignore these tokens in prefix position as they are handled elsewhere
		p.nextToken() // Advance past the token
		return &ast.Identifier{Token: p.curToken, Value: "IGNORED_SEPARATOR"}
	}
	p.addError(fmt.Sprintf("unexpected token %s in prefix position", p.curToken.Type))
	p.nextToken() // Advance to avoid infinite loops
	return &ast.Identifier{Token: p.curToken, Value: "UNEXPECTED_PREFIX"}
}

// parseErrorToken handles lexer error tokens.
func (p *Parser) parseErrorToken() ast.Expression {
	p.addError(fmt.Sprintf("lexer error: %s", p.curToken.Lexeme))
	p.nextToken() // Advance past the error token
	return &ast.Identifier{Token: p.curToken, Value: "LEXER_ERROR"}
}

// parseNotExpression parses a 'not' prefix expression (e.g., not x).
func (p *Parser) parseNotExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lexeme, // "not"
	}
	p.nextToken() // Consume NOT
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

// parseInExpression parses an 'in' infix expression (e.g., x in list).
func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Lexeme, // "in"
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken() // Consume IN
	expr.Right = p.parseExpression(precedence)
	return expr
}

// parseArrowFunctionExpression parses an arrow function expression (e.g., (a, b) => a + b or (a) => { return a; }).
// It assumes the LEFT_PAREN (or identifier for single param) has already been consumed.
func (p *Parser) parseArrowFunctionExpression(isAsync bool) ast.Expression {
	token := p.curToken // Should be LEFT_PAREN or IDENTIFIER for single param

	var parameters []*ast.Identifier
	if p.curTokenIs(lexer.LEFT_PAREN) {
		parameters = p.parseFunctionParameters() // Consumes LEFT_PAREN and RIGHT_PAREN
		if parameters == nil {
			return nil
		}
	} else if p.curTokenIs(lexer.IDENTIFIER) {
		// Single parameter without parens: e.g., x => x * 2
		param := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		parameters = append(parameters, param)
		p.nextToken() // Consume the identifier
	} else {
		p.addError(fmt.Sprintf("expected '(' or identifier for arrow function parameters, got %s", p.curToken.Type))
		return nil
	}

	// Handle return type annotation (e.g., (a) -> int => a)
	var returnType string
	if p.peekTokenIs(lexer.COLON) || p.peekTokenIs(lexer.ARROW_RETURN) {
		p.nextToken() // Consume COLON or ARROW_RETURN
		if p.expectPeek(lexer.IDENTIFIER) || p.expectPeek(lexer.ANY_TYPE) || p.expectPeek(lexer.INT_TYPE) || p.expectPeek(lexer.STRING_TYPE) || p.expectPeek(lexer.FLOAT_TYPE) || p.expectPeek(lexer.BOOL_TYPE) {
			returnType = p.curToken.Lexeme
		} else {
			p.addError(fmt.Sprintf("expected return type identifier after %s", p.curToken.Lexeme))
			return nil
		}
	} else {
		returnType = "ANY"
	}

	p.skipNewlines()
	if !p.expectPeek(lexer.ARROW_RETURN) {
		p.addError("expected '->' for arrow function body")
		return nil
	}
	p.nextToken() // Consume ARROW_RETURN

	var body ast.Expression
	var blockBody *ast.BlockStatement

	p.skipNewlines()
	if p.curTokenIs(lexer.LEFT_BRACE) {
		blockBody = p.parseBlockStatement()
		if blockBody == nil {
			return nil
		}
	} else {
		body = p.parseExpression(LOWEST)
		if body == nil {
			return nil
		}
	}

	return &ast.ArrowFunctionExpression{
		Token:      token,
		Parameters: parameters,
		ReturnType: returnType,
		Expression: body,
		Body:       blockBody,
		IsAsync:    isAsync,
	}
}

// parseArrowFunctionExpressionInfix parses an arrow function expression when it's used as an infix operator.
// This is for cases like `(a, b) -> a + b` where `(a, b)` is the left expression.
func (p *Parser) parseArrowFunctionExpressionInfix(left ast.Expression) ast.Expression {
	// The 'left' expression should be a GroupedExpression representing parameters.
	// Or, if it's a single identifier, it's a single parameter.
	var parameters []*ast.Identifier
	isAsync := false // Infix arrow functions are not explicitly async unless 'async' was a prefix

	// The 'left' expression should be a grouped expression (e.g., (a, b)) or a single identifier (e.g., x).
	// If it's a grouped expression, parseGroupedExpression returns the inner expression.
	// If the inner expression is a CallExpression (like a function call with arguments),
	// we extract the arguments as parameters.
	// If the inner expression is just an an Identifier, it's a single parameter.
	switch l := left.(type) {
	case *ast.CallExpression:
		// This case handles `(a, b, c)` where `left` is a CallExpression with arguments.
		for _, arg := range l.Arguments {
			if ident, ok := arg.(*ast.Identifier); ok {
				parameters = append(parameters, ident)
			} else {
				p.addError(fmt.Sprintf("expected identifier in arrow function parameters, got %T", arg))
				return nil
			}
		}
	case *ast.Identifier:
		// Single parameter without explicit parentheses: `x -> x * 2`
		parameters = append(parameters, l)
	default:
		p.addError(fmt.Sprintf("expected grouped expression or identifier for arrow function parameters, got %T", left))
		return nil
	}

	// The current token is ARROW_RETURN, consume it.
	token := p.curToken
	p.nextToken()

	var body ast.Expression
	var blockBody *ast.BlockStatement

	p.skipNewlines()
	if p.curTokenIs(lexer.LEFT_BRACE) {
		blockBody = p.parseBlockStatement()
		if blockBody == nil {
			return nil
		}
	} else {
		body = p.parseExpression(LOWEST)
		if body == nil {
			return nil
		}
	}

	return &ast.ArrowFunctionExpression{
		Token:      token,
		Parameters: parameters,
		Expression: body,
		Body:       blockBody,
		IsAsync:    isAsync,
	}
}

// parseExpressionList parses a comma-separated list of expressions until the 'end' token is found.
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken() // Consume 'end' token
		return list
	}

	p.nextToken() // Advance to first expression
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // Consume COMMA
		p.nextToken() // Advance to next expression
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseAsExpression parses an 'as' type conversion expression (e.g., value as Type).
func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	token := p.curToken // The 'as' token

	// Expect an identifier for the type name
	if !p.expectPeek(lexer.IDENTIFIER) {
		return nil
	}

	typeName := p.curToken.Lexeme

	return &ast.AsExpression{
		Token:    token,
		Left:     left,
		TypeName: typeName,
	}
}

// parseModifierInExpression handles modifiers that appear in expression context (should not happen).
func (p *Parser) parseModifierInExpression() ast.Expression {
	p.addError(fmt.Sprintf("modifier '%s' should not appear in expression context", p.curToken.Lexeme))
	p.nextToken() // Advance to avoid infinite loops
	return &ast.Identifier{Token: p.curToken, Value: "INVALID_MODIFIER_IN_EXPRESSION"}
}

// parseWalrusAssignInExpression handles WALRUS_ASSIGN that appears in expression context (should not happen).
func (p *Parser) parseWalrusAssignInExpression() ast.Expression {
	p.addError("':=' should not appear in expression context")
	p.nextToken() // Advance to avoid infinite loops
	return &ast.Identifier{Token: p.curToken, Value: "INVALID_WALRUS_ASSIGN_IN_EXPRESSION"}
}

// Helper functions

// expectPeek checks if the peek token is of the expected type and advances the parser if it is.
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected %s, got %s", t, p.peekToken.Type))
	return false
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// skipNewlines advances the parser past any newline tokens.
func (p *Parser) skipNewlines() {
	for p.curToken.Type == lexer.NEWLINE {
		p.nextToken()
	}
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() int {
	return tokenPrecedence(p.peekToken.Type)
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	return tokenPrecedence(p.curToken.Type)
}

// tokenPrecedence returns the precedence value for a given token type.
func tokenPrecedence(tt lexer.TokenType) int {
	switch tt {
	case lexer.EQUAL, lexer.PLUS_EQUAL, lexer.MINUS_EQUAL, lexer.STAR_EQUAL, lexer.SLASH_EQUAL, lexer.PERCENT_EQUAL:
		return ASSIGN
	case lexer.OR:
		return ANDOR
	case lexer.AND:
		return ANDOR
	case lexer.EQUAL_EQUAL, lexer.BANG_EQUAL:
		return EQUALS
	case lexer.LESS, lexer.LESS_EQUAL, lexer.GREATER, lexer.GREATER_EQUAL:
		return LESSGREATER
	case lexer.PLUS, lexer.MINUS:
		return SUM
	case lexer.SLASH, lexer.STAR, lexer.PERCENT, lexer.FLOOR_DIVIDE:
		return PRODUCT
	case lexer.POWER:
		return POWER_PREC
	case lexer.DOT:
		return CALL
	case lexer.LEFT_PAREN:
		return CALL
	case lexer.LEFT_BRACKET:
		return INDEX
	case lexer.RANGE:
		return SUM
	case lexer.NOT:
		return PREFIX
	case lexer.IN:
		return EQUALS
	case lexer.ARROW_RETURN: // Added for arrow functions
		return ASSIGN // Low precedence, similar to assignment
	case lexer.AS: // Type conversion operator
		return CALL // High precedence, just below parentheses and function calls
	default:
		return LOWEST
	}
}
