// Package errors - Professional Error System for Zylo Language
package errors

import "fmt"

// Error codes and their descriptions for Zylo language
const (
	// Syntax Errors (001-099)
	ZYLO_ERR_001 = "ZYLO_ERR_001"  // Token inesperado
	ZYLO_ERR_002 = "ZYLO_ERR_002"  // Declaración incompleta
	ZYLO_ERR_003 = "ZYLO_ERR_003"  // Expresión incompleta
	ZYLO_ERR_004 = "ZYLO_ERR_004"  // Paréntesis desbalanceado
	ZYLO_ERR_005 = "ZYLO_ERR_005"  // Llave desbalanceada
	ZYLO_ERR_006 = "ZYLO_ERR_006"  // Corchete desbalanceado
	ZYLO_ERR_007 = "ZYLO_ERR_007"  // Operador binario sin operando derecho
	ZYLO_ERR_008 = "ZYLO_ERR_008"  // Operador unario sin operando
	ZYLO_ERR_009 = "ZYLO_ERR_009"  // Separador inesperado

	// Execution Errors (101-199)
	ZYLO_ERR_101 = "ZYLO_ERR_101"  // Tipo incompatible
	ZYLO_ERR_102 = "ZYLO_ERR_102"  // Variable no definida
	ZYLO_ERR_103 = "ZYLO_ERR_103"  // Conversión de tipo fallida
	ZYLO_ERR_104 = "ZYLO_ERR_104"  // Operación inválida entre tipos
	ZYLO_ERR_105 = "ZYLO_ERR_105"  // División por cero
	ZYLO_ERR_106 = "ZYLO_ERR_106"  // Index fuera de rango
	ZYLO_ERR_107 = "ZYLO_ERR_107"  // Función con return inválido
	ZYLO_ERR_108 = "ZYLO_ERR_108"  // Llamada a función inválida
	ZYLO_ERR_109 = "ZYLO_ERR_109"  // Stack overflow
	ZYLO_ERR_110 = "ZYLO_ERR_110"  // Null pointer access

	// Semantic Errors (201-299)
	ZYLO_ERR_201 = "ZYLO_ERR_201"  // Identificador duplicado
	ZYLO_ERR_202 = "ZYLO_ERR_202"  // Tipo no definido
	ZYLO_ERR_203 = "ZYLO_ERR_203"  // Asignación a constante
	ZYLO_ERR_204 = "ZYLO_ERR_204"  // Función duplicada
	ZYLO_ERR_205 = "ZYLO_ERR_205"  // Parámetro duplicado
	ZYLO_ERR_206 = "ZYLO_ERR_206"  // Atributo duplicado
	ZYLO_ERR_207 = "ZYLO_ERR_207"  // Método duplicado
	ZYLO_ERR_208 = "ZYLO_ERR_208"  // Referencia circular detectada
)

// Error type enumeration
type ZyloErrorType string

const (
	ErrorTypeSyntax   ZyloErrorType = "Sintaxis"
	ErrorTypeExecution ZyloErrorType = "Ejecución"
	ErrorTypeSemantic ZyloErrorType = "Semántico"
)

// ErrorInfo contains error information
type ErrorInfo struct {
	Code        string
	Type        ZyloErrorType
	Description string
}

// Error definitions map - Complete professional error system
var ErrorDefinitions = map[string]ErrorInfo{
	// Syntax Errors
	ZYLO_ERR_001: {Code: ZYLO_ERR_001, Type: ErrorTypeSyntax, Description: "Token inesperado en la expresión"},
	ZYLO_ERR_002: {Code: ZYLO_ERR_002, Type: ErrorTypeSyntax, Description: "Declaración incompleta o mal formada"},
	ZYLO_ERR_003: {Code: ZYLO_ERR_003, Type: ErrorTypeSyntax, Description: "Expresión incompleta"},
	ZYLO_ERR_004: {Code: ZYLO_ERR_004, Type: ErrorTypeSyntax, Description: "Paréntesis desbalanceado o faltante"},
	ZYLO_ERR_005: {Code: ZYLO_ERR_005, Type: ErrorTypeSyntax, Description: "Llaves desbalanceadas"},
	ZYLO_ERR_006: {Code: ZYLO_ERR_006, Type: ErrorTypeSyntax, Description: "Corchetes desbalanceados"},
	ZYLO_ERR_007: {Code: ZYLO_ERR_007, Type: ErrorTypeSyntax, Description: "Operador binario sin operando derecho"},
	ZYLO_ERR_008: {Code: ZYLO_ERR_008, Type: ErrorTypeSyntax, Description: "Operador unario sin operando"},
	ZYLO_ERR_009: {Code: ZYLO_ERR_009, Type: ErrorTypeSyntax, Description: "Separador o coma inesperados"},

	// Execution Errors
	ZYLO_ERR_101: {Code: ZYLO_ERR_101, Type: ErrorTypeExecution, Description: "Tipos incompatibles en la operación"},
	ZYLO_ERR_102: {Code: ZYLO_ERR_102, Type: ErrorTypeExecution, Description: "Variable no definida"},
	ZYLO_ERR_103: {Code: ZYLO_ERR_103, Type: ErrorTypeExecution, Description: "Conversión de tipo fallida"},
	ZYLO_ERR_104: {Code: ZYLO_ERR_104, Type: ErrorTypeExecution, Description: "Operación inválida entre los tipos especificados"},
	ZYLO_ERR_105: {Code: ZYLO_ERR_105, Type: ErrorTypeExecution, Description: "División por cero detectada"},
	ZYLO_ERR_106: {Code: ZYLO_ERR_106, Type: ErrorTypeExecution, Description: "Index fuera del rango válido"},
	ZYLO_ERR_107: {Code: ZYLO_ERR_107, Type: ErrorTypeExecution, Description: "Return inválido en función void"},
	ZYLO_ERR_108: {Code: ZYLO_ERR_108, Type: ErrorTypeExecution, Description: "Llamada a función con argumentos inválidos"},
	ZYLO_ERR_109: {Code: ZYLO_ERR_109, Type: ErrorTypeExecution, Description: "Stack overflow en recursión profunda"},
	ZYLO_ERR_110: {Code: ZYLO_ERR_110, Type: ErrorTypeExecution, Description: "Acceso a valor null"},

	// Semantic Errors
	ZYLO_ERR_201: {Code: ZYLO_ERR_201, Type: ErrorTypeSemantic, Description: "Identificador ya definido anteriormente"},
	ZYLO_ERR_202: {Code: ZYLO_ERR_202, Type: ErrorTypeSemantic, Description: "Tipo de dato no definido"},
	ZYLO_ERR_203: {Code: ZYLO_ERR_203, Type: ErrorTypeSemantic, Description: "Asignación a constante no permitida"},
	ZYLO_ERR_204: {Code: ZYLO_ERR_204, Type: ErrorTypeSemantic, Description: "Función ya definida"},
	ZYLO_ERR_205: {Code: ZYLO_ERR_205, Type: ErrorTypeSemantic, Description: "Parámetro duplicado en función"},
	ZYLO_ERR_206: {Code: ZYLO_ERR_206, Type: ErrorTypeSemantic, Description: "Atributo duplicado en clase"},
	ZYLO_ERR_207: {Code: ZYLO_ERR_207, Type: ErrorTypeSemantic, Description: "Método duplicado en clase"},
	ZYLO_ERR_208: {Code: ZYLO_ERR_208, Type: ErrorTypeSemantic, Description: "Referencia circular detectada"},
}

// FormatZyloError formats a Zylo error message according to the professional standard
// Format: [ZYLO_ERR_XXX] Tipo: Descripción profesional. Línea: X, Columna: Y
func FormatZyloError(errorCode string, line int, column int) string {
	info, exists := ErrorDefinitions[errorCode]
	if !exists {
		return fmt.Sprintf("[ZYLO_ERR_000] Error desconocido: Código no encontrado. Línea: %d, Columna: %d", line, column)
	}

	return fmt.Sprintf("[%s] %s: %s. Línea: %d, Columna: %d",
		info.Code, info.Type, info.Description, line, column)
}

// FormatZyloErrorWithContext formats a Zylo error with additional context
// Format: [ZYLO_ERR_XXX] Tipo: Descripción: Contexto adicional. Línea: X, Columna: Y
func FormatZyloErrorWithContext(errorCode string, line int, column int, context string) string {
	info, exists := ErrorDefinitions[errorCode]
	if !exists {
		return fmt.Sprintf("[ZYLO_ERR_000] Error desconocido: Código no encontrado. Línea: %d, Columna: %d", line, column)
	}

	return fmt.Sprintf("[%s] %s: %s. Línea: %d, Columna: %d",
		info.Code, info.Type, context, line, column)
}

// ShowError generates a professional error message for display
func ShowError(errorCode string, line int, column int) string {
	return FormatZyloError(errorCode, line, column)
}

// ShowErrorWithContext generates a professional error message with additional context
func ShowErrorWithContext(errorCode string, line int, column int, context string) string {
	return FormatZyloErrorWithContext(errorCode, line, column, context)
}
