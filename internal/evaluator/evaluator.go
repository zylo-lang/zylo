package evaluator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/zylo-lang/zylo/internal/ast"
)

// ZyloObject representa un objeto en tiempo de ejecución de Zylo
type ZyloObject interface {
	Type() string
	Inspect() string
}

// String representa un objeto string
type String struct {
	Value string
}

func pow(a, b float64) float64 {
    return math.Pow(a, b)
}

func (s *String) Type() string    { return "STRING_OBJ" }
func (s *String) Inspect() string { return s.Value }

// Integer representa un objeto integer
type Integer struct {
	Value int64
}

func (i *Integer) Type() string    { return "INTEGER_OBJ" }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Float representa un objeto float
type Float struct {
	Value float64
}

func (f *Float) Type() string    { return "FLOAT_OBJ" }
func (f *Float) Inspect() string { return fmt.Sprintf("%g", f.Value) }

// List representa un objeto list
type List struct {
	Items []Value
}

func (l *List) Type() string { return "LIST_OBJ" }
func (l *List) Inspect() string {
	parts := make([]string, len(l.Items))
	for i, el := range l.Items {
		if obj, ok := el.(ZyloObject); ok {
			parts[i] = obj.Inspect()
		} else {
			parts[i] = fmt.Sprintf("%v", el)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// MapObject representa un objeto map
type MapObject struct {
	Pairs map[string]Value
}

func (m *MapObject) Type() string { return "Map" }
func (m *MapObject) Inspect() string {
	var out strings.Builder
	out.WriteString("{")
	first := true
	for k, v := range m.Pairs {
		if !first {
			out.WriteString(", ")
		}
		out.WriteString(k + ": ")
		if obj, ok := v.(ZyloObject); ok {
			out.WriteString(obj.Inspect())
		} else {
			out.WriteString(fmt.Sprintf("%v", v))
		}
		first = false
	}
	out.WriteString("}")
	return out.String()
}

// Boolean representa un objeto boolean
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() string    { return "BOOLEAN_OBJ" }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Null representa un objeto null
type Null struct{}

func (n *Null) Type() string    { return "NULL_OBJ" }
func (n *Null) Inspect() string { return "null" }

// Value representa un valor en tiempo de ejecución de Zylo
type Value interface{}

// Future representa un resultado de una operación asíncrona
type Future struct {
	Result chan ZyloObject
	value  ZyloObject
	once   bool
}

func (f *Future) Type() string { return "FUTURE_OBJ" }
func (f *Future) Inspect() string { return "future" }

// Environment representa el entorno de ejecución con variables
type Environment struct {
	variables map[string]Value
	constants map[string]bool
	types     map[string]string // Variable name to type
	parent    *Environment
}

// NewEnvironment crea un nuevo entorno
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		types:     make(map[string]string),
		parent:    nil,
	}
}

// NewChildEnvironment crea un entorno hijo
func (e *Environment) NewChildEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		types:     make(map[string]string),
		parent:    e,
	}
}

// NewEnclosedEnvironment crea un entorno encerrado
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		variables: make(map[string]Value),
		constants: make(map[string]bool),
		types:     make(map[string]string),
		parent:    outer,
	}
}

// Get obtiene el valor de una variable
func (e *Environment) Get(name string) (Value, bool) {
	if value, exists := e.variables[name]; exists {
		return value, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set establece el valor de una variable
func (e *Environment) Set(name string, value Value) {
	e.variables[name] = value
}

// Update actualiza una variable existente
func (e *Environment) Update(name string, value Value) bool {
	if _, exists := e.variables[name]; exists {
		e.variables[name] = value
		return true
	}
	if e.parent != nil {
		return e.parent.Update(name, value)
	}
	return false
}

// IsConstant verifica si una variable es constante
func (e *Environment) IsConstant(name string) bool {
	if isConst, exists := e.constants[name]; exists {
		return isConst
	}
	if e.parent != nil {
		return e.parent.IsConstant(name)
	}
	return false
}

// GetType obtiene el tipo de una variable
func (e *Environment) GetType(name string) (string, bool) {
	if typ, exists := e.types[name]; exists {
		return typ, true
	}
	if e.parent != nil {
		return e.parent.GetType(name)
	}
	return "", false
}

// SetType establece el tipo de una variable
func (e *Environment) SetType(name string, typ string) {
	e.types[name] = typ
}

// Evaluator evalúa expresiones y sentencias de Zylo
type Evaluator struct {
	env            *Environment
	reader         *bufio.Reader
	callDepth      int
	evaluateDepth  int
	httpHandler    *ZyloFunction
	httpServer     *http.Server
}

// EvaluateProgram evalúa un programa completo
func (e *Evaluator) EvaluateProgram(program *ast.Program) error {
	for _, stmt := range program.Statements {
		_, err := e.evaluateStatement(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewEvaluator crea un nuevo evaluador
func NewEvaluator() *Evaluator {
	eval := &Evaluator{
		env:            NewEnvironment(),
		reader:         bufio.NewReader(os.Stdin),
		callDepth:      0,
		evaluateDepth:  0,
		httpHandler:    nil,
		httpServer:     nil,
	}
	eval.InitBuiltins()
	return eval
}

// InitBuiltins inicializa las funciones incorporadas
func (e *Evaluator) InitBuiltins() {
	// Constantes globales
	e.env.Set("null", &Null{})
	e.env.Set("true", &Boolean{Value: true})
	e.env.Set("false", &Boolean{Value: false})

	// show.log
	e.env.Set("show.log", &BuiltinFunction{
		Name: "show.log",
		Fn: func(args []Value) (Value, error) {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				if obj, ok := arg.(ZyloObject); ok {
					fmt.Print(obj.Inspect())
				} else {
					fmt.Print(arg)
				}
			}
			fmt.Println()
			os.Stdout.Sync()
			return &Null{}, nil
		},
	})

	// show.error
	e.env.Set("show.error", &BuiltinFunction{
		Name: "show.error",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("show.error expects 1 argument")
			}
			if str, ok := args[0].(*String); ok {
				// Error message is already formatted, print it directly
				fmt.Println(str.Value)
				os.Stdout.Sync()
				return &Null{}, nil
			}
			return nil, fmt.Errorf("show.error expects a string argument")
		},
	})

	// read.line
	e.env.Set("read.line", &BuiltinFunction{
		Name: "read.line",
		Fn: func(args []Value) (Value, error) {
			fmt.Print("> ")
			os.Stdout.Sync()
			input, _ := e.reader.ReadString('\n')
			return &String{Value: strings.TrimSpace(input)}, nil
		},
	})

	// read.int
	e.env.Set("read.int", &BuiltinFunction{
		Name: "read.int",
		Fn: func(args []Value) (Value, error) {
			for {
				fmt.Print("> ")
				os.Stdout.Sync()
				input, _ := e.reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if n, err := strconv.Atoi(input); err == nil {
					return &Integer{Value: int64(n)}, nil
				}
				fmt.Println("Error: no es un número válido")
			}
		},
	})

	// string() - Convierte a string
	e.env.Set("string", &BuiltinFunction{
		Name: "string",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("string() espera 1 argumento")
			}
			switch arg := args[0].(type) {
			case *Integer:
				return &String{Value: fmt.Sprintf("%d", arg.Value)}, nil
			case *Float:
				return &String{Value: fmt.Sprintf("%g", arg.Value)}, nil
			case *String:
				return arg, nil
			case *Boolean:
				return &String{Value: fmt.Sprintf("%t", arg.Value)}, nil
			default:
				if obj, ok := arg.(ZyloObject); ok {
					return &String{Value: obj.Inspect()}, nil
				}
				return &String{Value: fmt.Sprintf("%v", arg)}, nil
			}
		},
	})

	// string_list() - Convierte una lista a string legible
	e.env.Set("string_list", &BuiltinFunction{
		Name: "string_list",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("string_list() espera 1 argumento")
			}
			if obj, ok := args[0].(ZyloObject); ok {
				return &String{Value: obj.Inspect()}, nil
			}
			return &String{Value: fmt.Sprintf("%v", args[0])}, nil
		},
	})

	// string_map() - Convierte un mapa a string legible
	e.env.Set("string_map", &BuiltinFunction{
		Name: "string_map",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("string_map() espera 1 argumento")
			}
			if obj, ok := args[0].(ZyloObject); ok {
				return &String{Value: obj.Inspect()}, nil
			}
			return &String{Value: fmt.Sprintf("%v", args[0])}, nil
		},
	})

	// map_keys() - Retorna una lista con las claves del mapa
	e.env.Set("map_keys", &BuiltinFunction{
		Name: "map_keys",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("map_keys() espera 1 argumento")
			}
			if m, ok := args[0].(*MapObject); ok {
				keys := make([]Value, 0, len(m.Pairs))
				for k := range m.Pairs {
					keys = append(keys, &String{Value: k})
				}
				return &List{Items: keys}, nil
			}
			return nil, fmt.Errorf("map_keys() espera un mapa")
		},
	})

	// map_values() - Retorna una lista con los valores del mapa
	e.env.Set("map_values", &BuiltinFunction{
		Name: "map_values",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("map_values() espera 1 argumento")
			}
			if m, ok := args[0].(*MapObject); ok {
				values := make([]Value, 0, len(m.Pairs))
				for _, v := range m.Pairs {
					values = append(values, v)
				}
				return &List{Items: values}, nil
			}
			return nil, fmt.Errorf("map_values() espera un mapa")
		},
	})

	// int() - Convierte a entero
	e.env.Set("int", &BuiltinFunction{
		Name: "int",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("int() espera 1 argumento")
			}
			switch arg := args[0].(type) {
			case *String:
				if n, err := strconv.ParseInt(arg.Value, 10, 64); err == nil {
					return &Integer{Value: n}, nil
				}
				return nil, fmt.Errorf("no se puede convertir '%s' a int", arg.Value)
			case *Integer:
				return arg, nil
			case *Float:
				return &Integer{Value: int64(arg.Value)}, nil
			case *Boolean:
				if arg.Value {
					return &Integer{Value: 1}, nil
				}
				return &Integer{Value: 0}, nil
			default:
				return nil, fmt.Errorf("int() no soportado para %T", arg)
			}
		},
	})

	// float() - Convierte a float
	e.env.Set("float", &BuiltinFunction{
		Name: "float",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("float() espera 1 argumento")
			}
			switch arg := args[0].(type) {
			case *String:
				if f, err := strconv.ParseFloat(arg.Value, 64); err == nil {
					return &Float{Value: f}, nil
				}
				return nil, fmt.Errorf("no se puede convertir '%s' a float", arg.Value)
			case *Integer:
				return &Float{Value: float64(arg.Value)}, nil
			case *Float:
				return arg, nil
			default:
				return nil, fmt.Errorf("float() no soportado para %T", arg)
			}
		},
	})

	// bool() - Convierte a booleano
	e.env.Set("bool", &BuiltinFunction{
		Name: "bool",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("bool() espera 1 argumento")
			}
			return &Boolean{Value: e.isTruthy(args[0])}, nil
		},
	})

	// len()
	e.env.Set("len", &BuiltinFunction{
		Name: "len",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len() espera 1 argumento")
			}
			switch arg := args[0].(type) {
			case *List:
				return &Integer{Value: int64(len(arg.Items))}, nil
			case *String:
				return &Integer{Value: int64(len(arg.Value))}, nil
			default:
				return nil, fmt.Errorf("len() no soportado para %T", arg)
			}
		},
	})

	// ReadLine - Alias de read.line
	e.env.Set("ReadLine", &BuiltinFunction{
		Name: "ReadLine",
		Fn: func(args []Value) (Value, error) {
			input, _ := e.reader.ReadString('\n')
			return &String{Value: strings.TrimSpace(input)}, nil
		},
	})

	// ToNumber - Convierte string a número
	e.env.Set("ToNumber", &BuiltinFunction{
		Name: "ToNumber",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ToNumber() espera 1 argumento")
			}
			switch arg := args[0].(type) {
			case *String:
				if n, err := strconv.ParseInt(arg.Value, 10, 64); err == nil {
					return &Integer{Value: n}, nil
				}
				if f, err := strconv.ParseFloat(arg.Value, 64); err == nil {
					return &Float{Value: f}, nil
				}
				return &String{Value: "ERROR"}, nil
			case *Integer:
				return arg, nil
			case *Float:
				return arg, nil
			default:
				return &String{Value: "ERROR"}, nil
			}
		},
	})

	// ToInt - Convierte a entero
	e.env.Set("ToInt", &BuiltinFunction{
		Name: "ToInt",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ToInt() espera 1 argumento")
			}
			switch arg := args[0].(type) {
			case *String:
				if n, err := strconv.ParseInt(arg.Value, 10, 64); err == nil {
					return &Integer{Value: n}, nil
				}
				return &Integer{Value: 0}, nil
			case *Integer:
				return arg, nil
			case *Float:
				return &Integer{Value: int64(arg.Value)}, nil
			default:
				return &Integer{Value: 0}, nil
			}
		},
	})

	// ToBool - Convierte a booleano
	e.env.Set("ToBool", &BuiltinFunction{
		Name: "ToBool",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ToBool() espera 1 argumento")
			}
			return &Boolean{Value: e.isTruthy(args[0])}, nil
		},
	})

	// TypeOf - Retorna el tipo del valor
	e.env.Set("TypeOf", &BuiltinFunction{
		Name: "TypeOf",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeOf() espera 1 argumento")
			}
			var typeName string
			switch args[0].(type) {
			case *Integer:
				typeName = "INTEGER"
			case *Float:
				typeName = "FLOAT"
			case *String:
				if s, ok := args[0].(*String); ok && s.Value == "ERROR" {
					typeName = "ERROR"
				} else {
					typeName = "STRING"
				}
			case *Boolean:
				typeName = "BOOLEAN"
			case *Null:
				typeName = "NULL"
			case *List:
				typeName = "LIST"
			case *MapObject:
				typeName = "MAP"
			default:
				typeName = "UNKNOWN"
			}
			return &String{Value: typeName}, nil
		},
	})

	// ToString - Convierte cualquier valor a string
	e.env.Set("ToString", &BuiltinFunction{
		Name: "ToString",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ToString() espera 1 argumento")
			}
			if obj, ok := args[0].(ZyloObject); ok {
				return &String{Value: obj.Inspect()}, nil
			}
			return &String{Value: fmt.Sprintf("%v", args[0])}, nil
		},
	})

	// Add - Suma dos valores
	e.env.Set("Add", &BuiltinFunction{
		Name: "Add",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Add() espera 2 argumentos")
			}
			left, right := args[0], args[1]

			if l, ok := left.(*Integer); ok {
				if r, ok := right.(*Integer); ok {
					return &Integer{Value: l.Value + r.Value}, nil
				}
				if r, ok := right.(*Float); ok {
					return &Float{Value: float64(l.Value) + r.Value}, nil
				}
			}

			if l, ok := left.(*Float); ok {
				if r, ok := right.(*Float); ok {
					return &Float{Value: l.Value + r.Value}, nil
				}
				if r, ok := right.(*Integer); ok {
					return &Float{Value: l.Value + float64(r.Value)}, nil
				}
			}

			return nil, fmt.Errorf("Add: tipos incompatibles")
		},
	})

	// Subtract - Resta dos valores
	e.env.Set("Subtract", &BuiltinFunction{
		Name: "Subtract",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Subtract() espera 2 argumentos")
			}
			left, right := args[0], args[1]

			if l, ok := left.(*Integer); ok {
				if r, ok := right.(*Integer); ok {
					return &Integer{Value: l.Value - r.Value}, nil
				}
				if r, ok := right.(*Float); ok {
					return &Float{Value: float64(l.Value) - r.Value}, nil
				}
			}

			if l, ok := left.(*Float); ok {
				if r, ok := right.(*Float); ok {
					return &Float{Value: l.Value - r.Value}, nil
				}
				if r, ok := right.(*Integer); ok {
					return &Float{Value: l.Value - float64(r.Value)}, nil
				}
			}

			return nil, fmt.Errorf("Subtract: tipos incompatibles")
		},
	})

	// Multiply - Multiplica dos valores
	e.env.Set("Multiply", &BuiltinFunction{
		Name: "Multiply",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Multiply() espera 2 argumentos")
			}
			left, right := args[0], args[1]

			if l, ok := left.(*Integer); ok {
				if r, ok := right.(*Integer); ok {
					return &Integer{Value: l.Value * r.Value}, nil
				}
				if r, ok := right.(*Float); ok {
					return &Float{Value: float64(l.Value) * r.Value}, nil
				}
			}

			if l, ok := left.(*Float); ok {
				if r, ok := right.(*Float); ok {
					return &Float{Value: l.Value * r.Value}, nil
				}
				if r, ok := right.(*Integer); ok {
					return &Float{Value: l.Value * float64(r.Value)}, nil
				}
			}

			return nil, fmt.Errorf("Multiply: tipos incompatibles")
		},
	})

	// Divide - Divide dos valores
	e.env.Set("Divide", &BuiltinFunction{
		Name: "Divide",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Divide() espera 2 argumentos")
			}
			left, right := args[0], args[1]

			if r, ok := right.(*Integer); ok && r.Value == 0 {
				return &String{Value: "ERROR"}, nil
			}
			if r, ok := right.(*Float); ok && r.Value == 0.0 {
				return &String{Value: "ERROR"}, nil
			}

			if l, ok := left.(*Integer); ok {
				if r, ok := right.(*Integer); ok {
					return &Integer{Value: l.Value / r.Value}, nil
				}
				if r, ok := right.(*Float); ok {
					return &Float{Value: float64(l.Value) / r.Value}, nil
				}
			}

			if l, ok := left.(*Float); ok {
				if r, ok := right.(*Float); ok {
					return &Float{Value: l.Value / r.Value}, nil
				}
				if r, ok := right.(*Integer); ok {
					return &Float{Value: l.Value / float64(r.Value)}, nil
				}
			}

			return nil, fmt.Errorf("Divide: tipos incompatibles")
		},
	})

	// HTTP functions
	e.env.Set("http.get", &BuiltinFunction{
		Name: "http.get",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 && len(args) != 2 {
				return nil, fmt.Errorf("http.get expects 1 or 2 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("http.get expects a string URL")
			}
			headers := make(map[string]string)
			timeout := 30
			if len(args) > 1 {
				if m, ok := args[1].(*MapObject); ok {
					for k, v := range m.Pairs {
						if s, ok := v.(*String); ok {
							headers[k] = s.Value
						}
					}
				}
			}
			return e.httpGet(url.Value, headers, timeout)
		},
	})
	e.env.Set("http.post_json", &BuiltinFunction{
		Name: "http.post_json",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 && len(args) != 3 {
				return nil, fmt.Errorf("http.post_json expects 2 or 3 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("http.post_json expects a string URL")
			}
			data := args[1]
			headers := make(map[string]string)
			timeout := 30
			if len(args) > 2 {
				if m, ok := args[2].(*MapObject); ok {
					for k, v := range m.Pairs {
						if s, ok := v.(*String); ok {
							headers[k] = s.Value
						}
					}
				}
			}
			return e.httpPostJSON(url.Value, data, headers, timeout)
		},
	})
	e.env.Set("http.listen", &BuiltinFunction{
		Name: "http.listen",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("http.listen expects 2 arguments, got %d", len(args))
			}
			port, ok := args[0].(*Integer)
			if !ok {
				return nil, fmt.Errorf("http.listen expects an integer port")
			}
			handler, ok := args[1].(*ZyloFunction)
			if !ok {
				return nil, fmt.Errorf("http.listen expects a function handler")
			}
			return e.httpListen(port.Value, handler)
		},
	})
	e.env.Set("http.get_async", &BuiltinFunction{
		Name: "http.get_async",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 && len(args) != 2 {
				return nil, fmt.Errorf("http.get_async expects 1 or 2 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("http.get_async expects a string URL")
			}
			headers := make(map[string]string)
			timeout := 30
			if len(args) > 1 {
				if m, ok := args[1].(*MapObject); ok {
					for k, v := range m.Pairs {
						if s, ok := v.(*String); ok {
							headers[k] = s.Value
						}
					}
				}
			}
			return e.httpGetAsync(url.Value, headers, timeout), nil
		},
	})
	e.env.Set("http.post_json_async", &BuiltinFunction{
		Name: "http.post_json_async",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 && len(args) != 3 {
				return nil, fmt.Errorf("http.post_json_async expects 2 or 3 arguments, got %d", len(args))
			}
			url, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("http.post_json_async expects a string URL")
			}
			data := args[1]
			headers := make(map[string]string)
			timeout := 30
			if len(args) > 2 {
				if m, ok := args[2].(*MapObject); ok {
					for k, v := range m.Pairs {
						if s, ok := v.(*String); ok {
							headers[k] = s.Value
						}
					}
				}
			}
			return e.httpPostJSONAsync(url.Value, data, headers, timeout), nil
		},
	})

	// Crear objeto http con métodos
	httpObj := &MapObject{Pairs: make(map[string]Value)}
	if getFn, exists := e.env.Get("http.get"); exists {
		httpObj.Pairs["get"] = getFn
	}
	if postFn, exists := e.env.Get("http.post_json"); exists {
		httpObj.Pairs["post_json"] = postFn
	}
	if listenFn, exists := e.env.Get("http.listen"); exists {
		httpObj.Pairs["listen"] = listenFn
	}
	if getAsyncFn, exists := e.env.Get("http.get_async"); exists {
		httpObj.Pairs["get_async"] = getAsyncFn
	}
	if postAsyncFn, exists := e.env.Get("http.post_json_async"); exists {
		httpObj.Pairs["post_json_async"] = postAsyncFn
	}
	e.env.Set("http", httpObj)
}


// evaluateStatement evalúa una sentencia
func (e *Evaluator) evaluateStatement(stmt ast.Statement) (Value, error) {
	if stmt == nil {
		return nil, fmt.Errorf("nil statement")
	}

	switch s := stmt.(type) {
	case *ast.VarStatement:
		return e.evaluateVarStatement(s)
	case *ast.ExpressionStatement:
		if s.Expression == nil {
			return &Null{}, nil
		}
		return e.evaluateExpression(s.Expression)
	case *ast.FuncStatement:
		return e.evaluateFuncStatement(s)
	case *ast.ReturnStatement:
		var value Value = &Null{}
		var err error
		if s.ReturnValue != nil {
			value, err = e.evaluateExpression(s.ReturnValue)
			if err != nil {
				return nil, err
			}
		}
		return &ReturnValue{Value: value}, nil
	case *ast.IfStatement:
		return e.evaluateIfStatement(s)
	case *ast.WhileStatement:
		return e.evaluateWhileStatement(s)
	case *ast.ForInStatement:
		return e.evaluateForInStatement(s)
	case *ast.BreakStatement:
		return &BreakValue{}, nil
	case *ast.ContinueStatement:
		return &ContinueValue{}, nil
	case *ast.ClassStatement:
		return e.evaluateClassStatement(s)
	case *ast.TryStatement:
		return e.evaluateTryStatement(s)
	case *ast.ThrowStatement:
		return e.evaluateThrowStatement(s)
	case *ast.ImportStatement:
		return e.evaluateImportStatement(s)
	case *ast.BlockStatement:
		return e.evaluateBlockStatement(s)
	default:
		return nil, fmt.Errorf("sentencia no soportada: %T", s)
	}
}

// evaluateVarStatement evalúa una declaración de variable
func (e *Evaluator) evaluateVarStatement(stmt *ast.VarStatement) (Value, error) {
	var value Value = &Null{}
	var err error

	if stmt.Value != nil {
		value, err = e.evaluateExpression(stmt.Value)
		if err != nil {
			return nil, err
		}
	}

// Para variables tipadas, aseguramos compatibilidad de runtime
	expectedType := unifyType(stmt.Name.TypeAnnotation)
	actualType := getNormalizedType(value)

	if expectedType != "" && expectedType != "any" && actualType != expectedType {
		// Intentar conversion automática si es compatible
		converted, err := e.convertToTypeAuto(value, expectedType)
		if err == nil {
			value = converted
			// Actualizar el tipo después de conversión
			actualType = getNormalizedType(converted)
		} else {
			return nil, fmt.Errorf("tipo incompatible: esperado %s, recibido %s", expectedType, actualType)
		}
	}

	e.env.Set(stmt.Name.Value, value)
	e.env.SetType(stmt.Name.Value, expectedType)
	if stmt.IsConstant {
		e.env.constants[stmt.Name.Value] = true
	}
	return value, nil
}

// evaluateFuncStatement evalúa una declaración de función
func (e *Evaluator) evaluateFuncStatement(stmt *ast.FuncStatement) (Value, error) {
	zyloFunc := &ZyloFunction{
		Name:       stmt.Name.Value,
		Parameters: stmt.Parameters,
		Body:       stmt.Body,
		Env:        e.env,
		IsAsync:    stmt.IsAsync,
	}
	e.env.Set(stmt.Name.Value, zyloFunc)
	return &Null{}, nil
}

// evaluateIfStatement evalúa una sentencia if
func (e *Evaluator) evaluateIfStatement(stmt *ast.IfStatement) (Value, error) {
	condition, err := e.evaluateExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if e.isTruthy(condition) {
		return e.evaluateBlockStatement(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return e.evaluateBlockStatement(stmt.Alternative)
	}

	return &Null{}, nil
}

// evaluateTryStatement evalúa una sentencia try-catch
func (e *Evaluator) evaluateTryStatement(stmt *ast.TryStatement) (Value, error) {
	if stmt.TryBlock == nil {
		return nil, fmt.Errorf("nil try block")
	}

	result, err := e.evaluateBlockStatement(stmt.TryBlock)

	if err != nil && stmt.CatchClause != nil {
		childEnv := e.env.NewChildEnvironment()
		oldEnv := e.env
		e.env = childEnv

		if stmt.CatchClause.Parameter != nil {
			e.env.Set(stmt.CatchClause.Parameter.Value, &String{Value: err.Error()})
		}

		result, _ = e.evaluateBlockStatement(stmt.CatchClause.CatchBlock)
		e.env = oldEnv
	}

	if stmt.FinallyBlock != nil {
		_, _ = e.evaluateBlockStatement(stmt.FinallyBlock)
	}

	return result, nil
}

// evaluateThrowStatement evalúa una sentencia throw
func (e *Evaluator) evaluateThrowStatement(stmt *ast.ThrowStatement) (Value, error) {
	value, _ := e.evaluateExpression(stmt.Exception)
	if str, ok := value.(*String); ok {
		return nil, fmt.Errorf("%s", str.Value)
	}
	return nil, fmt.Errorf("thrown exception")
}

// evaluateBlockStatement evalúa un bloque de sentencias
func (e *Evaluator) evaluateBlockStatement(stmt *ast.BlockStatement) (Value, error) {
	childEnv := e.env.NewChildEnvironment()
	oldEnv := e.env
	e.env = childEnv
	defer func() { e.env = oldEnv }()

	var lastValue Value = &Null{}

	for _, bodyStmt := range stmt.Statements {
		value, err := e.evaluateStatement(bodyStmt)
		if err != nil {
			return nil, err
		}

		if _, ok := value.(*BreakValue); ok {
			return value, nil
		}
		if _, ok := value.(*ContinueValue); ok {
			return value, nil
		}

		// Propagar ReturnValue inmediatamente
		if _, ok := value.(*ReturnValue); ok {
			return value, nil
		}

		lastValue = value
	}

	return lastValue, nil
}

// evaluateWhileStatement evalúa una sentencia while
func (e *Evaluator) evaluateWhileStatement(stmt *ast.WhileStatement) (Value, error) {
	for {
		condition, err := e.evaluateExpression(stmt.Condition)
		if err != nil {
			return nil, err
		}

		if !e.isTruthy(condition) {
			break
		}

		for _, bodyStmt := range stmt.Body.Statements {
			value, err := e.evaluateStatement(bodyStmt)
			if err != nil {
				return nil, err
			}

			if _, ok := value.(*BreakValue); ok {
				return &Null{}, nil
			}
			if _, ok := value.(*ContinueValue); ok {
				break
			}
		}
	}

	return &Null{}, nil
}

// evaluateForInStatement evalúa una sentencia for in
func (e *Evaluator) evaluateForInStatement(stmt *ast.ForInStatement) (Value, error) {
	iterable, err := e.evaluateExpression(stmt.Iterable)
	if err != nil {
		return nil, err
	}

	switch iter := iterable.(type) {
	case *List:
		for _, element := range iter.Items {
			e.env.Set(stmt.Identifier.Value, element)

			result, err := e.evaluateBlockStatement(stmt.Body)
			if err != nil {
				return nil, err
			}

			if _, ok := result.(*BreakValue); ok {
				break
			}
			if _, ok := result.(*ContinueValue); ok {
				continue
			}
		}
	case *String:
		for _, char := range iter.Value {
			e.env.Set(stmt.Identifier.Value, &String{Value: string(char)})

			result, err := e.evaluateBlockStatement(stmt.Body)
			if err != nil {
				return nil, err
			}

			if _, ok := result.(*BreakValue); ok {
				break
			}
			if _, ok := result.(*ContinueValue); ok {
				continue
			}
		}
	default:
		return nil, fmt.Errorf("cannot iterate over %T", iterable)
	}

	return &Null{}, nil
}

// evaluateImportStatement evalúa una declaración de import
func (e *Evaluator) evaluateImportStatement(stmt *ast.ImportStatement) (Value, error) {
	if stmt.ModuleName == nil {
		return nil, fmt.Errorf("import sin nombre de módulo")
	}
	return &Null{}, nil
}

// evaluateClassStatement evalúa una declaración de clase
func (e *Evaluator) evaluateClassStatement(stmt *ast.ClassStatement) (Value, error) {
	classObj := &ZyloClass{
		Name:       stmt.Name.Value,
		Attributes: make(map[string]Value),
		Methods:    make(map[string]*ZyloFunction),
		InitMethod: nil,
	}

	for _, attr := range stmt.Attributes {
		if attr.Value != nil {
			value, err := e.evaluateExpression(attr.Value)
			if err != nil {
				return nil, err
			}
			classObj.Attributes[attr.Name.Value] = value
		} else {
			classObj.Attributes[attr.Name.Value] = &Null{}
		}
	}

	for _, method := range stmt.Methods {
		zyloFunc := &ZyloFunction{
			Name:       method.Name.Value,
			Parameters: method.Parameters,
			Body:       method.Body,
			Env:        e.env,
			IsAsync:    method.IsAsync,
		}
		classObj.Methods[method.Name.Value] = zyloFunc
	}

	if stmt.InitMethod != nil {
		zyloFunc := &ZyloFunction{
			Name:       "init",
			Parameters: stmt.InitMethod.Parameters,
			Body:       stmt.InitMethod.Body,
			Env:        e.env,
		}
		classObj.InitMethod = zyloFunc
	}

	// Set superclass
	if stmt.SuperClass != nil {
		if superClass, exists := e.env.Get(stmt.SuperClass.Value); exists {
			if zyloSuperClass, ok := superClass.(*ZyloClass); ok {
				classObj.SuperClass = zyloSuperClass
			}
		}
	}

	e.env.Set(stmt.Name.Value, classObj)
	return &Null{}, nil
}

// evaluateExpression evalúa una expresión
func (e *Evaluator) evaluateExpression(exp ast.Expression) (Value, error) {
	const MaxEvaluateDepth = 10000
	if e.evaluateDepth > MaxEvaluateDepth {
		return nil, fmt.Errorf("evaluation depth too deep")
	}
	e.evaluateDepth++
	defer func() { e.evaluateDepth-- }()

	if exp == nil {
		return nil, fmt.Errorf("nil expression")
	}

	switch ex := exp.(type) {
	case *ast.Identifier:
		return e.evaluateIdentifier(ex)
	case *ast.StringLiteral:
		return &String{Value: ex.Value}, nil
	case *ast.NumberLiteral:
		if ex.Value == nil {
			return &Integer{Value: 0}, nil
		}

		switch v := ex.Value.(type) {
		case float64:
			return &Float{Value: v}, nil
		case int64:
			return &Integer{Value: v}, nil
		case int:
			return &Integer{Value: int64(v)}, nil
		default:
			// Intentar convertir si es otro tipo
			return &Integer{Value: 0}, fmt.Errorf("tipo de número no soportado: %T", ex.Value)
		}
	case *ast.BooleanLiteral:
		return &Boolean{Value: ex.Value}, nil
	case *ast.NullLiteral:
		return &Null{}, nil
	case *ast.CallExpression:
		return e.evaluateCallExpression(ex)
	case *ast.DotExpression:
		return e.evaluateDotExpression(ex)
	case *ast.MemberExpression:
		return e.evaluateMemberExpression(ex)
	case *ast.ListLiteral:
		elements := make([]Value, len(ex.Elements))
		for i, el := range ex.Elements {
			var err error
			elements[i], err = e.evaluateExpression(el)
			if err != nil {
				return nil, err
			}
		}
		return &List{Items: elements}, nil
	case *ast.MapLiteral:
	    pairs := make(map[string]Value)
	    for k, v := range ex.Pairs {
	        value, err := e.evaluateExpression(v)
	        if err != nil {
	            return nil, err
	        }
	        pairs[k] = value
	    }
	    return &MapObject{Pairs: pairs}, nil
	case *ast.IndexExpression:
		left, err := e.evaluateExpression(ex.Left)
		if err != nil {
			return nil, err
		}
		index, err := e.evaluateExpression(ex.Index)
		if err != nil {
			return nil, err
		}
		return e.indexValue(left, index)
	case *ast.RangeExpression:
		return e.evaluateRangeExpression(ex)
	case *ast.InfixExpression:
		return e.evaluateInfixExpression(ex)
	case *ast.PrefixExpression:
		return e.evaluatePrefixExpression(ex)
	case *ast.AssignmentExpression:
		return e.evaluateAssignmentExpression(ex)
	case *ast.ThisExpression:
		return e.evaluateThisExpression(ex)
	case *ast.SuperExpression:
		return e.evaluateSuperExpression(ex)
	case *ast.AwaitExpression:
		return e.evaluateAwaitExpression(ex)
	case *ast.AsExpression:
		return e.evaluateAsExpression(ex)
	case *ast.BlockExpression:
		// Un BlockExpression en contexto de expresión evalúa el bloque y retorna su último valor
		if ex.Block != nil {
			return e.evaluateBlockStatement(ex.Block)
		}
		return nil, fmt.Errorf("BlockExpression sin BlockStatement válido")
	case *ast.FunctionLiteral:
		zyloFunc := &ZyloFunction{
			Name:       "", // anonymous function
			Parameters: ex.Parameters,
			Body:       ex.Body,
			Env:        e.env,
		}
		return zyloFunc, nil
	default:
		return nil, fmt.Errorf("expresión no soportada: %T", ex)
	}
}

// evaluateDotExpression evalúa expresiones de punto como show.log
func (e *Evaluator) evaluateDotExpression(exp *ast.DotExpression) (Value, error) {
	if exp.Left == nil {
		return nil, fmt.Errorf("nil left side in dot expression")
	}
	if exp.Property == nil {
		return nil, fmt.Errorf("nil property in dot expression")
	}

	if identifier, ok := exp.Left.(*ast.Identifier); ok {
		objName := identifier.Value
		propName := exp.Property.Value
		fullName := objName + "." + propName

		if builtin, exists := e.env.Get(fullName); exists {
			return builtin, nil
		}
	}

	obj, err := e.evaluateExpression(exp.Left)
	if err != nil {
		return nil, err
	}

	if list, ok := obj.(*List); ok {
		switch exp.Property.Value {
		case "length":
			return &Integer{Value: int64(len(list.Items))}, nil
		case "append":
			return &BuiltinFunction{
				Name: "List.append",
				Fn: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("append() espera 1 argumento")
					}
					list.Items = append(list.Items, args[0])
					return &Null{}, nil
				},
			}, nil
		}
	}

	if instance, ok := obj.(*ZyloInstance); ok {
		if field, exists := instance.Fields[exp.Property.Value]; exists {
			return field, nil
		}
		// Check methods in class and superclasses
		currentClass := instance.Class
		for currentClass != nil {
			if method, exists := currentClass.Methods[exp.Property.Value]; exists {
				return &BoundMethod{
					Instance: instance,
					Method:   method,
				}, nil
			}
			currentClass = currentClass.SuperClass
		}
	}

	if superObj, ok := obj.(*SuperObject); ok {
		if exp.Property.Value == "init" {
			if superObj.Instance.Class.SuperClass.InitMethod != nil {
				return &BoundMethod{
					Instance: superObj.Instance,
					Method:   superObj.Instance.Class.SuperClass.InitMethod,
				}, nil
			}
		}
		if method, exists := superObj.Instance.Class.SuperClass.Methods[exp.Property.Value]; exists {
			return &BoundMethod{
				Instance: superObj.Instance,
				Method:   method,
			}, nil
		}
	}

	return nil, fmt.Errorf("property '%s' not found", exp.Property.Value)
}

// evaluateIdentifier evalúa un identificador
func (e *Evaluator) evaluateIdentifier(exp *ast.Identifier) (Value, error) {
	// Manejar identificadores especiales del parser
	if exp.Value == "IGNORED_SEPARATOR" {
		return &Null{}, nil
	}

	// Manejar 'super'
	if exp.Value == "super" {
		if this, exists := e.env.Get("this"); exists {
			if instance, ok := this.(*ZyloInstance); ok && instance.Class.SuperClass != nil {
				return &SuperObject{Instance: instance}, nil
			}
		}
		return nil, fmt.Errorf("'super' no disponible en este contexto")
	}

	value, exists := e.env.Get(exp.Value)
	if !exists {
		return nil, fmt.Errorf("variable no definida: %s", exp.Value)
	}
	return value, nil
}

// evaluateMemberExpression evalúa una expresión de acceso a miembro
func (e *Evaluator) evaluateMemberExpression(exp *ast.MemberExpression) (Value, error) {
	obj, err := e.evaluateExpression(exp.Object)
	if err != nil {
		return nil, err
	}

	if list, ok := obj.(*List); ok {
		switch exp.Property.Value {
		case "Get":
			return &BuiltinFunction{
				Name: "List.Get",
				Fn: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("Get() espera 1 argumento")
					}
					idx, ok := args[0].(*Integer)
					if !ok {
						return nil, fmt.Errorf("índice debe ser integer")
					}
					if idx.Value < 0 || int(idx.Value) >= len(list.Items) {
						return nil, fmt.Errorf("índice fuera de rango")
					}
					return list.Items[idx.Value], nil
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("member '%s' not found", exp.Property.Value)
}

// evaluateCallExpression evalúa una llamada a función
func (e *Evaluator) evaluateCallExpression(exp *ast.CallExpression) (Value, error) {
	const MaxCallDepth = 100000
	if e.callDepth > MaxCallDepth {
		return nil, fmt.Errorf("stack overflow: recursion too deep")
	}
	e.callDepth++
	defer func() { e.callDepth-- }()

	fn, err := e.evaluateExpression(exp.Function)
	if err != nil {
		return nil, err
	}

	if class, ok := fn.(*ZyloClass); ok {
		return e.instantiateClass(class, exp.Arguments)
	}

	args := make([]Value, len(exp.Arguments))
	for i, arg := range exp.Arguments {
		args[i], err = e.evaluateExpression(arg)
		if err != nil {
			return nil, err
		}
	}

	return e.callFunction(fn, args)
}

// evaluateInfixExpression evalúa una expresión infija
func (e *Evaluator) evaluateInfixExpression(exp *ast.InfixExpression) (Value, error) {
	left, err := e.evaluateExpression(exp.Left)
	if err != nil {
		return nil, err
	}

	// Short-circuit evaluation para && y ||
	switch exp.Operator {
	case "and", "&&":
		// Si el izquierdo es falso, retornar false sin evaluar el derecho
		if !e.isTruthy(left) {
			return &Boolean{Value: false}, nil
		}
		// Si el izquierdo es verdadero, evaluar el derecho
		right, err := e.evaluateExpression(exp.Right)
		if err != nil {
			return nil, err
		}
		return &Boolean{Value: e.isTruthy(right)}, nil

	case "or", "||":
		// Si el izquierdo es verdadero, retornar true sin evaluar el derecho
		if e.isTruthy(left) {
			return &Boolean{Value: true}, nil
		}
		// Si el izquierdo es falso, evaluar el derecho
		right, err := e.evaluateExpression(exp.Right)
		if err != nil {
			return nil, err
		}
		return &Boolean{Value: e.isTruthy(right)}, nil

	default:
		// Para otros operadores, evaluar normalmente
		right, err := e.evaluateExpression(exp.Right)
		if err != nil {
			return nil, err
		}
		return e.applyOperator(exp.Operator, left, right)
	}
}

// evaluatePrefixExpression evalúa una expresión prefija
func (e *Evaluator) evaluatePrefixExpression(exp *ast.PrefixExpression) (Value, error) {
	right, err := e.evaluateExpression(exp.Right)
	if err != nil {
		return nil, err
	}

	switch exp.Operator {
	case "!":
		return &Boolean{Value: !e.isTruthy(right)}, nil
	case "not":
		return &Boolean{Value: !e.isTruthy(right)}, nil
	case "-":
		if num, ok := right.(*Integer); ok {
			return &Integer{Value: -num.Value}, nil
		}
		if num, ok := right.(*Float); ok {
			return &Float{Value: -num.Value}, nil
		}
		return nil, fmt.Errorf("operador '-' no soportado para %T", right)
	default:
		return nil, fmt.Errorf("operador prefijo no soportado: %s", exp.Operator)
	}
}

// evaluateAssignmentExpression evalúa una asignación
func (e *Evaluator) evaluateAssignmentExpression(exp *ast.AssignmentExpression) (Value, error) {
	value, err := e.evaluateExpression(exp.Value)
	if err != nil {
		return nil, err
	}

	// Determine the target of the assignment (identifier, index, or dot expression)
	switch nameExp := exp.Name.(type) {
	case *ast.Identifier:
		// Check if it's a constant
		if e.env.IsConstant(nameExp.Value) {
			return nil, fmt.Errorf("no se puede reasignar constante: %s", nameExp.Value)
		}
		// Check type compatibility - only if explicit type annotation was used
		expectedType, exists := e.env.GetType(nameExp.Value)
		if exists && expectedType != "" && expectedType != "ANY" {
			if !e.isTypeCompatible(value, expectedType) {
				return nil, fmt.Errorf("tipo incompatible: esperado %s, recibido %s", expectedType, e.getValueType(value))
			}
		}
		// Handle simple identifier assignment
		if exp.Operator != "=" {
			oldValue, exists := e.env.Get(nameExp.Value)
			if !exists {
				return nil, fmt.Errorf("variable no definida: %s", nameExp.Value)
			}
			var baseOp string
			switch exp.Operator {
			case "+=":
				baseOp = "+"
			case "-=":
				baseOp = "-"
			case "*=":
				baseOp = "*"
			case "/=":
				baseOp = "/"
			case "%=":
				baseOp = "%"
			default:
				return nil, fmt.Errorf("operador de asignación no soportado: %s", exp.Operator)
			}
			value, err = e.applyOperator(baseOp, oldValue, value)
			if err != nil {
				return nil, err
			}
		}
		if !e.env.Update(nameExp.Value, value) {
			return nil, fmt.Errorf("variable no definida: %s", nameExp.Value)
		}
	case *ast.IndexExpression:
		// Handle index assignment (e.g., list[0] = 10)
		left, err := e.evaluateExpression(nameExp.Left)
		if err != nil {
			return nil, err
		}
		index, err := e.evaluateExpression(nameExp.Index)
		if err != nil {
			return nil, err
		}
		return e.assignIndexValue(left, index, value, exp.Operator)
	case *ast.DotExpression:
		// Handle dot assignment (e.g., obj.prop = 10)
		obj, err := e.evaluateExpression(nameExp.Left)
		if err != nil {
			return nil, err
		}
		return e.assignDotValue(obj, nameExp.Property.Value, value, exp.Operator)
	default:
		return nil, fmt.Errorf("lado izquierdo de la asignación no es asignable: %T", exp.Name)
	}

	return value, nil
}

// assignIndexValue asigna un valor a un índice de una lista o mapa
func (e *Evaluator) assignIndexValue(left, index, value Value, operator string) (Value, error) {
	switch l := left.(type) {
	case *List:
		idx, ok := index.(*Integer)
		if !ok {
			return nil, fmt.Errorf("índice de lista debe ser integer")
		}
		if idx.Value < 0 || int(idx.Value) >= len(l.Items) {
			return nil, fmt.Errorf("índice de lista fuera de rango")
		}
		if operator != "=" {
			oldValue := l.Items[idx.Value]
			newValue, err := e.applyOperator(strings.TrimSuffix(operator, "="), oldValue, value)
			if err != nil {
				return nil, err
			}
			l.Items[idx.Value] = newValue
		} else {
			l.Items[idx.Value] = value
		}
		return value, nil
	case *MapObject:
		key, ok := index.(*String)
		if !ok {
			return nil, fmt.Errorf("clave de mapa debe ser string")
		}
		if operator != "=" {
			oldValue, exists := l.Pairs[key.Value]
			if !exists {
				return nil, fmt.Errorf("clave de mapa no definida: %s", key.Value)
			}
			newValue, err := e.applyOperator(strings.TrimSuffix(operator, "="), oldValue, value)
			if err != nil {
				return nil, err
			}
			l.Pairs[key.Value] = newValue
		} else {
			l.Pairs[key.Value] = value
		}
		return value, nil
	default:
		return nil, fmt.Errorf("no se puede asignar a índice de tipo %T", left)
	}
}

// assignDotValue asigna un valor a una propiedad de un objeto
func (e *Evaluator) assignDotValue(obj Value, property string, value Value, operator string) (Value, error) {
	switch o := obj.(type) {
	case *ZyloInstance:
		if operator != "=" {
			oldValue, exists := o.Fields[property]
			if !exists {
				return nil, fmt.Errorf("propiedad de instancia no definida: %s", property)
			}
			newValue, err := e.applyOperator(strings.TrimSuffix(operator, "="), oldValue, value)
			if err != nil {
				return nil, err
			}
			o.Fields[property] = newValue
		} else {
			o.Fields[property] = value
		}
		return value, nil
	default:
		return nil, fmt.Errorf("no se puede asignar a propiedad de tipo %T", obj)
	}
}

// callFunction llama a una función
func (e *Evaluator) callFunction(fn Value, args []Value) (Value, error) {
	switch f := fn.(type) {
	case *ZyloFunction:
		return e.callZyloFunction(f, args)
	case *BuiltinFunction:
		return f.Fn(args)
	case *BoundMethod:
		return e.callBoundMethod(f, args)
	default:
		return nil, fmt.Errorf("no se puede llamar a: %T", fn)
	}
}

// instantiateClass crea una instancia de una clase
func (e *Evaluator) instantiateClass(class *ZyloClass, args []ast.Expression) (Value, error) {
	instance := &ZyloInstance{
		Class:  class,
		Fields: make(map[string]Value),
	}

	for name, value := range class.Attributes {
		instance.Fields[name] = value
	}

	if class.InitMethod != nil {
		evalArgs := make([]Value, len(args))
		for i, arg := range args {
			var err error
			evalArgs[i], err = e.evaluateExpression(arg)
			if err != nil {
				return nil, err
			}
		}

		funcEnv := class.InitMethod.Env.NewChildEnvironment()
		funcEnv.Set("this", instance)

		for i, param := range class.InitMethod.Parameters {
			if i < len(evalArgs) {
				funcEnv.Set(param.Value, evalArgs[i])
			}
		}

		oldEnv := e.env
		e.env = funcEnv
		defer func() { e.env = oldEnv }()

		_, err := e.evaluateBlockStatement(class.InitMethod.Body)
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

// callZyloFunction llama a una función Zylo
func (e *Evaluator) callZyloFunction(fn *ZyloFunction, args []Value) (Value, error) {
	if fn.IsAsync {
		future := &Future{
			Result: make(chan ZyloObject, 1),
			value:  nil,
			once:   false,
		}
		go func() {
			result, _ := e.callZyloFunctionSync(fn, args)
			future.Result <- result.(ZyloObject)
		}()
		return future, nil
	}
	return e.callZyloFunctionSync(fn, args)
}

// callZyloFunctionSync llama a una función Zylo de forma síncrona
func (e *Evaluator) callZyloFunctionSync(fn *ZyloFunction, args []Value) (Value, error) {
	funcEnv := NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		if i < len(args) {
			funcEnv.Set(param.Value, args[i])
		}
	}

	oldEnv := e.env
	e.env = funcEnv
	defer func() { e.env = oldEnv }()

	result, err := e.evaluateBlockStatement(fn.Body)
	if err != nil {
		return nil, err
	}

	// Desenvolver ReturnValue
	if returnValue, ok := result.(*ReturnValue); ok {
		return returnValue.Value, nil
	}

	return result, nil
}

// callBoundMethod llama a un método ligado
func (e *Evaluator) callBoundMethod(boundMethod *BoundMethod, args []Value) (Value, error) {
	funcEnv := boundMethod.Method.Env.NewChildEnvironment()
	funcEnv.Set("this", boundMethod.Instance)

	for i, param := range boundMethod.Method.Parameters {
		if i < len(args) {
			funcEnv.Set(param.Value, args[i])
		}
	}

	oldEnv := e.env
	e.env = funcEnv
	defer func() { e.env = oldEnv }()

	result, err := e.evaluateBlockStatement(boundMethod.Method.Body)
	if err != nil {
		return nil, err
	}

	// Desenvolver ReturnValue
	if returnValue, ok := result.(*ReturnValue); ok {
		return returnValue.Value, nil
	}

	return result, nil
}

// evaluateThisExpression evalúa una expresión 'this'
func (e *Evaluator) evaluateThisExpression(exp *ast.ThisExpression) (Value, error) {
	value, exists := e.env.Get("this")
	if !exists {
		return nil, fmt.Errorf("'this' no disponible en este contexto")
	}
	return value, nil
}

// evaluateSuperExpression evalúa una expresión 'super'
func (e *Evaluator) evaluateSuperExpression(exp *ast.SuperExpression) (Value, error) {
	if this, exists := e.env.Get("this"); exists {
		if instance, ok := this.(*ZyloInstance); ok && instance.Class.SuperClass != nil {
			return &SuperObject{Instance: instance}, nil
		}
	}
	return nil, fmt.Errorf("'super' no disponible en este contexto")
}

// evaluateAwaitExpression evalúa una expresión 'await'
func (e *Evaluator) evaluateAwaitExpression(exp *ast.AwaitExpression) (Value, error) {
	arg, err := e.evaluateExpression(exp.Argument)
	if err != nil {
		return nil, err
	}

	if future, ok := arg.(*Future); ok {
		result := <-future.Result
		return result, nil
	}

	return nil, fmt.Errorf("await expects a future, got %T", arg)
}

// applyOperator aplica un operador binario
func (e *Evaluator) applyOperator(operator string, left, right Value) (Value, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("operandos nulos para '%s'", operator)
	}

	switch operator {
	case "+":
		if leftStr, ok := left.(*String); ok {
			if rightStr, ok := right.(*String); ok {
				return &String{Value: leftStr.Value + rightStr.Value}, nil
			}
			if rightNum, ok := right.(*Integer); ok {
				return &String{Value: leftStr.Value + fmt.Sprintf("%d", rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &String{Value: leftStr.Value + fmt.Sprintf("%g", rightFloat.Value)}, nil
			}
			if _, ok := right.(*Null); ok {
				return &String{Value: leftStr.Value + "null"}, nil
			}
		}
		if _, ok := left.(*Null); ok {
			if rightStr, ok := right.(*String); ok {
				return &String{Value: "null" + rightStr.Value}, nil
			}
		}
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Integer{Value: leftNum.Value + rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: float64(leftNum.Value) + rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Float{Value: leftFloat.Value + float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: leftFloat.Value + rightFloat.Value}, nil
			}
		}
	case "-":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Integer{Value: leftNum.Value - rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: float64(leftNum.Value) - rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Float{Value: leftFloat.Value - float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: leftFloat.Value - rightFloat.Value}, nil
			}
		}
	case "*":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Integer{Value: leftNum.Value * rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: float64(leftNum.Value) * rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Float{Value: leftFloat.Value * float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: leftFloat.Value * rightFloat.Value}, nil
			}
		}
	case "/":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				if rightNum.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Integer{Value: leftNum.Value / rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				if rightFloat.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Float{Value: float64(leftNum.Value) / rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				if rightNum.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Float{Value: leftFloat.Value / float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				if rightFloat.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Float{Value: leftFloat.Value / rightFloat.Value}, nil
			}
		}
	case "%":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				if rightNum.Value == 0 {
					return nil, fmt.Errorf("módulo por cero")
				}
				return &Integer{Value: leftNum.Value % rightNum.Value}, nil
			}
		}
	case "**", "^":
		switch l := left.(type) {
		case *Integer:
			switch r := right.(type) {
			case *Integer:
				return &Float{Value: pow(float64(l.Value), float64(r.Value))}, nil
			case *Float:
				return &Float{Value: pow(float64(l.Value), r.Value)}, nil
			}
		case *Float:
			switch r := right.(type) {
			case *Integer:
				return &Float{Value: pow(l.Value, float64(r.Value))}, nil
			case *Float:
				return &Float{Value: pow(l.Value, r.Value)}, nil
			}
		}
	
	case "==":
		if leftStr, ok := left.(*String); ok {
			if rightStr, ok := right.(*String); ok {
				return &Boolean{Value: leftStr.Value == rightStr.Value}, nil
			}
		}
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value == rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) == rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value == float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value == rightFloat.Value}, nil
			}
		}
		if leftBool, ok := left.(*Boolean); ok {
			if rightBool, ok := right.(*Boolean); ok {
				return &Boolean{Value: leftBool.Value == rightBool.Value}, nil
			}
		}
		return &Boolean{Value: false}, nil
	case "!=":
		result, err := e.applyOperator("==", left, right)
		if err != nil {
			return nil, err
		}
		if b, ok := result.(*Boolean); ok {
			return &Boolean{Value: !b.Value}, nil
		}
		return &Boolean{Value: true}, nil
		
	case "<":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value < rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) < rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value < float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value < rightFloat.Value}, nil
			}
		}
	case ">":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value > rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) > rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value > float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value > rightFloat.Value}, nil
			}
		}
	case "<=":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value <= rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) <= rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value <= float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value <= rightFloat.Value}, nil
			}
		}
	case ">=":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value >= rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) >= rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value >= float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value >= rightFloat.Value}, nil
			}
		}
	case "and", "&&":
		leftBool := e.isTruthy(left)
		if !leftBool {
			return &Boolean{Value: false}, nil
		}
		rightBool := e.isTruthy(right)
		return &Boolean{Value: rightBool}, nil
	case "or", "||":
		leftBool := e.isTruthy(left)
		if leftBool {
			return &Boolean{Value: true}, nil
		}
		rightBool := e.isTruthy(right)
		return &Boolean{Value: rightBool}, nil
	}

	return nil, fmt.Errorf("operador '%s' no soportado para %T y %T", operator, left, right)
	
}

// isTruthy determina si un valor es "verdadero"
func (e *Evaluator) isTruthy(value Value) bool {
	if value == nil {
		return false
	}
	if boolVal, ok := value.(*Boolean); ok {
		return boolVal.Value
	}
	if intVal, ok := value.(*Integer); ok {
		return intVal.Value != 0
	}
	if strVal, ok := value.(*String); ok {
		return len(strVal.Value) > 0
	}
	if _, ok := value.(*Null); ok {
		return false
	}
	return true
}

// getValueType devuelve el tipo de un valor como string
func (e *Evaluator) getValueType(value Value) string {
	switch value.(type) {
	case *Integer:
		return "INT"
	case *Float:
		return "FLOAT"
	case *String:
		return "STRING"
	case *Boolean:
		return "BOOL"
	case *List:
		return "LIST"
	case *MapObject:
		return "MAP"
	case *Null:
		return "NULL"
	default:
		return "UNKNOWN"
	}
}

// isTypeCompatible verifica si un valor es compatible con un tipo esperado
func (e *Evaluator) isTypeCompatible(value Value, expectedType string) bool {
	actualType := e.getValueType(value)
	if expectedType == "ANY" {
		return true
	}
	return actualType == expectedType
}

// evaluateRangeExpression evalúa una expresión de rango (e.g., 0..10)
func (e *Evaluator) evaluateRangeExpression(exp *ast.RangeExpression) (Value, error) {
	start, err := e.evaluateExpression(exp.Start)
	if err != nil {
		return nil, err
	}
	end, err := e.evaluateExpression(exp.End)
	if err != nil {
		return nil, err
	}

	startInt, ok := start.(*Integer)
	if !ok {
		return nil, fmt.Errorf("range start must be integer")
	}
	endInt, ok := end.(*Integer)
	if !ok {
		return nil, fmt.Errorf("range end must be integer")
	}

	var items []Value
	for i := startInt.Value; i < endInt.Value; i++ {
		items = append(items, &Integer{Value: i})
	}
	return &List{Items: items}, nil
}

// indexValue handles indexing for arrays and strings
func (e *Evaluator) indexValue(left, index Value) (Value, error) {
	if left == nil {
		return nil, fmt.Errorf("no se puede indexar valor nulo")
	}

	// Si el objeto a indexar es Null, devolver Null
	if _, ok := left.(*Null); ok {
		return &Null{}, nil
	}

	switch l := left.(type) {
	case *List:
		idx, ok := index.(*Integer)
		if !ok {
			return nil, fmt.Errorf("índice debe ser integer")
		}
		if idx.Value < 0 || int(idx.Value) >= len(l.Items) {
			return nil, fmt.Errorf("índice fuera de rango")
		}
		return l.Items[idx.Value], nil
	case *String:
		idx, ok := index.(*Integer)
		if !ok {
			return nil, fmt.Errorf("índice debe ser integer")
		}
		if idx.Value < 0 || int(idx.Value) >= len(l.Value) {
			return nil, fmt.Errorf("índice fuera de rango")
		}
		return &String{Value: string(l.Value[idx.Value])}, nil
	case *MapObject:
		key, ok := index.(*String)
		if !ok {
			return nil, fmt.Errorf("clave de mapa debe ser string")
		}
		value, exists := l.Pairs[key.Value]
		if !exists {
			return &Null{}, nil // Devolver Null si la clave no existe
		}
		return value, nil
	default:
		return nil, fmt.Errorf("no se puede indexar %T", left)
	}
}

// ZyloFunction representa una función definida en Zylo
type ZyloFunction struct {
	Name       string
	Parameters []*ast.Identifier
	ReturnType string
	Body       *ast.BlockStatement
	Env        *Environment
	IsAsync    bool
}

// BuiltinFunction representa una función built-in
type BuiltinFunction struct {
	Name string
	Fn   func([]Value) (Value, error)
}

// ZyloClass representa una clase definida en Zylo
type ZyloClass struct {
	Name       string
	Attributes map[string]Value
	Methods    map[string]*ZyloFunction
	InitMethod *ZyloFunction
	SuperClass *ZyloClass
}

func (c *ZyloClass) Type() string { return "CLASS_OBJ" }
func (c *ZyloClass) Inspect() string {
	return fmt.Sprintf("class %s", c.Name)
}

// ZyloInstance representa una instancia de una clase Zylo
type ZyloInstance struct {
	Class  *ZyloClass
	Fields map[string]Value
}

func (i *ZyloInstance) Type() string { return "INSTANCE_OBJ" }
func (i *ZyloInstance) Inspect() string {
	return fmt.Sprintf("instance of %s", i.Class.Name)
}

// BoundMethod representa un método ligado a una instancia
type BoundMethod struct {
	Instance *ZyloInstance
	Method   *ZyloFunction
}

func (b *BoundMethod) Type() string { return "BOUND_METHOD_OBJ" }
func (b *BoundMethod) Inspect() string {
	return fmt.Sprintf("bound method %s", b.Method.Name)
}

// SuperObject representa el acceso a la superclase
type SuperObject struct {
	Instance *ZyloInstance
}

func (s *SuperObject) Type() string { return "SUPER_OBJ" }
func (s *SuperObject) Inspect() string {
	return "super"
}

// Control flow types for break and continue
type BreakValue struct{}
type ContinueValue struct{}

// ReturnValue representa un valor de retorno
type ReturnValue struct {
	Value Value
}

func (b *BreakValue) Type() string    { return "BREAK_OBJ" }
func (b *BreakValue) Inspect() string { return "break" }

func (c *ContinueValue) Type() string    { return "CONTINUE_OBJ" }
func (c *ContinueValue) Inspect() string { return "continue" }

func (r *ReturnValue) Type() string { return "RETURN_VALUE_OBJ" }
func (r *ReturnValue) Inspect() string {
	if r.Value == nil {
		return "return"
	}
	if obj, ok := r.Value.(ZyloObject); ok {
		return obj.Inspect()
	}
	return fmt.Sprintf("%v", r.Value)
}

// unifyType normaliza tipos a minúscula obligatoria
func unifyType(t string) string {
	switch strings.ToLower(t) {
	case "int", "integer", "INT", "INTEGER":
		return "int"
	case "float", "FLOAT":
		return "float"
	case "string", "STRING":
		return "string"
	case "bool", "boolean", "BOOL", "BOOLEAN":
		return "bool"
	default:
		return t
	}
}

// getNormalizedType obtiene el tipo normalizado de un valor
func getNormalizedType(value Value) string {
	switch value.(type) {
	case *Integer:
		return "int"
	case *Float:
		return "float"
	case *String:
		return "string"
	case *Boolean:
		return "bool"
	case *List:
		return "list"
	case *MapObject:
		return "map"
	case *Null:
		return "null"
	case *ZyloClass:
		return "class"
	case *ZyloInstance:
		return "instance"
	default:
		return "unknown"
	}
}

// evaluateAsExpression evalúa una expresión de conversión de tipo (e.g., value as Type).
func (e *Evaluator) evaluateAsExpression(exp *ast.AsExpression) (Value, error) {
	value, err := e.evaluateExpression(exp.Left)
	if err != nil {
		return nil, err
	}

	switch exp.TypeName {
	case "Int", "int":
		return e.convertToInt(value)
	case "Float", "float":
		return e.convertToFloat(value)
	case "String", "string":
		return e.convertToString(value)
	case "Bool", "bool":
		return e.convertToBool(value)
	default:
		return nil, fmt.Errorf("tipo desconocido para conversión: %s", exp.TypeName)
	}
}

// convertToInt convierte un valor a entero
func (e *Evaluator) convertToInt(value Value) (Value, error) {
	switch v := value.(type) {
	case *Integer:
		return v, nil
	case *Float:
		return &Integer{Value: int64(v.Value)}, nil
	case *String:
		if n, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
			return &Integer{Value: n}, nil
		}
		return nil, fmt.Errorf("no se puede convertir string '%s' a int", v.Value)
	case *Boolean:
		if v.Value {
			return &Integer{Value: 1}, nil
		}
		return &Integer{Value: 0}, nil
	default:
		return nil, fmt.Errorf("no se puede convertir %T a int", value)
	}
}

// convertToFloat convierte un valor a float
func (e *Evaluator) convertToFloat(value Value) (Value, error) {
	switch v := value.(type) {
	case *Float:
		return v, nil
	case *Integer:
		return &Float{Value: float64(v.Value)}, nil
	case *String:
		if f, err := strconv.ParseFloat(v.Value, 64); err == nil {
			return &Float{Value: f}, nil
		}
		return nil, fmt.Errorf("no se puede convertir string '%s' a float", v.Value)
	default:
		return nil, fmt.Errorf("no se puede convertir %T a float", value)
	}
}

// convertToString convierte un valor a string
func (e *Evaluator) convertToString(value Value) (Value, error) {
	switch v := value.(type) {
	case *String:
		return v, nil
	case *Integer:
		return &String{Value: fmt.Sprintf("%d", v.Value)}, nil
	case *Float:
		return &String{Value: fmt.Sprintf("%g", v.Value)}, nil
	case *Boolean:
		return &String{Value: fmt.Sprintf("%t", v.Value)}, nil
	case *Null:
		return &String{Value: "null"}, nil
	default:
		if obj, ok := v.(ZyloObject); ok {
			return &String{Value: obj.Inspect()}, nil
		}
		return &String{Value: fmt.Sprintf("%v", v)}, nil
	}
}

// convertToBool convierte un valor a booleano
func (e *Evaluator) convertToBool(value Value) (Value, error) {
	return &Boolean{Value: e.isTruthy(value)}, nil
}

// convertToTypeAuto intenta convertir un valor al tipo esperado automáticamente
func (e *Evaluator) convertToTypeAuto(value Value, expectedType string) (Value, error) {
	switch expectedType {
	case "int":
		return e.convertToInt(value)
	case "float":
		return e.convertToFloat(value)
	case "string":
		return e.convertToString(value)
	case "bool":
		return e.convertToBool(value)
	default:
		// Si no hay conversión automática posible, devolver el valor original
		return value, nil
	}
}



// httpGet realiza una petición GET HTTP
func (e *Evaluator) httpGet(url string, headers map[string]string, timeout int) (Value, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &String{Value: fmt.Sprintf("Error creating request: %v", err)}, nil
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &String{Value: fmt.Sprintf("Error making request: %v", err)}, nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &String{Value: fmt.Sprintf("Error reading response: %v", err)}, nil
	}

	responseMap := &MapObject{Pairs: make(map[string]Value)}
	responseMap.Pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
	responseMap.Pairs["body"] = &String{Value: string(body)}

	headersMap := &MapObject{Pairs: make(map[string]Value)}
	for key, values := range resp.Header {
		headersMap.Pairs[key] = &String{Value: strings.Join(values, ", ")}
	}
	responseMap.Pairs["headers"] = headersMap

	return responseMap, nil
}

// httpPostJSON realiza una petición POST con JSON
func (e *Evaluator) httpPostJSON(url string, data Value, headers map[string]string, timeout int) (Value, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var jsonData []byte
	var err error

	switch d := data.(type) {
	case *MapObject:
		jsonData, err = json.Marshal(e.mapToGoMap(d))
		if err != nil {
			return &String{Value: fmt.Sprintf("Error marshaling JSON: %v", err)}, nil
		}
	case *List:
		jsonData, err = json.Marshal(e.listToGoSlice(d))
		if err != nil {
			return &String{Value: fmt.Sprintf("Error marshaling JSON: %v", err)}, nil
		}
	default:
		jsonData, err = json.Marshal(e.valueToInterface(data))
		if err != nil {
			return &String{Value: fmt.Sprintf("Error marshaling JSON: %v", err)}, nil
		}
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return &String{Value: fmt.Sprintf("Error creating request: %v", err)}, nil
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &String{Value: fmt.Sprintf("Error making request: %v", err)}, nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &String{Value: fmt.Sprintf("Error reading response: %v", err)}, nil
	}

	responseMap := &MapObject{Pairs: make(map[string]Value)}
	responseMap.Pairs["status"] = &Integer{Value: int64(resp.StatusCode)}
	responseMap.Pairs["body"] = &String{Value: string(body)}

	headersMap := &MapObject{Pairs: make(map[string]Value)}
	for key, values := range resp.Header {
		headersMap.Pairs[key] = &String{Value: strings.Join(values, ", ")}
	}
	responseMap.Pairs["headers"] = headersMap

	return responseMap, nil
}

// httpListen inicia un servidor HTTP
func (e *Evaluator) httpListen(port int64, handler *ZyloFunction) (Value, error) {
	if e.httpServer != nil {
		return &String{Value: "Server already running"}, nil
	}

	e.httpHandler = handler

	mux := http.NewServeMux()
	mux.HandleFunc("/", e.httpHandleRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	e.httpServer = server

	go func() {
		fmt.Printf("HTTP server listening on port %d\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	return &String{Value: "Server started"}, nil
}

// httpHandleRequest maneja las peticiones HTTP entrantes
func (e *Evaluator) httpHandleRequest(w http.ResponseWriter, r *http.Request) {
	if e.httpHandler == nil {
		http.Error(w, "No handler", http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Crear mapa de petición
	reqMap := &MapObject{Pairs: make(map[string]Value)}
	reqMap.Pairs["method"] = &String{Value: r.Method}
	reqMap.Pairs["url"] = &String{Value: r.URL.String()}
	reqMap.Pairs["body"] = &String{Value: string(body)}

	headersMap := &MapObject{Pairs: make(map[string]Value)}
	for key, values := range r.Header {
		headersMap.Pairs[key] = &String{Value: strings.Join(values, ", ")}
	}
	reqMap.Pairs["headers"] = headersMap

	// Llamar al handler de Zylo
	args := []Value{reqMap}
	result, err := e.callZyloFunction(e.httpHandler, args)
	if err != nil {
		http.Error(w, fmt.Sprintf("Handler error: %v", err), http.StatusInternalServerError)
		return
	}

	// Procesar respuesta
	switch res := result.(type) {
	case *MapObject:
		status := 200
		body := ""
		headers := make(map[string]string)

		if statusVal, ok := res.Pairs["status"].(*Integer); ok {
			status = int(statusVal.Value)
		}
		if bodyVal, ok := res.Pairs["body"].(*String); ok {
			body = bodyVal.Value
		}
		if headersVal, ok := res.Pairs["headers"].(*MapObject); ok {
			for k, v := range headersVal.Pairs {
				if s, ok := v.(*String); ok {
					headers[k] = s.Value
				}
			}
		}

		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	case *String:
		w.WriteHeader(200)
		w.Write([]byte(res.Value))
	default:
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("%v", res)))
	}
}

// mapToGoMap convierte un Map de Zylo a map[string]interface{}
func (e *Evaluator) mapToGoMap(m *MapObject) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m.Pairs {
		result[k] = e.valueToInterface(v)
	}
	return result
}

// listToGoSlice convierte un List de Zylo a []interface{}
func (e *Evaluator) listToGoSlice(l *List) []interface{} {
	result := make([]interface{}, len(l.Items))
	for i, v := range l.Items {
		result[i] = e.valueToInterface(v)
	}
	return result
}

// valueToInterface convierte un Value de Zylo a interface{}
func (e *Evaluator) valueToInterface(v Value) interface{} {
	switch val := v.(type) {
	case *String:
		return val.Value
	case *Integer:
		return val.Value
	case *Float:
		return val.Value
	case *Boolean:
		return val.Value
	case *MapObject:
		return e.mapToGoMap(val)
	case *List:
		return e.listToGoSlice(val)
	case *Null:
		return nil
	default:
		return val
	}
}

// httpGetAsync realiza una petición GET asíncrona
func (e *Evaluator) httpGetAsync(url string, headers map[string]string, timeout int) *Future {
	future := &Future{
		Result: make(chan ZyloObject, 1),
		value:  nil,
		once:   false,
	}
	go func() {
		result, _ := e.httpGet(url, headers, timeout)
		future.Result <- result.(ZyloObject)
	}()
	return future
}

// httpPostJSONAsync realiza una petición POST JSON asíncrona
func (e *Evaluator) httpPostJSONAsync(url string, data Value, headers map[string]string, timeout int) *Future {
	future := &Future{
		Result: make(chan ZyloObject, 1),
		value:  nil,
		once:   false,
	}
	go func() {
		result, _ := e.httpPostJSON(url, data, headers, timeout)
		future.Result <- result.(ZyloObject)
	}()
	return future
}
