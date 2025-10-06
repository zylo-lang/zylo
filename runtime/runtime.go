package zyloruntime


import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// --- Interfaz de Objeto ---

type ObjectType string

const (
	INTEGER_OBJ  = "INTEGER"
	FLOAT_OBJ    = "FLOAT"
	STRING_OBJ   = "STRING"
	BOOL_OBJ     = "BOOL"
	NULL_OBJ     = "NULL"
	LIST_OBJ     = "LIST"
	MAP_OBJ      = "MAP"
	ERROR_OBJ    = "ERROR"
	BUILTIN_OBJ  = "BUILTIN"
	FUNCTION_OBJ = "FUNCTION"
	FUTURE_OBJ   = "FUTURE"
)

type ZyloObject interface {
	Type() ObjectType
	Inspect() string
}

// --- Tipos de Objetos ---

type Integer struct{ Value int64 }

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return strconv.FormatInt(i.Value, 10) }

type Float struct{ Value float64 }

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

type String struct{ Value string }

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type Bool struct{ Value bool }

func (b *Bool) Type() ObjectType { return BOOL_OBJ }
func (b *Bool) Inspect() string  { return strconv.FormatBool(b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type List struct{ Elements []ZyloObject }

func (l *List) Type() ObjectType { return LIST_OBJ }
func (l *List) Inspect() string {
	var out strings.Builder
	elements := []string{}
	for _, e := range l.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

type Map struct{ Pairs map[string]ZyloObject }

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
	var out strings.Builder
	pairs := []string{}
	for k, v := range m.Pairs {
		pairs = append(pairs, fmt.Sprintf("%q: %s", k, v.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type Error struct{ Message string }

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Builtin representa una función built-in del lenguaje
type Builtin struct {
	Fn func(...ZyloObject) ZyloObject
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

// Future representa un resultado de una operación asíncrona
type Future struct {
	Result chan ZyloObject
	value  ZyloObject
	once   bool
}

func (f *Future) Type() ObjectType { return FUTURE_OBJ }
func (f *Future) Inspect() string  { return "future" }

// --- Funciones de Creación ---

func NewInteger(value int64) *Integer {
	return &Integer{Value: value}
}

func NewFloat(value float64) *Float {
	return &Float{Value: value}
}

func NewString(value string) *String {
	return &String{Value: value}
}

func NewBool(value bool) *Bool {
	return &Bool{Value: value}
}

func NewNull() *Null {
	return &Null{}
}

func NewList(elements ...ZyloObject) *List {
	return &List{Elements: elements}
}

func NewMap(pairs map[string]ZyloObject) *Map {
	if pairs == nil {
		pairs = make(map[string]ZyloObject)
	}
	return &Map{Pairs: pairs}
}

func NewError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func NewBuiltin(fn func(...ZyloObject) ZyloObject) *Builtin {
	return &Builtin{Fn: fn}
}

// --- Funciones Built-in ---

// --- Operaciones de String ---

func builtinSplit(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("split expects 2 arguments, got %d", len(args))
	}
	s, ok1 := args[0].(*String)
	delimiter, ok2 := args[1].(*String)
	if !ok1 {
		return NewError("split expects a string for the first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("split expects a string for the second argument, got %s", args[1].Type())
	}
	parts := strings.Split(s.Value, delimiter.Value)
	elements := make([]ZyloObject, len(parts))
	for i, p := range parts {
		elements[i] = &String{Value: p}
	}
	return &List{Elements: elements}
}

func builtinJoin(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("join expects 2 arguments, got %d", len(args))
	}
	l, ok1 := args[0].(*List)
	delimiter, ok2 := args[1].(*String)
	if !ok1 {
		return NewError("join expects a list for the first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("join expects a string for the second argument, got %s", args[1].Type())
	}
	elements := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		if s, ok := e.(*String); ok {
			elements[i] = s.Value
		} else {
			return NewError("join expects a list of strings, but found element of type %s", e.Type())
		}
	}
	return &String{Value: strings.Join(elements, delimiter.Value)}
}

func builtinSubstring(args ...ZyloObject) ZyloObject {
	if len(args) != 3 {
		return NewError("substring expects 3 arguments, got %d", len(args))
	}
	s, ok1 := args[0].(*String)
	st, ok2 := args[1].(*Integer)
	en, ok3 := args[2].(*Integer)
	if !ok1 {
		return NewError("substring expects a string for the first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("substring expects an integer for the second argument, got %s", args[1].Type())
	}
	if !ok3 {
		return NewError("substring expects an integer for the third argument, got %s", args[2].Type())
	}
	if st.Value < 0 || en.Value > int64(len(s.Value)) || st.Value > en.Value {
		return NewError("substring indices out of bounds: start %d, end %d for string length %d", st.Value, en.Value, len(s.Value))
	}
	return &String{Value: s.Value[st.Value:en.Value]}
}

func builtinReplace(args ...ZyloObject) ZyloObject {
	if len(args) != 3 {
		return NewError("replace expects 3 arguments, got %d", len(args))
	}
	s, ok1 := args[0].(*String)
	o, ok2 := args[1].(*String)
	n, ok3 := args[2].(*String)
	if !ok1 {
		return NewError("replace expects a string for the first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("replace expects a string for the second argument, got %s", args[1].Type())
	}
	if !ok3 {
		return NewError("replace expects a string for the third argument, got %s", args[2].Type())
	}
	return &String{Value: strings.ReplaceAll(s.Value, o.Value, n.Value)}
}

func builtinTrim(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("trim expects 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*String)
	if !ok {
		return NewError("trim expects a string, got %s", args[0].Type())
	}
	return &String{Value: strings.TrimSpace(s.Value)}
}

// --- Operaciones de Lista ---

func builtinLen(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("len expects 1 argument, got %d", len(args))
	}
	switch o := args[0].(type) {
	case *String:
		return &Integer{Value: int64(len(o.Value))}
	case *List:
		return &Integer{Value: int64(len(o.Elements))}
	case *Map:
		return &Integer{Value: int64(len(o.Pairs))}
	default:
		return NewError("len does not support type %s", args[0].Type())
	}
}

func builtinAppend(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("append expects 2 arguments, got %d", len(args))
	}
	l, ok := args[0].(*List)
	if !ok {
		return NewError("append expects a list as first argument, got %s", args[0].Type())
	}
	newElements := make([]ZyloObject, len(l.Elements)+1)
	copy(newElements, l.Elements)
	newElements[len(l.Elements)] = args[1]
	return &List{Elements: newElements}
}

func builtinPrepend(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("prepend expects 2 arguments, got %d", len(args))
	}
	l, ok := args[0].(*List)
	if !ok {
		return NewError("prepend expects a list as first argument, got %s", args[0].Type())
	}
	return &List{Elements: append([]ZyloObject{args[1]}, l.Elements...)}
}

func builtinSlice(args ...ZyloObject) ZyloObject {
	if len(args) != 3 {
		return NewError("slice expects 3 arguments, got %d", len(args))
	}
	l, ok1 := args[0].(*List)
	st, ok2 := args[1].(*Integer)
	en, ok3 := args[2].(*Integer)
	if !ok1 {
		return NewError("slice expects a list as first argument, got %s", args[0].Type())
	}
	if !ok2 || !ok3 {
		return NewError("slice expects integers for start and end indices")
	}
	if st.Value < 0 || en.Value > int64(len(l.Elements)) || st.Value > en.Value {
		return NewError("slice indices out of bounds")
	}
	newElements := make([]ZyloObject, en.Value-st.Value)
	copy(newElements, l.Elements[st.Value:en.Value])
	return &List{Elements: newElements}
}

func builtinSort(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("sort expects 1 argument, got %d", len(args))
	}
	l, ok := args[0].(*List)
	if !ok {
		return NewError("sort expects a list, got %s", args[0].Type())
	}
	// Crear una copia para no modificar la lista original
	newElements := make([]ZyloObject, len(l.Elements))
	copy(newElements, l.Elements)
	
	sort.SliceStable(newElements, func(i, j int) bool {
		return newElements[i].Inspect() < newElements[j].Inspect()
	})
	return &List{Elements: newElements}
}

func builtinReverse(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("reverse expects 1 argument, got %d", len(args))
	}
	l, ok := args[0].(*List)
	if !ok {
		return NewError("reverse expects a list, got %s", args[0].Type())
	}
	// Crear una copia para no modificar la lista original
	newElements := make([]ZyloObject, len(l.Elements))
	copy(newElements, l.Elements)
	
	for i, j := 0, len(newElements)-1; i < j; i, j = i+1, j-1 {
		newElements[i], newElements[j] = newElements[j], newElements[i]
	}
	return &List{Elements: newElements}
}

// --- Conversiones de Tipo ---

func inspectValue(obj interface{}) string {
	switch v := obj.(type) {
	case ZyloObject:
		return v.Inspect()
	case []interface{}:
		elements := make([]string, len(v))
		for i, e := range v {
			elements[i] = inspectValue(e)
		}
		return "[" + strings.Join(elements, ", ") + "]"
	case map[string]interface{}:
		pairs := make([]string, 0, len(v))
		for k, val := range v {
			pairs = append(pairs, fmt.Sprintf("%q: %s", k, inspectValue(val)))
		}
		return "{" + strings.Join(pairs, ", ") + "}"
	case string:
		return fmt.Sprintf("%q", v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("<unknown: %T>", v)
	}
}

func builtinToString(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("ToString expects 1 argument, got %d", len(args))
	}
	return &String{Value: inspectValue(args[0])}
}

// --- Advanced Type Conversions ---

// AdvancedAutoCast realiza conversión automática inteligente entre tipos
func AdvancedAutoCast(left, right ZyloObject, operator string) (ZyloObject, ZyloObject, bool) {
	// Auto-casting entre tipos numéricos
	if isNumeric(left) && isNumeric(right) {
		convertedLeft, convertedRight, changed := autoCastNumeric(left, right)
		if changed {
			return convertedLeft, convertedRight, true
		}
	}

	// Auto-casting para concatenación
	if operator == "+" {
		if left.Type() == STRING_OBJ && right.Type() != STRING_OBJ && isDisplayable(right) {
			return left, builtinToString(right), true
		} else if right.Type() == STRING_OBJ && left.Type() != STRING_OBJ && isDisplayable(left) {
			return builtinToString(left), right, true
		}
	}

	return left, right, false
}

// autoCastNumeric convierte tipos numéricos automáticamente para operaciones
func autoCastNumeric(left, right ZyloObject) (ZyloObject, ZyloObject, bool) {
	// Convertir siempre al tipo más general: int → float
	inputLeft, leftIsInt := left.(*Integer)
	inputRight, rightIsInt := right.(*Integer)

	if leftIsInt && rightIsInt {
		return left, right, false
	}

	_, leftIsFloat := left.(*Float)
	_, rightIsFloat := right.(*Float)

	if leftIsFloat && rightIsFloat {
		return left, right, false
	}

	if leftIsInt && rightIsFloat {
		return &Float{Value: float64(inputLeft.Value)}, right, true
	}

	if leftIsFloat && rightIsInt {
		return left, &Float{Value: float64(inputRight.Value)}, true
	}

	return left, right, false
}

// isNumeric determina si un objeto es numérico
func isNumeric(obj ZyloObject) bool {
	return obj.Type() == INTEGER_OBJ || obj.Type() == FLOAT_OBJ
}

// isDisplayable determina si un objeto se puede convertir a string
func isDisplayable(obj ZyloObject) bool {
	switch obj.Type() {
	case INTEGER_OBJ, FLOAT_OBJ, STRING_OBJ, BOOL_OBJ, NULL_OBJ:
		return true
	default:
		return false
	}
}

// --- Enhanced Error Handling ---

// ZyloTryCatch representa una estructura para manejo de errores avanzado
type ZyloTryCatch struct {
	Result     ZyloObject
	Error      ZyloObject
	StackTrace string
}

// NewZyloTryCatch crea una nueva estructura de manejo de errores
func NewZyloTryCatch() *ZyloTryCatch {
	return &ZyloTryCatch{}
}

// Execute ejecuta una función con manejo de errores
func (tc *ZyloTryCatch) Execute(fn func() ZyloObject) ZyloObject {
	defer func() {
		if r := recover(); r != nil {
			tc.Error = NewError("Runtime panic: %v", r)
			tc.StackTrace = "Panic recovered in runtime execution"
		}
	}()

	result := fn()
	tc.Result = result
	return result
}

// WasSuccessful indica si la ejecución fue exitosa
func (tc *ZyloTryCatch) WasSuccessful() bool {
	return tc.Error == nil
}

// GetError retorna el error si existe
func (tc *ZyloTryCatch) GetError() ZyloObject {
	return tc.Error
}

// GetStackTrace retorna el stack trace
func (tc *ZyloTryCatch) GetStackTrace() string {
	return tc.StackTrace
}

// GenerateStackTrace genera un stack trace detallado
func GenerateStackTrace() string {
	// En una implementación real, esto analizaría el stack de llamadas
	return "Stack trace: [runtime functions, caller functions,...]"
}

// ValidateArrayBounds valida límites de array con mensajes detallados
func ValidateArrayBounds(array *List, index int64) error {
	if index < 0 {
		return fmt.Errorf("array index cannot be negative: %d", index)
	}
	if index >= int64(len(array.Elements)) {
		return fmt.Errorf("array index out of bounds: %d (length: %d)", index, len(array.Elements))
	}
	return nil
}

// SafeArrayAccess accede a un array con validación robusta
func SafeArrayAccess(array *List, index int64) (ZyloObject, error) {
	if err := ValidateArrayBounds(array, index); err != nil {
		return NewError(err.Error()), err
	}
	return array.Elements[index], nil
}

// --- Time/Date Operations Avanzadas ---

// TimeModule representa operaciones temporales avanzadas
type TimeModule struct{}

// NewTimeModule crea un nuevo módulo de tiempo
func NewTimeModule() *TimeModule {
	return &TimeModule{}
}

// Now retorna la hora actual como timestamp
func (tm *TimeModule) Now() *Integer {
	return &Integer{Value: time.Now().Unix()}
}

// FormatTime formatea una hora unix como string
func (tm *TimeModule) FormatTime(timestamp int64, format string) *String {
	t := time.Unix(timestamp, 0)
	formatted := t.Format(format)
	return &String{Value: formatted}
}

// ParseTime parsea una fecha desde string
func (tm *TimeModule) ParseTime(dateStr, layout string) (*Integer, error) {
	parsed, err := time.Parse(layout, dateStr)
	if err != nil {
		return &Integer{Value: 0}, err
	}
	return &Integer{Value: parsed.Unix()}, nil
}

// DurationModule representa operaciones de duración
type DurationModule struct{}

// NewDurationModule crea un módulo de duración
func NewDurationModule() *DurationModule {
	return &DurationModule{}
}

// HoursToSeconds convierte horas a segundos
func (dm *DurationModule) HoursToSeconds(hours float64) *Integer {
	return &Integer{Value: int64(hours * 3600)}
}

// DaysToHours convierte días a horas
func (dm *DurationModule) DaysToHours(days float64) *Float {
	return &Float{Value: days * 24}
}

// --- RegExp Operations ---

// RegExpModule representa operaciones de expresiones regulares
type RegExpModule struct{}

// NewRegExpModule crea un módulo de regexp
func NewRegExpModule() *RegExpModule {
	return &RegExpModule{}
}

// Match verifica si un texto coincide con un patrón regex
func (rem *RegExpModule) Match(pattern, text string) *Bool {
	// Implementación básica de regex matching (simplificada)
	// En producción usaría regexp.Compile

	// Soporte básico para patrones simples
	if pattern == ".*" {
		return &Bool{Value: true}
	}

	if len(pattern) == 0 {
		return &Bool{Value: len(text) == 0}
	}

	// Exact match simple
	return &Bool{Value: pattern == text}
}

// Replace reemplaza ocurrencias usando patrones
func (rem *RegExpModule) Replace(pattern, replacement, text string) *String {
	// Implementación básica de reemplazo (simplificada)
	// En producción usaría regexp.Compile

	if pattern == ".*" {
		return &String{Value: replacement}
	}

	// Simple exact match replacement
	result := strings.ReplaceAll(text, pattern, replacement)
	return &String{Value: result}
}

// FindAll encuentra todas las coincidencias de un patrón
func (rem *RegExpModule) FindAll(pattern, text string) *List {
	// Implementación básica
	results := []ZyloObject{}

	if pattern == ".*" {
		results = append(results, &String{Value: text})
	} else if strings.Contains(text, pattern) {
		// Simple substring matching
		start := 0
		for {
			pos := strings.Index(text[start:], pattern)
			if pos == -1 {
				break
			}
			results = append(results, &String{Value: text[start:start+pos+len(pattern)]})
			start += pos + len(pattern)
		}
	}

	return &List{Elements: results}
}

// --- Integration Functions ---

// builtinTryCatch implementa try-catch en runtime
func builtinTryCatch(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("try_catch expects 1 argument (function), got %d", len(args))
	}

	// fn, ok := args[0].(*FunctionObject) // Asumiendo que tenemos objetos Function
	// if !ok {
	//     return NewError("try_catch expects a function argument")
	// }

	// tc := NewZyloTryCatch()
	// result := tc.Execute(func() ZyloObject {
	//     return fn.Call([]ZyloObject{})
	// })

	// Para esta implementación básica, solo imitamos manejo de errores
	return NewError("try_catch not fully implemented - requires function objects")
}

// builtinAutoCast convierte tipos automáticamente
func builtinAutoCast(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("auto_cast expects 2 arguments, got %d", len(args))
	}

	left, right := args[0], args[1]
	newLeft, newRight, changed := AdvancedAutoCast(left, right, "")

	if changed {
		// Retornar tuple-like con objetos convertidos
		result := &Map{Pairs: map[string]ZyloObject{}}
		result.Pairs["left_casted"] = newLeft
		result.Pairs["right_casted"] = newRight
		result.Pairs["changed"] = &Bool{Value: changed}
		return result
	}

	return &Bool{Value: false}
}

func builtinToNumber(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("ToNumber expects 1 argument, got %d", len(args))
	}
	switch o := args[0].(type) {
	case *String:
		if val, err := strconv.ParseFloat(o.Value, 64); err == nil {
			return &Float{Value: val}
		}
		return NewError("could not convert string '%s' to number", o.Value)
	case *Integer:
		return &Float{Value: float64(o.Value)}
	case *Float:
		return o
	default:
		return NewError("ToNumber expects a string, integer or float, got %s", args[0].Type())
	}
}

func builtinToInt(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("ToInt expects 1 argument, got %d", len(args))
	}
	switch o := args[0].(type) {
	case *Float:
		return &Integer{Value: int64(o.Value)}
	case *String:
		if val, err := strconv.ParseInt(o.Value, 10, 64); err == nil {
			return &Integer{Value: val}
		}
		return NewError("could not convert string '%s' to int", o.Value)
	case *Integer:
		return o
	default:
		return NewError("ToInt expects a string, integer or float, got %s", args[0].Type())
	}
}

func builtinToBool(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("ToBool expects 1 argument, got %d", len(args))
	}
	switch o := args[0].(type) {
	case *Bool:
		return o
	case *Null:
		return &Bool{Value: false}
	case *Integer:
		return &Bool{Value: o.Value != 0}
	case *Float:
		return &Bool{Value: o.Value != 0.0}
	case *String:
		return &Bool{Value: o.Value != ""}
	case *List:
		return &Bool{Value: len(o.Elements) > 0}
	case *Map:
		return &Bool{Value: len(o.Pairs) > 0}
	default:
		return &Bool{Value: true}
	}
}

func builtinStringList(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("string_list expects 1 argument, got %d", len(args))
	}
	return &String{Value: inspectValue(args[0])}
}

func builtinStringMap(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("string_map expects 1 argument, got %d", len(args))
	}
	return &String{Value: inspectValue(args[0])}
}

// --- Operaciones Matemáticas ---

func builtinAdd(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Add expects 2 arguments, got %d", len(args))
	}
	a, b := args[0], args[1]
	
	if af, ok := a.(*Float); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: af.Value + bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			return &Float{Value: af.Value + float64(bi.Value)}
		}
	}
	if ai, ok := a.(*Integer); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: float64(ai.Value) + bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			return &Integer{Value: ai.Value + bi.Value}
		}
	}
	if as, ok := a.(*String); ok {
		if bs, ok := b.(*String); ok {
			return &String{Value: as.Value + bs.Value}
		}
	}
	return NewError("unsupported types for +: %s and %s", a.Type(), b.Type())
}

func builtinSubtract(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Subtract expects 2 arguments, got %d", len(args))
	}
	a, b := args[0], args[1]
	
	if af, ok := a.(*Float); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: af.Value - bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			return &Float{Value: af.Value - float64(bi.Value)}
		}
	}
	if ai, ok := a.(*Integer); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: float64(ai.Value) - bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			return &Integer{Value: ai.Value - bi.Value}
		}
	}
	return NewError("unsupported types for -: %s and %s", a.Type(), b.Type())
}

func builtinMultiply(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Multiply expects 2 arguments, got %d", len(args))
	}
	a, b := args[0], args[1]
	
	if af, ok := a.(*Float); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: af.Value * bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			return &Float{Value: af.Value * float64(bi.Value)}
		}
	}
	if ai, ok := a.(*Integer); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: float64(ai.Value) * bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			return &Integer{Value: ai.Value * bi.Value}
		}
	}
	return NewError("unsupported types for *: %s and %s", a.Type(), b.Type())
}

func builtinDivide(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Divide expects 2 arguments, got %d", len(args))
	}
	a, b := args[0], args[1]
	
	if af, ok := a.(*Float); ok {
		if bf, ok := b.(*Float); ok {
			if bf.Value == 0 {
				return NewError("division by zero")
			}
			return &Float{Value: af.Value / bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			if bi.Value == 0 {
				return NewError("division by zero")
			}
			return &Float{Value: af.Value / float64(bi.Value)}
		}
	}
	if ai, ok := a.(*Integer); ok {
		if bf, ok := b.(*Float); ok {
			if bf.Value == 0 {
				return NewError("division by zero")
			}
			return &Float{Value: float64(ai.Value) / bf.Value}
		}
		if bi, ok := b.(*Integer); ok {
			if bi.Value == 0 {
				return NewError("division by zero")
			}
			return &Float{Value: float64(ai.Value) / float64(bi.Value)}
		}
	}
	return NewError("unsupported types for /: %s and %s", a.Type(), b.Type())
}

func builtinPower(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Power expects 2 arguments, got %d", len(args))
	}
	
	var base, exp float64
	
	switch b := args[0].(type) {
	case *Float:
		base = b.Value
	case *Integer:
		base = float64(b.Value)
	default:
		return NewError("Power expects numeric arguments, got %s", args[0].Type())
	}
	
	switch e := args[1].(type) {
	case *Float:
		exp = e.Value
	case *Integer:
		exp = float64(e.Value)
	default:
		return NewError("Power expects numeric arguments, got %s", args[1].Type())
	}
	
	return &Float{Value: math.Pow(base, exp)}
}

func builtinSqrt(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("Sqrt expects 1 argument, got %d", len(args))
	}
	
	var num float64
	switch n := args[0].(type) {
	case *Float:
		num = n.Value
	case *Integer:
		num = float64(n.Value)
	default:
		return NewError("Sqrt expects a numeric argument, got %s", args[0].Type())
	}
	
	if num < 0 {
		return NewError("Sqrt of negative number")
	}
	
	return &Float{Value: math.Sqrt(num)}
}

func builtinAbs(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("Abs expects 1 argument, got %d", len(args))
	}
	
	switch n := args[0].(type) {
	case *Float:
		return &Float{Value: math.Abs(n.Value)}
	case *Integer:
		if n.Value < 0 {
			return &Integer{Value: -n.Value}
		}
		return n
	default:
		return NewError("Abs expects a numeric argument, got %s", args[0].Type())
	}
}

func builtinRound(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("Round expects 1 argument, got %d", len(args))
	}
	
	var num float64
	switch n := args[0].(type) {
	case *Float:
		num = n.Value
	case *Integer:
		return n
	default:
		return NewError("Round expects a numeric argument, got %s", args[0].Type())
	}
	
	return &Float{Value: math.Round(num)}
}

func builtinMin(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Min expects 2 arguments, got %d", len(args))
	}
	a, b := args[0], args[1]
	
	if af, ok := a.(*Float); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: math.Min(af.Value, bf.Value)}
		}
		if bi, ok := b.(*Integer); ok {
			return &Float{Value: math.Min(af.Value, float64(bi.Value))}
		}
	}
	if ai, ok := a.(*Integer); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: math.Min(float64(ai.Value), bf.Value)}
		}
		if bi, ok := b.(*Integer); ok {
			if ai.Value < bi.Value {
				return ai
			}
			return bi
		}
	}
	return NewError("unsupported types for Min: %s and %s", a.Type(), b.Type())
}

func builtinMax(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("Max expects 2 arguments, got %d", len(args))
	}
	a, b := args[0], args[1]
	
	if af, ok := a.(*Float); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: math.Max(af.Value, bf.Value)}
		}
		if bi, ok := b.(*Integer); ok {
			return &Float{Value: math.Max(af.Value, float64(bi.Value))}
		}
	}
	if ai, ok := a.(*Integer); ok {
		if bf, ok := b.(*Float); ok {
			return &Float{Value: math.Max(float64(ai.Value), bf.Value)}
		}
		if bi, ok := b.(*Integer); ok {
			if ai.Value > bi.Value {
				return ai
			}
			return bi
		}
	}
	return NewError("unsupported types for Max: %s and %s", a.Type(), b.Type())
}

// --- Operaciones de I/O ---

func builtinReadLine(args ...ZyloObject) ZyloObject {
	if len(args) != 0 {
		return NewError("ReadLine takes no arguments, got %d", len(args))
	}
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return NewError("error reading input: %s", err.Error())
	}
	return &String{Value: strings.TrimSpace(text)}
}

func builtinReadFile(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("ReadFile expects 1 argument, got %d", len(args))
	}
	p, ok := args[0].(*String)
	if !ok {
		return NewError("ReadFile expects a string path, got %s", args[0].Type())
	}
	content, err := ioutil.ReadFile(p.Value)
	if err != nil {
		return NewError("could not read file: %s", err)
	}
	return &String{Value: string(content)}
}

func builtinWriteFile(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("WriteFile expects 2 arguments, got %d", len(args))
	}
	p, ok1 := args[0].(*String)
	c, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("WriteFile expects two strings")
	}
	err := ioutil.WriteFile(p.Value, []byte(c.Value), 0644)
	if err != nil {
		return NewError("could not write file: %s", err)
	}
	return &Bool{Value: true}
}

// --- Funciones de Utilidad ---

func builtinTypeOf(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("TypeOf expects 1 argument, got %d", len(args))
	}
	return &String{Value: string(args[0].Type())}
}

func builtinIsNull(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("IsNull expects 1 argument, got %d", len(args))
	}
	return &Bool{Value: args[0].Type() == NULL_OBJ}
}

func builtinIsEmpty(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("IsEmpty expects 1 argument, got %d", len(args))
	}
	switch o := args[0].(type) {
	case *String:
		return &Bool{Value: len(o.Value) == 0}
	case *List:
		return &Bool{Value: len(o.Elements) == 0}
	case *Map:
		return &Bool{Value: len(o.Pairs) == 0}
	default:
		return &Bool{Value: false}
	}
}

// --- Funciones de Soporte para el Compilador ---

func Println(args ...ZyloObject) {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.Inspect())
	}
	fmt.Println()
}

// builtinPrint para uso interno
func builtinPrint(args ...ZyloObject) ZyloObject {
	Println(args...)
	return NewNull()
}

// --- Funciones Asíncronas ---

func builtinSpawn(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("Spawn expects 1 argument (function), got %d", len(args))
	}
	// Esta es una implementación básica
	// En una implementación real, necesitarías ejecutar la función en una goroutine
	return NewError("Spawn not fully implemented - requires function execution context")
}

func builtinAwait(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("Await expects 1 argument, got %d", len(args))
	}
	future, ok := args[0].(*Future)
	if !ok {
		return NewError("Await expects a future object, got %s", args[0].Type())
	}
	if !future.once {
		future.value = <-future.Result
		future.once = true
	}
	return future.value
}

// --- Registro de Funciones Built-in ---

// GetBuiltins devuelve un mapa con todas las funciones built-in disponibles
func GetBuiltins() map[string]*Builtin {
	return map[string]*
	Builtin{
		// Operaciones de String
		"split":     NewBuiltin(builtinSplit),
		"join":      NewBuiltin(builtinJoin),
		"substring": NewBuiltin(builtinSubstring),
		"replace":   NewBuiltin(builtinReplace),
		"trim":      NewBuiltin(builtinTrim),

		// Operaciones de Lista
		"len":     NewBuiltin(builtinLen),
		"append":  NewBuiltin(builtinAppend),
		"prepend": NewBuiltin(builtinPrepend),
		"slice":   NewBuiltin(builtinSlice),
		"sort":    NewBuiltin(builtinSort),
		"reverse": NewBuiltin(builtinReverse),

		// Conversiones de Tipo
		"ToString": NewBuiltin(builtinToString),
		"ToNumber": NewBuiltin(builtinToNumber),
		"ToInt":    NewBuiltin(builtinToInt),
		"ToBool":   NewBuiltin(builtinToBool),

		// Operaciones Matemáticas
		"Add":      NewBuiltin(builtinAdd),
		"Subtract": NewBuiltin(builtinSubtract),
		"Multiply": NewBuiltin(builtinMultiply),
		"Divide":   NewBuiltin(builtinDivide),
		"Power":    NewBuiltin(builtinPower),
		"Sqrt":     NewBuiltin(builtinSqrt),
		"Abs":      NewBuiltin(builtinAbs),
		"Round":    NewBuiltin(builtinRound),
		"Min":      NewBuiltin(builtinMin),
		"Max":      NewBuiltin(builtinMax),

		// Operaciones de I/O
		"ReadLine":  NewBuiltin(builtinReadLine),
		"ReadFile":  NewBuiltin(builtinReadFile),
		"WriteFile": NewBuiltin(builtinWriteFile),

		// Funciones de Utilidad
		"TypeOf":  NewBuiltin(builtinTypeOf),
		"IsNull":  NewBuiltin(builtinIsNull),
		"IsEmpty": NewBuiltin(builtinIsEmpty),
		"print":   NewBuiltin(builtinPrint),

		// Funciones Asíncronas
		"Spawn": NewBuiltin(builtinSpawn),
		"Await": NewBuiltin(builtinAwait),
	}
}

// --- Funciones auxiliares para operaciones ---

// IsTruthy determina si un objeto se evalúa como verdadero
func IsTruthy(obj ZyloObject) bool {
	switch obj := obj.(type) {
	case *Null:
		return false
	case *Bool:
		return obj.Value
	case *Integer:
		return obj.Value != 0
	case *Float:
		return obj.Value != 0.0
	case *String:
		return obj.Value != ""
	case *List:
		return len(obj.Elements) > 0
	case *Map:
		return len(obj.Pairs) > 0
	default:
		return true
	}
}

// IsError verifica si un objeto es un error
func IsError(obj ZyloObject) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

// CompareObjects compara dos objetos y devuelve si son iguales
func CompareObjects(left, right ZyloObject) bool {
	if left.Type() != right.Type() {
		return false
	}

	switch left := left.(type) {
	case *Integer:
		return left.Value == right.(*Integer).Value
	case *Float:
		return left.Value == right.(*Float).Value
	case *String:
		return left.Value == right.(*String).Value
	case *Bool:
		return left.Value == right.(*Bool).Value
	case *Null:
		return true
	default:
		return false
	}
}

// NativeBoolToBooleanObject convierte un bool nativo a un objeto Bool
func NativeBoolToBooleanObject(input bool) *Bool {
	if input {
		return &Bool{Value: true}
	}
	return &Bool{Value: false}
}

// --- Funciones adicionales para Map ---

func builtinMapGet(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("map_get expects 2 arguments, got %d", len(args))
	}
	m, ok1 := args[0].(*Map)
	key, ok2 := args[1].(*String)
	if !ok1 {
		return NewError("map_get expects a map as first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("map_get expects a string as key, got %s", args[1].Type())
	}
	
	if val, exists := m.Pairs[key.Value]; exists {
		return val
	}
	return NewNull()
}

func builtinMapSet(args ...ZyloObject) ZyloObject {
	if len(args) != 3 {
		return NewError("map_set expects 3 arguments, got %d", len(args))
	}
	m, ok1 := args[0].(*Map)
	key, ok2 := args[1].(*String)
	if !ok1 {
		return NewError("map_set expects a map as first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("map_set expects a string as key, got %s", args[1].Type())
	}
	
	// Crear un nuevo mapa para inmutabilidad
	newPairs := make(map[string]ZyloObject)
	for k, v := range m.Pairs {
		newPairs[k] = v
	}
	newPairs[key.Value] = args[2]
	
	return &Map{Pairs: newPairs}
}

func builtinMapHas(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("map_has expects 2 arguments, got %d", len(args))
	}
	m, ok1 := args[0].(*Map)
	key, ok2 := args[1].(*String)
	if !ok1 {
		return NewError("map_has expects a map as first argument, got %s", args[0].Type())
	}
	if !ok2 {
		return NewError("map_has expects a string as key, got %s", args[1].Type())
	}
	
	_, exists := m.Pairs[key.Value]
	return &Bool{Value: exists}
}

func builtinMapKeys(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("map_keys expects 1 argument, got %d", len(args))
	}
	m, ok := args[0].(*Map)
	if !ok {
		return NewError("map_keys expects a map, got %s", args[0].Type())
	}
	
	keys := make([]ZyloObject, 0, len(m.Pairs))
	for k := range m.Pairs {
		keys = append(keys, &String{Value: k})
	}
	return &List{Elements: keys}
}

func builtinMapValues(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("map_values expects 1 argument, got %d", len(args))
	}
	m, ok := args[0].(*Map)
	if !ok {
		return NewError("map_values expects a map, got %s", args[0].Type())
	}
	
	values := make([]ZyloObject, 0, len(m.Pairs))
	for _, v := range m.Pairs {
		values = append(values, v)
	}
	return &List{Elements: values}
}

// --- Funciones matemáticas adicionales ---

func builtinFloor(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("floor expects 1 argument, got %d", len(args))
	}
	
	var num float64
	switch n := args[0].(type) {
	case *Float:
		num = n.Value
	case *Integer:
		return n
	default:
		return NewError("floor expects a numeric argument, got %s", args[0].Type())
	}
	
	return &Float{Value: math.Floor(num)}
}

func builtinCeil(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("ceil expects 1 argument, got %d", len(args))
	}
	
	var num float64
	switch n := args[0].(type) {
	case *Float:
		num = n.Value
	case *Integer:
		return n
	default:
		return NewError("ceil expects a numeric argument, got %s", args[0].Type())
	}
	
	return &Float{Value: math.Ceil(num)}
}

func builtinMod(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("mod expects 2 arguments, got %d", len(args))
	}
	
	a, b := args[0], args[1]
	
	if ai, ok := a.(*Integer); ok {
		if bi, ok := b.(*Integer); ok {
			if bi.Value == 0 {
				return NewError("modulo by zero")
			}
			return &Integer{Value: ai.Value % bi.Value}
		}
	}
	
	return NewError("mod expects two integers, got %s and %s", a.Type(), b.Type())
}

// --- Funciones de string adicionales ---

func builtinToUpper(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("to_upper expects 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*String)
	if !ok {
		return NewError("to_upper expects a string, got %s", args[0].Type())
	}
	return &String{Value: strings.ToUpper(s.Value)}
}

func builtinToLower(args ...ZyloObject) ZyloObject {
	if len(args) != 1 {
		return NewError("to_lower expects 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*String)
	if !ok {
		return NewError("to_lower expects a string, got %s", args[0].Type())
	}
	return &String{Value: strings.ToLower(s.Value)}
}

func builtinContains(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("contains expects 2 arguments, got %d", len(args))
	}
	s, ok1 := args[0].(*String)
	substr, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("contains expects two strings")
	}
	return &Bool{Value: strings.Contains(s.Value, substr.Value)}
}

func builtinStartsWith(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("starts_with expects 2 arguments, got %d", len(args))
	}
	s, ok1 := args[0].(*String)
	prefix, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("starts_with expects two strings")
	}
	return &Bool{Value: strings.HasPrefix(s.Value, prefix.Value)}
}

func builtinEndsWith(args ...ZyloObject) ZyloObject {
	if len(args) != 2 {
		return NewError("ends_with expects 2 arguments, got %d", len(args))
	}
	s, ok1 := args[0].(*String)
	suffix, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("ends_with expects two strings")
	}
	return &Bool{Value: strings.HasSuffix(s.Value, suffix.Value)}
}

// --- Funciones para el Code Generator ---

// ToInt convierte un valor interface{} a un objeto Integer
func ToInt(value interface{}) interface{} {
	switch v := value.(type) {
	case *Integer:
		return v
	case *Float:
		return &Integer{Value: int64(v.Value)}
	case *String:
		if n, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
			return &Integer{Value: n}
		}
		return &Integer{Value: 0}
	default:
		return &Integer{Value: 0}
	}
}

// ToFloat convierte un valor interface{} a un objeto Float
func ToFloat(value interface{}) interface{} {
	switch v := value.(type) {
	case *Float:
		return v
	case *Integer:
		return &Float{Value: float64(v.Value)}
	case *String:
		if f, err := strconv.ParseFloat(v.Value, 64); err == nil {
			return &Float{Value: f}
		}
		return &Float{Value: 0.0}
	default:
		return &Float{Value: 0.0}
	}
}

// ToString convierte un valor interface{} a un objeto String
func ToString(value interface{}) interface{} {
	switch v := value.(type) {
	case *String:
		return v
	case *Integer:
		return &String{Value: fmt.Sprintf("%d", v.Value)}
	case *Float:
		return &String{Value: fmt.Sprintf("%g", v.Value)}
	case *Bool:
		return &String{Value: fmt.Sprintf("%t", v.Value)}
	default:
		return &String{Value: "<unknown>"}
	}
}

// ToBool convierte un valor interface{} a un objeto Bool
func ToBool(value interface{}) interface{} {
	switch v := value.(type) {
	case *Bool:
		return v
	case *Null:
		return &Bool{Value: false}
	case *Integer:
		return &Bool{Value: v.Value != 0}
	case *Float:
		return &Bool{Value: v.Value != 0.0}
	case *String:
		return &Bool{Value: v.Value != ""}
	case *List:
		return &Bool{Value: len(v.Elements) > 0}
	case *Map:
		return &Bool{Value: len(v.Pairs) > 0}
	default:
		return &Bool{Value: false}
	}
}

// ToZyloObject convierte un valor interface{} a ZyloObject
func ToZyloObject(value interface{}) ZyloObject {
	if obj, ok := value.(ZyloObject); ok {
		return obj
	}
	// Si no es ZyloObject, convertir a string
	return &String{Value: fmt.Sprintf("%v", value)}
}

// --- Funciones de métodos de colección ---

// ListPush añade un elemento al final de la lista
func ListPush(list interface{}, element interface{}) interface{} {
	if l, ok := list.(*List); ok {
		newElements := make([]ZyloObject, len(l.Elements)+1)
		copy(newElements, l.Elements)
		newElements[len(l.Elements)] = ToZyloObject(element)
		return &List{Elements: newElements}
	}
	return NewError("ListPush expects a List as first argument")
}

// ListPop elimina y devuelve el último elemento de la lista
func ListPop(list interface{}) interface{} {
	if l, ok := list.(*List); ok {
		if len(l.Elements) == 0 {
			return NewNull()
		}
		lastIdx := len(l.Elements) - 1
		element := l.Elements[lastIdx]
		newElements := make([]ZyloObject, lastIdx)
		copy(newElements, l.Elements[:lastIdx])
		return element
	}
	return NewError("ListPop expects a List")
}

// ListShift elimina y devuelve el primer elemento de la lista
func ListShift(list interface{}) interface{} {
	if l, ok := list.(*List); ok {
		if len(l.Elements) == 0 {
			return NewNull()
		}
		element := l.Elements[0]
		newElements := make([]ZyloObject, len(l.Elements)-1)
		copy(newElements, l.Elements[1:])
		return element
	}
	return NewError("ListShift expects a List")
}

// ListUnshift añade un elemento al inicio de la lista
func ListUnshift(list interface{}, element interface{}) interface{} {
	if l, ok := list.(*List); ok {
		newElements := make([]ZyloObject, len(l.Elements)+1)
		copy(newElements[1:], l.Elements)
		newElements[0] = ToZyloObject(element)
		return &List{Elements: newElements}
	}
	return NewError("ListUnshift expects a List as first argument")
}

// ListIndexOf devuelve el índice del elemento dado
func ListIndexOf(list interface{}, element interface{}) interface{} {
	if l, ok := list.(*List); ok {
		elem := ToZyloObject(element)
		for i, e := range l.Elements {
			if CompareObjects(e, elem) {
				return &Integer{Value: int64(i)}
			}
		}
		return &Integer{Value: -1}
	}
	return NewError("ListIndexOf expects a List as first argument")
}

// ListIncludes verifica si la lista contiene un elemento
func ListIncludes(list interface{}, element interface{}) interface{} {
	idx := ListIndexOf(list, element)
	if i, ok := idx.(*Integer); ok {
		return &Bool{Value: i.Value >= 0}
	}
	return idx // Return error
}

// ListSlice devuelve una porción de la lista
func ListSlice(list interface{}, start interface{}, end interface{}) interface{} {
	if l, ok := list.(*List); ok {
		var startIdx, endIdx int64 = 0, int64(len(l.Elements))

		if s, ok := start.(*Integer); ok {
			startIdx = s.Value
		}
		if e, ok := end.(*Integer); ok {
			endIdx = e.Value
		}

		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx > int64(len(l.Elements)) {
			endIdx = int64(len(l.Elements))
		}
		if startIdx > endIdx {
			return &List{Elements: []ZyloObject{}}
		}

		newElements := make([]ZyloObject, endIdx-startIdx)
		copy(newElements, l.Elements[startIdx:endIdx])
		return &List{Elements: newElements}
	}
	return NewError("ListSlice expects a List as first argument")
}

// ListReverse invierte el orden de los elementos de la lista
func ListReverse(list interface{}) interface{} {
	if l, ok := list.(*List); ok {
		newElements := make([]ZyloObject, len(l.Elements))
		copy(newElements, l.Elements)
		for i, j := 0, len(newElements)-1; i < j; i, j = i+1, j-1 {
			newElements[i], newElements[j] = newElements[j], newElements[i]
		}
		return &List{Elements: newElements}
	}
	return NewError("ListReverse expects a List")
}

// ListSort ordena la lista (por representación string)
func ListSort(list interface{}) interface{} {
	if l, ok := list.(*List); ok {
		newElements := make([]ZyloObject, len(l.Elements))
		copy(newElements, l.Elements)
		sort.SliceStable(newElements, func(i, j int) bool {
			return newElements[i].Inspect() < newElements[j].Inspect()
		})
		return &List{Elements: newElements}
	}
	return NewError("ListSort expects a List")
}

// ListConcat concatena dos listas
func ListConcat(list1 interface{}, list2 interface{}) interface{} {
	var elements1, elements2 []ZyloObject

	if l1, ok := list1.(*List); ok {
		elements1 = l1.Elements
	} else {
		return NewError("ListConcat expects Lists as arguments")
	}

	if l2, ok := list2.(*List); ok {
		elements2 = l2.Elements
	} else {
		return NewError("ListConcat expects Lists as arguments")
	}

	newElements := make([]ZyloObject, len(elements1)+len(elements2))
	copy(newElements, elements1)
	copy(newElements[len(elements1):], elements2)
	return &List{Elements: newElements}
}

// --- Funciones de Map ---

// MapSet establece un valor en el mapa
func MapSet(m interface{}, key interface{}, value interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		keyStr := ""
		if k, ok := key.(*String); ok {
			keyStr = k.Value
		} else {
			return NewError("MapSet expects a String as key")
		}

		// Create a new map for immutability
		newPairs := make(map[string]ZyloObject)
		for k, v := range mapObj.Pairs {
			newPairs[k] = v
		}
		newPairs[keyStr] = ToZyloObject(value)
		return &Map{Pairs: newPairs}
	}
	return NewError("MapSet expects a Map as first argument")
}

// MapGet obtiene un valor del mapa
func MapGet(m interface{}, key interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		keyStr := ""
		if k, ok := key.(*String); ok {
			keyStr = k.Value
		} else {
			return NewError("MapGet expects a String as key")
		}

		if val, exists := mapObj.Pairs[keyStr]; exists {
			return val
		}
		return NewNull()
	}
	return NewError("MapGet expects a Map as first argument")
}

// MapHas verifica si el mapa contiene una clave
func MapHas(m interface{}, key interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		keyStr := ""
		if k, ok := key.(*String); ok {
			keyStr = k.Value
		} else {
			return NewError("MapHas expects a String as key")
		}

		_, exists := mapObj.Pairs[keyStr]
		return &Bool{Value: exists}
	}
	return NewError("MapHas expects a Map as first argument")
}

// MapDelete elimina una clave del mapa
func MapDelete(m interface{}, key interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		keyStr := ""
		if k, ok := key.(*String); ok {
			keyStr = k.Value
		} else {
			return NewError("MapDelete expects a String as key")
		}

		newPairs := make(map[string]ZyloObject)
		for k, v := range mapObj.Pairs {
			if k != keyStr {
				newPairs[k] = v
			}
		}
		return &Map{Pairs: newPairs}
	}
	return NewError("MapDelete expects a Map as first argument")
}

// MapClear elimina todos los elementos del mapa
func MapClear(m interface{}) interface{} {
	if _, ok := m.(*Map); ok {
		return &Map{Pairs: make(map[string]ZyloObject)}
	}
	return NewError("MapClear expects a Map")
}

// MapSize devuelve el número de elementos en el mapa
func MapSize(m interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		return &Integer{Value: int64(len(mapObj.Pairs))}
	}
	return NewError("MapSize expects a Map")
}

// MapKeys devuelve una lista con todas las claves del mapa
func MapKeys(m interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		keys := make([]ZyloObject, 0, len(mapObj.Pairs))
		for k := range mapObj.Pairs {
			keys = append(keys, &String{Value: k})
		}
		return &List{Elements: keys}
	}
	return NewError("MapKeys expects a Map")
}

// MapValues devuelve una lista con todos los valores del mapa
func MapValues(m interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		values := make([]ZyloObject, 0, len(mapObj.Pairs))
		for _, v := range mapObj.Pairs {
			values = append(values, v)
		}
		return &List{Elements: values}
	}
	return NewError("MapValues expects a Map")
}

// MapEntries devuelve una lista de pares [key, value] del mapa
func MapEntries(m interface{}) interface{} {
	if mapObj, ok := m.(*Map); ok {
		entries := make([]ZyloObject, 0, len(mapObj.Pairs))
		for k, v := range mapObj.Pairs {
			pair := []ZyloObject{&String{Value: k}, v}
			entries = append(entries, &List{Elements: pair})
		}
		return &List{Elements: entries}
	}
	return NewError("MapEntries expects a Map")
}

// --- Funciones públicas para operaciones binarias para el Compilador ---

// Add ejecuta una suma binaria entre dos ZyloObject
func Add(left, right ZyloObject) ZyloObject {
	return builtinAdd(left, right)
}

// Subtract ejecuta una resta binaria entre dos ZyloObject
func Subtract(left, right ZyloObject) ZyloObject {
	return builtinSubtract(left, right)
}

// Multiply ejecuta una multiplicación binaria entre dos ZyloObject
func Multiply(left, right ZyloObject) ZyloObject {
	return builtinMultiply(left, right)
}

// Divide ejecuta una división binaria entre dos ZyloObject
func Divide(left, right ZyloObject) ZyloObject {
	return builtinDivide(left, right)
}

// Equals compara si dos ZyloObject son iguales
func Equals(left, right ZyloObject) ZyloObject {
	if left.Type() != right.Type() {
		return &Bool{Value: false}
	}

	switch l := left.(type) {
	case *Integer:
		r := right.(*Integer)
		return &Bool{Value: l.Value == r.Value}
	case *Float:
		r := right.(*Float)
		return &Bool{Value: l.Value == r.Value}
	case *String:
		r := right.(*String)
		return &Bool{Value: l.Value == r.Value}
	case *Bool:
		r := right.(*Bool)
		return &Bool{Value: l.Value == r.Value}
	case *Null:
		return &Bool{Value: true}
	default:
		return &Bool{Value: false}
	}
}

// NotEquals compara si dos ZyloObject son diferentes
func NotEquals(left, right ZyloObject) ZyloObject {
	result := Equals(left, right)
	return &Bool{Value: !result.(*Bool).Value}
}

// Modulo calcula el módulo entre dos enteros
func Modulo(left, right ZyloObject) ZyloObject {
	if li, ok := left.(*Integer); ok {
		if ri, ok := right.(*Integer); ok {
			if ri.Value == 0 {
				return NewError("modulo by zero")
			}
			return &Integer{Value: li.Value % ri.Value}
		}
	}
	return NewError("modulo requires integer operands")
}

// GetExtendedBuiltins devuelve un mapa con TODAS las funciones built-in
func GetExtendedBuiltins() map[string]*Builtin {
	builtins := GetBuiltins()

	// Agregar funciones adicionales
	builtins["map_get"] = NewBuiltin(builtinMapGet)
	builtins["map_set"] = NewBuiltin(builtinMapSet)
	builtins["map_has"] = NewBuiltin(builtinMapHas)
	builtins["map_keys"] = NewBuiltin(builtinMapKeys)
	builtins["map_values"] = NewBuiltin(builtinMapValues)

	builtins["floor"] = NewBuiltin(builtinFloor)
	builtins["ceil"] = NewBuiltin(builtinCeil)
	builtins["mod"] = NewBuiltin(builtinMod)

	builtins["to_upper"] = NewBuiltin(builtinToUpper)
	builtins["to_lower"] = NewBuiltin(builtinToLower)
	builtins["contains"] = NewBuiltin(builtinContains)
	builtins["starts_with"] = NewBuiltin(builtinStartsWith)
	builtins["ends_with"] = NewBuiltin(builtinEndsWith)

	builtins["string_list"] = NewBuiltin(builtinStringList)
	builtins["string_map"] = NewBuiltin(builtinStringMap)

	return builtins
}
