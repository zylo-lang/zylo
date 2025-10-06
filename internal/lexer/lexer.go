package lexer

import (
	"strconv"
	"strings"
	"unicode"
)

// Lexer se encarga de convertir el código fuente en una secuencia de tokens.
type Lexer struct {
	source      []rune // El código fuente como un slice de runas para soportar Unicode.
	start       int    // Posición de inicio del token actual.
	current     int    // Posición actual en el slice de runas.
	line        int    // Línea actual.
	column      int    // Columna actual en la línea.
	startLine   int    // Línea de inicio del token actual.
	startColumn int    // Columna de inicio del token actual.
}

// New crea un nuevo Lexer para el código fuente proporcionado.
func New(source string) *Lexer {
	return &Lexer{
		source: []rune(source),
		line:   1,
		column: 1,
	}
}

// isAtEnd comprueba si hemos llegado al final del código fuente.
func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

// advance consume la runa actual y avanza la posición.
func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0 // EOF
	}
	r := l.source[l.current]
	l.current++
	if r == '\n' {
		l.line++
		l.column = 1 // Se reinicia a columna 1
	} else {
		l.column++
	}
	return r
}

// peek devuelve la runa actual sin consumirla.
func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

// peekNext devuelve la siguiente runa sin consumirla.
func (l *Lexer) peekNext() rune {
	if l.current+1 >= len(l.source) {
		return 0
	}
	return l.source[l.current+1]
}

// match comprueba si la runa actual coincide con la esperada. Si es así, la consume.
func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.advance()
	return true
}

// makeToken crea un nuevo token con la información de posición actual.
func (l *Lexer) makeToken(tokenType TokenType, literal interface{}) Token {
	lexeme := string(l.source[l.start:l.current])
	endLine := l.line
	endCol := l.column - 1
	if endCol < 1 {
		endCol = 1
	}
	if tokenType == NEWLINE {
		endLine = l.startLine
		endCol = l.startColumn
	}
	if tokenType == EOF {
		endCol = 0
	}

	return Token{
		Type:      tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
		StartLine: l.startLine,
		StartCol:  l.startColumn,
		EndLine:   endLine,
		EndCol:    endCol,
	}
}

// errorToken crea un token de error.
func (l *Lexer) errorToken(message string) Token {
	return Token{
		Type:      "ERROR",
		Lexeme:    message,
		StartLine: l.line,
		StartCol:  l.column,
		EndLine:   l.line,
		EndCol:    l.column,
	}
}

// skipWhitespace consume todos los espacios en blanco y tabulaciones, pero no los newlines.
func (l *Lexer) skipWhitespace() {
	for {
		switch l.peek() {
		case ' ', '\r', '\t':
			l.advance()
		case '\n':
			return
		case '/':
			if l.peekNext() == '/' {
				for l.peek() != '\n' && !l.isAtEnd() {
					l.advance()
				}
			} else if l.peekNext() == '*' {
				l.advance()
				l.advance()
				l.skipMultiLineComment()
			} else {
				return
			}
		case '#':
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		default:
			return
		}
	}
}

// skipMultiLineComment consume un comentario multilínea, incluyendo anidamiento.
func (l *Lexer) skipMultiLineComment() {
	nestingLevel := 1
	for nestingLevel > 0 && !l.isAtEnd() {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance()
			l.advance()
			nestingLevel--
		} else if l.peek() == '/' && l.peekNext() == '*' {
			l.advance()
			l.advance()
			nestingLevel++
		} else {
			l.advance()
		}
	}
}

// isAlpha comprueba si una runa es una letra o un guion bajo.
func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isDigit comprueba si una runa es un dígito.
func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// isHexDigit comprueba si una runa es un dígito hexadecimal.
func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// identifier procesa un identificador o una palabra clave.
func (l *Lexer) identifier() Token {
	for isAlpha(l.peek()) || isDigit(l.peek()) {
		l.advance()
	}
	text := string(l.source[l.start:l.current])
	tokenType, isKeyword := keywords[text]
	if !isKeyword {
		tokenType = IDENTIFIER
	}
	return l.makeToken(tokenType, nil)
}

// number procesa un número literal.
func (l *Lexer) number() Token {
	isFloat := false
	for isDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && isDigit(l.peekNext()) {
		isFloat = true
		l.advance()
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	lexeme := string(l.source[l.start:l.current])
	if isFloat {
		value, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			return l.errorToken("Invalid float number.")
		}
		return l.makeToken(NUMBER, value)
	}

	value, err := strconv.ParseInt(lexeme, 10, 64)
	if err != nil {
		return l.errorToken("Invalid integer number.")
	}
	return l.makeToken(NUMBER, value)
}

// stringLiteral procesa una cadena literal entre comillas simples o dobles.
func (l *Lexer) stringLiteral(quote rune) Token {
	var builder strings.Builder
	for {
		if l.peek() == quote || l.isAtEnd() {
			break
		}

		if l.peek() == '\\' {
			l.advance()
			switch l.peek() {
			case 'n':
				builder.WriteRune('\n')
			case 't':
				builder.WriteRune('\t')
			case '"':
				builder.WriteRune('"')
			case '\'':
				builder.WriteRune('\'')
			case '\\':
				builder.WriteRune('\\')
			case 'u':
				l.advance()
				hex := make([]rune, 4)
				for i := 0; i < 4; i++ {
					if !isHexDigit(l.peek()) {
						return l.errorToken("Invalid Unicode escape sequence: expected 4 hex digits.")
					}
					hex[i] = l.advance()
				}
				hexVal, err := strconv.ParseInt(string(hex), 16, 32)
				if err != nil {
					return l.errorToken("Invalid Unicode escape sequence.")
				}
				builder.WriteRune(rune(hexVal))
				continue
			default:
				builder.WriteRune('\\')
				builder.WriteRune(l.peek())
			}
			l.advance()
		} else {
			if l.peek() == '\n' {
				return l.errorToken("Unterminated string.")
			}
			builder.WriteRune(l.advance())
		}
	}

	if l.isAtEnd() {
		return l.errorToken("Unterminated string.")
	}

	l.advance()
	return l.makeToken(STRING, builder.String())
}

// tripleQuotedStringLiteral procesa una cadena multilínea.
func (l *Lexer) tripleQuotedStringLiteral() Token {
	l.advance()
	l.advance()

	var builder strings.Builder
	for {
		if l.isAtEnd() {
			return l.errorToken("Unterminated multi-line string.")
		}
		if l.peek() == '"' && l.peekNext() == '"' && l.peekN(2) == '"' {
			break
		}
		builder.WriteRune(l.advance())
	}

	l.advance()
	l.advance()
	l.advance()

	content := builder.String()
	if len(content) > 0 && content[0] == '\n' {
		content = content[1:]
	}

	return l.makeToken(STRING, content)
}

// templateStringLiteral procesa una cadena de plantilla (template string) entre backticks.
func (l *Lexer) templateStringLiteral() Token {
	var builder strings.Builder
	for {
		if l.peek() == '`' || l.isAtEnd() {
			break
		}
		if l.peek() == '$' && l.peekNext() == '{' {
			l.advance()
			l.advance()
			for !l.isAtEnd() && l.peek() != '}' {
				l.advance()
			}
			if l.peek() == '}' {
				l.advance()
			} else {
				return l.errorToken("Unterminated template string interpolation.")
			}
		} else {
			builder.WriteRune(l.advance())
		}
	}

	if l.isAtEnd() {
		return l.errorToken("Unterminated template string.")
	}

	l.advance()
	return l.makeToken(TEMPLATE_STRING, builder.String())
}

// peekN devuelve la runa en la posición current + n.
func (l *Lexer) peekN(n int) rune {
	if l.current+n >= len(l.source) {
		return 0
	}
	return l.source[l.current+n]
}

// NextToken escanea y devuelve el siguiente token del código fuente.
func (l *Lexer) NextToken() Token {
	// Skip BOM if present
	if l.current == 0 && !l.isAtEnd() && l.source[0] == '\ufeff' {
		l.current = 1
	}
	l.skipWhitespace()
	l.start = l.current
	l.startLine = l.line
	l.startColumn = l.column

	if l.isAtEnd() {
		return l.makeToken(EOF, nil)
	}

	r := l.advance()

	if isAlpha(r) {
		return l.identifier()
	}
	if isDigit(r) {
		return l.number()
	}

	switch r {
	case '(':
		return l.makeToken(LEFT_PAREN, nil)
	case ')':
		return l.makeToken(RIGHT_PAREN, nil)
	case '{':
		return l.makeToken(LEFT_BRACE, nil)
	case '}':
		return l.makeToken(RIGHT_BRACE, nil)
	case '[':
		return l.makeToken(LEFT_BRACKET, nil)
	case ']':
		return l.makeToken(RIGHT_BRACKET, nil)
	case ';':
		return l.makeToken(SEMICOLON, nil)
	case ',':
		return l.makeToken(COMMA, nil)
	case '.':
		if l.match('.') {
			return l.makeToken(RANGE, nil)
		}
		return l.makeToken(DOT, nil)
	case '-':
		if l.match('>') {
			return l.makeToken(ARROW_RETURN, nil)
		}
		if l.match('=') {
			return l.makeToken(MINUS_EQUAL, nil)
		}
		return l.makeToken(MINUS, nil)
	case '+':
		if l.match('=') {
			return l.makeToken(PLUS_EQUAL, nil)
		}
		return l.makeToken(PLUS, nil)
	case '/':
		if l.match('=') {
			return l.makeToken(SLASH_EQUAL, nil)
		}
		if l.match('/') {
			return l.makeToken(FLOOR_DIVIDE, nil)
		}
		return l.makeToken(SLASH, nil)
	case '*':
		if l.match('=') {
			return l.makeToken(STAR_EQUAL, nil)
		}
		if l.match('*') {
			return l.makeToken(POWER, nil)
		}
		return l.makeToken(STAR, nil)
	case '%':
		if l.match('=') {
			return l.makeToken(PERCENT_EQUAL, nil)
		}
		return l.makeToken(PERCENT, nil)
	case '^':
		return l.makeToken(POWER, nil) // Caret for exponentiation
	case ':':
		// Skip whitespace after :
		for l.peek() == ' ' || l.peek() == '\t' {
			l.advance()
		}
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(WALRUS_ASSIGN, nil)
		}
		return l.makeToken(COLON, nil)
	case '!':
		if l.match('=') {
			return l.makeToken(BANG_EQUAL, nil)
		}
		return l.makeToken(BANG, nil)
	case '=':
		if l.match('=') {
			return l.makeToken(EQUAL_EQUAL, nil)
		}
		if l.match('>') {
			return l.makeToken(ARROW_FUNC, nil)
		}
		return l.makeToken(EQUAL, nil)
	case '<':
		if l.match('=') {
			return l.makeToken(LESS_EQUAL, nil)
		}
		return l.makeToken(LESS, nil)
	case '>':
		if l.match('=') {
			return l.makeToken(GREATER_EQUAL, nil)
		}
		return l.makeToken(GREATER, nil)
	case '&':
		if l.match('&') {
			return l.makeToken(AND, nil)
		}
		return l.errorToken("Unexpected character '&'. Did you mean 'and' or '&&'?")
	case '|':
		if l.match('|') {
			return l.makeToken(OR, nil)
		}
		return l.errorToken("Unexpected character '|'. Did you mean 'or' or '||'?")
	case '\n':
		return l.makeToken(NEWLINE, nil)
	case '"':
		if l.peek() == '"' && l.peekNext() == '"' {
			return l.tripleQuotedStringLiteral()
		}
		return l.stringLiteral('"')
	case '\'':
		return l.stringLiteral('\'')
	case '`':
		return l.templateStringLiteral()
	}

	return l.errorToken("Unexpected character.")
}
