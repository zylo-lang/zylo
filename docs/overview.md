# Zylo Programming Language - Documentación Técnica Completa

## 📋 Índice

- [Arquitectura](#arquitectura)
- [Sintaxis Completa](#sintaxis-completa)
- [Tipos y Type System](#tipos-y-type-system)
- [Estructuras de Datos](#estructuras-de-datos)
- [Control de Flujo](#control-de-flujo)
- [Functions](#functions)
- [Classes y OOP](#classes-y-oop)
- [Error Handling](#error-handling)
- [Concurrency](#concurrency)
- [Standard Library](#standard-library)
- [Performance](#performance)
- [Debugging](#debugging)

## 🏗️ Arquitectura

### Compilador Pipeline

```
Zylo Source Code (.zylo)
         ↓
    1. Lexer/Tokenization
         ↓
    2. Parser/Syntax Analysis (AST)
         ↓
    3. Semantic Analysis (Symbol Table)
         ↓
    4. Code Generation (Go Source)
         ↓
    5. Go Compilation (Binary)
         ↓
Executable Binary (.exe)
```

### Componentes Principales

#### internal/lexer - Tokenización
- **Responsabilidad**: Convertir código fuente en tokens
- **Tokens Soportados**: Keywords, identifiers, literals, operators
- **Unicode Support**: UTF-8 completo

#### internal/parser - Análisis Sintáctico
- **Output**: Abstract Syntax Tree (AST)
- **Grammar**: LL(1), predictive parsing
- **Error Recovery**: Panic mode para errores no fatales

#### internal/codegen - Generación de Código
- **Target**: Go 1.21+ source code
- **Strategy**: Template-based code generation
- **Optimization**: Built-in helper functions para operations complejas

#### internal/runtime - Runtime System
- **Built-ins**: 70+ funciones runtime integradas
- **Collections**: Arrays y maps optimizados
- **Type System**: Dynamic dispatch para interface{} types

## 🔤 Sintaxis Completa

### Estructura de Programa

```zylo
// Comentarios de una línea

/*
Comentarios multi-línea
*/

// Imports (si se implementa)
import math
import json

// Declaración de constantes
PI := 3.14159
VERSION := "1.0.0"

// Declaración de variables globales
global_counter := 0

// Funciones
func suma(a, b) {
    return a + b
}

// Función main (entry point)
func main() {
    show.log("Hola Zylo!")
}
```

### Identificadores y Keywords

#### Keywords Reservadas
```
func, if, else, elif, for, while, return, break, continue
try, catch, throw, async, await, spawn
class, this, super, extends, import, export
void, public, private, static, final
true, false, null, in, as, is
var, const, let, type, interface
```

#### Identificadores
- **Formato**: `^[a-zA-Z_][a-zA-Z0-9_]*`
- **Case-sensitive**: `variable` ≠ `Variable`
- **Longitud**: Ilimitada

### Literales

#### Números
```zylo
// Enteros
decimal := 42
hex := 0x2A
binary := 0b101010

// Flotantes
float := 3.14
scientific := 6.022e23

// Notación
positivo := +42
negativo := -15
cero := 0
```

#### Strings
```zylo
// Strings simples
simple := "Hola mundo"

// Multi-línea (si se implementa)
multi := `Esta es una
string multi-línea`

// Interpolación (planeado)
interpolado := "Valor: ${variable}"
```

#### Arrays
```zylo
// Arrays vacíos
vacio := []

// Arrays con elementos
numeros := [1, 2, 3, 4, 5]
strings := ["a", "b", "c"]
mixto := [1, "dos", 3.14, true]

// Arrays anidados
matrix := [
    [1, 2, 3],
    [4, 5, 6],
    [7, 8, 9]
]
```

#### Maps/Objects
```zylo
// Maps vacíos
vacio := {}

// Maps con datos
persona := {
    "nombre": "Ana",
    "edad": 25,
    "activo": true
}

// Maps anidados
complejo := {
    "usuario": {
        "id": 123,
        "perfil": {
            "nombre": "Ana",
            "email": "ana@example.com"
        }
    },
    "config": {
        "tema": "dark",
        "idioma": "es"
    }
}
```

## 📊 Tipos y Type System

### Type Inference Automática

Zylo usa un sistema de **tipo híbrido** con inference automática:

```zylo
// Type inference automática
numero := 42        // → int64
decimal := 3.14     // → float64
texto := "Hola"     // → string
booleano := true    // → bool

// Arrays heterogéneos
mixto := [1, "dos", 3.14]  // → []interface{}

// Arrays homogéneos (optimizados)
enteros := [1, 2, 3]        // → []int64 (inferred)
cadenas := ["a", "b"]       // → []string (inferred)
```

### Type Annotations Explícitas

```zylo
// Variables con tipos explícitos
var edad: int = 25
var nombre: string = "Ana"
var activo: bool = true
var precio: float = 99.99

// Parameter types
func procesar(datos: []int, config: map[string]interface{}): bool {
    // ...
}

// Return types
func calcular(): float {
    return 42.0 * 3.14
}

// Type casting
numero := 42
texto := string(numero)    // "42"
flotante := float64(numero) // 42.0
```

### Null Safety

```zylo
// Null checking
valor := obtenerValor()
if valor == null {
    show.log("Valor es nulo")
    return
}

// Safe navigation (planeado para futuro)
resultado := usuario?.perfil?.nombre ?? "Desconocido"

// Null coalescing
nombre := obtenerNombre() ?? "Invitado"
```

## 🗂️ Estructuras de Datos

### Arrays (Slices dinámicos)

```zylo
// Creación
arr := [1, 2, 3, 4, 5]

// Acceso por índice
primero := arr[0]     // 1
ultimo := arr[4]      // 5
negativo := arr[-1]   // 5 (último elemento)

// Slicing
subarray := arr[1:4]  // [2, 3, 4]
primeros := arr[0:3]  // [1, 2, 3]
ultimos := arr[2:]    // [3, 4, 5]

// Modificación
arr[0] = 99           // [99, 2, 3, 4, 5]

// Operaciones de colección
arr.push(6)           // Agregar al final
arr.push(7, 8)        // Agregar múltiples
elemento := arr.pop() // Remover último, retorna elemento
longitud := len(arr)  // Longitud del array

// Iteración
for elemento in arr {
    show.log(elemento)
}

// Búsqueda
existe := arr.contains(3)       // true/false
indice := arr.indexOf(99)       // 0
filtrado := arr.filter(func(x) { return x > 3 })  // [4, 5]
suma := arr.reduce(func(acc, x) { return acc + x }, 0)  // 123
```

### Maps/Dictionaries

```zylo
// Creación
persona := {
    "nombre": "Ana",
    "edad": 25,
    "profesion": "developer"
}

// Acceso
nombre := persona["nombre"]        // "Ana"
edad := persona["edad"]            // 25

// Acceso seguro (con valor por defecto)
ciudad := persona["ciudad"] ?? "Desconocido"  // "Desconocido"

// Modificación
persona["edad"] = 26                       // Actualizar
persona["ciudad"] = "Madrid"                // Agregar
delete(persona, "profesion")                // Eliminar

// Operaciones
claves := keys(persona)     // ["nombre", "edad", "ciudad"]
valores := values(persona)   // ["Ana", 26, "Madrid"]
longitud := len(persona)     // 3
existe := hasKey(persona, "nombre")  // true

// Iteración
for clave in keys(persona) {
    show.log(clave, "=", persona[clave])
}

// O usando map iteration
for clave, valor in persona {
    show.log(clave, "=", valor)
}
```

### Tipos Avanzados

```zylo
// Sets (conjuntos)
conjunto := set([1, 2, 3, 2, 1])  // {1, 2, 3}
conjunto.add(4)          // {1, 2, 3, 4}
conjunto.remove(2)       // {1, 3, 4}
existe := conjunto.has(3)  // true

// Queues (colas)
cola := queue()
cola.push(1)
cola.push(2)
primero := cola.shift()  // 1 (remueve y retorna)

// Stacks (pilas)
pila := stack()
pila.push("a")
pila.push("b")
tope := pila.pop()      // "b"
```

## 🔀 Control de Flujo

### Condicionales

```zylo
// If simple
if edad >= 18 {
    show.log("Mayor de edad")
}

// If/else
if temperatura > 30 {
    show.log("Hace calor")
} else {
    show.log("Temperatura normal")
}

// If/else if/else
if puntuacion >= 90 {
    show.log("Excelente")
} else if puntuacion >= 70 {
    show.log("Bueno")
} else if puntuacion >= 50 {
    show.log("Aprobado")
} else {
    show.log("Reprobado")
}

// Operador ternario (planeado)
resultado := condicion ? valor_verdadero : valor_falso

// Truthy/falsy evaluation
if usuario {           // Equivalente a usuario != null && usuario != ""
    show.log("Usuario válido")
}

if lista {             // Equivalente a len(lista) > 0
    show.log("Lista no vacía")
}

if numero {            // Equivalente a numero != 0
    show.log("Número no cero")
}
```

### Bucles

```zylo
// While loop
contador := 0
while contador < 10 {
    show.log(contador)
    contador += 1
}

// For loop tradicional
for i := 0; i < 5; i += 1 {
    show.log("Iteración", i)
}

// For-each loop en arrays
frutas := ["manzana", "banana", "pera"]
for fruta in frutas {
    show.log(fruta)
}

// For-each en maps
persona := {"nombre": "Ana", "edad": 25}
for clave, valor in persona {
    show.log(clave, "=", valor)
}

// For-in con índices
for i, fruta in frutas {
    show.log("Índice:", i, "Fruta:", fruta)
}

// Break y continue
for i := 0; i < 100; i += 1 {
    if i == 10 {
        break         // Salir del loop
    }
    if i % 2 == 0 {
        continue      // Saltar a siguiente iteración
    }
    show.log(i)
}

// Nested loops
for i := 0; i < 3; i += 1 {
    for j := 0; j < 3; j += 1 {
        if i == j {
            continue  // Saltar diagonal
        }
        show.log("(", i, ",", j, ")")
    }
}
```

### Switch/Match

```zylo
// Switch statement básico
opcion := "guardar"
switch opcion {
case "guardar":
    show.log("Guardando documento...")
case "abrir":
    show.log("Abriendo documento...")
case "cerrar":
    show.log("Cerrando aplicación...")
default:
    show.log("Opción no reconocida")
}

// Switch con múltiples valores
dia := "lunes"
switch dia {
case "lunes", "martes", "miércoles", "jueves", "viernes":
    show.log("Día laboral")
case "sábado", "domingo":
    show.log("Fin de semana")
default:
    show.log("Día inválido")
}

// Match con expresiones complejas (planeado para futuro)
resultado := match valor {
    0 => "cero"
    1..10 => "número pequeño"
    11..100 => "número mediano"
    _ => "número grande"
}
```

## 🔧 Functions

### Declaración de Functions

```zylo
// Función simple sin parámetros
func saludar() {
    show.log("¡Hola!")
}

// Función con parámetros (tipos inferidos)
func suma(a, b) {
    return a + b
}

// Función con tipos explícitos
func dividir(dividendo: float, divisor: float): float {
    if divisor == 0 {
        throw "División por cero"
    }
    return dividendo / divisor
}

// Función con parámetros opcionales (planeado)
func configurar(host: string, puerto: int = 8080) {
    show.log("Conectando a", host, "puerto", puerto)
}

// Función con múltiples return values
func dividirConResto(dividendo, divisor) {
    cociente := dividendo / divisor
    resto := dividendo % divisor
    return cociente, resto
}

// Uso
resultado, residuo := dividirConResto(17, 5)  // 3, 2
```

### Funciones Anónimas y Closures

```zylo
// Función anónima asignada a variable
multiplicar := func(x, y) { return x * y }

// Uso
resultado := multiplicar(5, 3)  // 15

// Función anónima en línea
numeros := [1, 2, 3, 4, 5]
pares := filter(numeros, func(n) { return n % 2 == 0 })  // [2, 4]

// Closures
func crearContador() {
    contador := 0
    return func() {
        contador += 1
        return contador
    }
}

incrementar := crearContador()
show.log(incrementar())  // 1
show.log(incrementar())  // 2
show.log(incrementar())  // 3
```

### Funciones de Orden Superior

```zylo
// Función que recibe función como parámetro
func aplicarOperacion(numeros: []int, operacion: func(int): int): []int {
    resultado := []
    for numero in numeros {
        resultado.push(operacion(numero))
    }
    return resultado
}

// Funciones callback
func procesarUsuario(datos, callback: func(user): user) {
    usuario := parsearUsuario(datos)
    if callback != null {
        usuario = callback(usuario)
    }
    return usuario
}

// Uso
numeros := [1, 2, 3, 4, 5]

// Doble cada número
doblados := aplicarOperacion(numeros, func(n) { return n * 2 })
// [2, 4, 6, 8, 10]

// Elevar al cuadrado
cuadrados := aplicarOperacion(numeros, func(n) { return n * n })
// [1, 4, 9, 16, 25]
```

### Functions Especiales

```zylo
// Constructor (para clases)
class Persona {
    func init(nombre: string, edad: int) {
        this.nombre = nombre
        this.edad = edad
    }
}

// Getter/Setter
class Configuracion {
    private datos := {}

    func get(clave: string) {
        return datos[clave]
    }

    func set(clave: string, valor) {
        datos[clave] = valor
    }
}

// Static methods
class Utilidades {
    static func max(a, b) {
        if a > b {
            return a
        }
        return b
    }
}

// Uso
maximo := Utilidades.max(10, 20)  // 20
```

## 🏗️ Classes y OOP

### Declaración de Clases

```zylo
// Clase básica
class Persona {
    // Atributos
    private nombre: string
    private edad: int

    // Constructor
    func init(nombre: string, edad: int) {
        this.nombre = nombre
        this.edad = edad
    }

    // Getter
    func getNombre(): string {
        return this.nombre
    }

    // Setter
    func setEdad(nuevaEdad: int) {
        if nuevaEdad >= 0 {
            this.edad = nuevaEdad
        }
    }

    // Método de instancia
    func saludar() {
        show.log("Hola, soy", this.nombre, "y tengo", this.edad, "años")
    }

    // Método de clase (static)
    static func crearPorDefecto(): Persona {
        return Persona("Invitado", 25)
    }
}

// Herencia
class Empleado extends Persona {
    private salario: float

    func init(nombre: string, edad: int, salario: float) {
        super.init(nombre, edad)  // Llamar constructor padre
        this.salario = salario
    }

    func getSalario(): float {
        return this.salario
    }

    // Override de método
    func saludar() {
        show.log("Soy empleado", this.nombre,
                "con salario de", this.salario)
    }

    func calcularBono(): float {
        return this.salario * 0.1
    }
}

// Uso
persona := Persona("Ana", 30)
persona.saludar()

empleado := Empleado("Carlos", 35, 50000.0)
empleado.saludar()  // Salida del método overrideado

// Creación con método static
invitado := Empleado.crearPorDefecto()
```

### Polimorfismo

```zylo
// Interfaz implícita (duck typing)
func procesarEntidad(entidad) {
    // Si tiene método 'saludar', lo llamamos
    if entidad.saludar != null {
        entidad.saludar()
    }

    // Si tiene método 'getSalario', es un empleado
    if entidad.getSalario != null {
        salario := entidad.getSalario()
        show.log("Salario:", salario)
    }
}

persona := Persona("Ana", 30)
empleado := Empleado("Carlos", 35, 50000.0)

procesarEntidad(persona)   // Solo saluda
procesarEntidad(empleado)  // Saluda y muestra salario
```

### Traits/Mixins (planeado)

```zylo
// Sintaxis propuesta para futuro
trait Serializable {
    func toJSON() {
        return JSON.stringify(this)
    }

    func fromJSON(jsonString) {
        parsed := JSON.parse(jsonString)
        for clave in keys(parsed) {
            this[clave] = parsed[clave]
        }
        return this
    }
}

class Usuario with Serializable {
    // ... atributos y métodos ...

    func saveToFile(filename) {
        json := this.toJSON()
        writeFile(filename, json)
    }

    func loadFromFile(filename) {
        json := readFile(filename)
        return this.fromJSON(json)
    }
}
```

## ⚠️ Error Handling

### Result Type

```zylo
// Result type integrado
func dividir(a, b): Result {
    if b == 0 {
        return Result{Error: "División por cero"}
    }
    return Result{Value: a / b}
}

func main() {
    // Uso básico
    result1 := dividir(10, 2)
    result2 := dividir(5, 0)

    if result1.isOk() {
        show.log("Resultado:", result1.unwrap())  // 5
    }

    if result2.isErr() {
        show.log("Error:", result2.Error)  // "División por cero"
    }
}

// Method chaining con Result
func validateEmail(email: string): Result {
    if email == "" {
        return Result{Error: "Email vacío"}
    }
    if !email.contains("@") {
        return Result{Error: "Email inválido"}
    }
    return Result{Value: email}
}

func processUser(email: string) {
    result := validateEmail(email)
               .map(func(e) { return toLower(e) })
               .map(func(e) { return "usuario: " + e })

    if result.isOk() {
        show.log("Usuario procesado:", result.unwrap())
    } else {
        show.log("Error de validación:", result.Error)
    }
}
```

### Try/Catch Blocks

```zylo
func procesarArchivo(filename: string) {
    try {
        contenido := readFile(filename)
        lineas := split(contenido, "\n")

        if len(lineas) == 0 {
            throw "Archivo vacío"
        }

        // Procesar primera línea
        primera := lineas[0]
        datos := JSON.parse(primera)  // Puede lanzar error

        show.log("Procesamiento exitoso")

    } catch error {
        // Manejar diferentes tipos de error
        if error.contains("File not found") {
            show.log("Archivo no encontrado:", filename)
        } else if error.contains("JSON parse error") {
            show.log("Error de formato JSON")
        } else {
            show.log("Error desconocido:", error)
        }
    } finally {
        // Siempre se ejecuta
        show.log("Operación completada")
    }
}
```

### Error Propagation

```zylo
// Propagación de errores
func obtenerUsuario(id: string) {
    usuario := database.findById(id)
    if usuario == null {
        throw "Usuario no encontrado: " + id
    }
    return usuario
}

func enviarEmail(usuario) {
    if usuario.email == null {
        throw "Usuario sin email: " + usuario.id
    }

    result := mailer.send(usuario.email, "Bienvenida")
    if result.isErr() {
        throw "Error enviando email: " + result.Error
    }
}

func registrarUsuario(datos): Result {
    try {
        usuario := database.create(datos)
        enviarEmail(usuario)
        return Result{Value: usuario}
    } catch error {
        return Result{Error: "Registro fallido: " + error}
    }
}
```

## 🔄 Concurrency

### Goroutines (async/await)

```zylo
// Función asíncrona
async func descargarArchivo(url: string): Result {
    try {
        respuesta := await fetch(url)
        contenido := await respuesta.text()
        return Result{Value: contenido}
    } catch error {
        return Result{Error: "Download failed: " + error}
    }
}

// Uso básico
func descargarArchivos() {
    // Ejecutar en paralelo
    file1 := spawn(descargarArchivo("http://example.com/file1.txt"))
    file2 := spawn(descargarArchivo("http://example.com/file2.txt"))

    // Esperar resultados
    result1 := await(file1)
    result2 := await(file2)

    if result1.isOk() && result2.isOk() {
        show.log("Descargas completadas exitosamente")
    }
}

// Async functions en clases
class DataLoader {
    async func loadFromAPI(endpoint: string): []interface{} {
        respuesta := await fetch(endpoint)
        return await respuesta.json()
    }

    async func loadMultiple(endpoints: []string): map[string]interface{} {
        result := {}

        // Iniciar todas las descargas en paralelo
        futures := {}
        for endpoint in endpoints {
            futures[endpoint] = spawn(this.loadFromAPI(endpoint))
        }

        // Esperar todas y recolectar resultados
        for endpoint, future in futures {
            result[endpoint] = await(future)
        }

        return result
    }
}
```

### Channels

```zylo
// Channels para comunicación entre goroutines
func producer(ch: Channel) {
    for i := 0; i < 10; i += 1 {
        ch.send(i * 2)  // Enviar dato al channel
        sleep(100)      // Esperar 100ms
    }
    ch.close()  // Cerrar channel
}

func consumer(ch: Channel) {
    total := 0
    while dato := ch.receive() {  // nil cuando se cierra
        total += dato
        show.log("Recibido:", dato, "Total:", total)
    }
    show.log("Consumidor terminado. Total:", total)
}

func main() {
    // Crear channel
    ch := Channel.make(5)  // Buffer de 5 elementos

    // Iniciar producer y consumer
    spawn(producer, ch)
    spawn(consumer, ch)

    // Esperar un tiempo
    sleep(2000)
}
```

### Atomic Operations

```zylo
class Counter {
    private valor := 0
    private lock := Mutex.new()

    func increment() {
        lock.lock()
        defer lock.unlock()
        valor += 1
    }

    func getValue(): int {
        lock.lock()
        defer lock.unlock()
        return valor
    }
}

func testCounter() {
    counter := Counter.new()

    // Incrementar en paralelo
    for i := 0; i < 1000; i += 1 {
        spawn(counter.increment)
    }

    sleep(1000)  // Esperar que terminen

    show.log("Valor final:", counter.getValue())  // Debería ser 1000
}
```

## 📚 Standard Library

### math.zylo

```zylo
import math

func main() {
    // Constantes
    show.log("Pi:", math.pi)           // 3.141592653589793
    show.log("E:", math.e)             // 2.718281828459045
    show.log("Sqrt(2):", math.sqrt2)   // 1.4142135623730951

    // Funciones trigonométricas
    seno := math.sin(math.pi / 2)      // 1.0
    coseno := math.cos(0)              // 1.0
    tangente := math.tan(math.pi / 4)  // 0.9999999999999999

    // Funciones exponenciales/logaritmicas
    exp := math.exp(1)                 // 2.718281828459045
    log := math.log(math.e)            // 1.0
    pow := math.pow(2, 3)              // 8.0

    // Funciones de redondeo
    floor := math.floor(3.7)           // 3.0
    ceil := math.ceil(3.2)             // 4.0
    round := math.round(3.5)           // 4.0 (banker's rounding)

    // Funciones misceláneas
    abs := math.abs(-42)               // 42
    min := math.min(10, 20)            // 10
    max := math.max(10, 20)            // 20
    sign := math.sign(-5)              // -1
}
```

### string.zylo

```zylo
import string

func main() {
    texto := "Hola Mundo Cruel"

    // Información básica
    longitud := string.len(texto)      // 16
    empty := string.isEmpty("")        // true
    blank := string.isBlank("   ")     // true

    // Conversión de caso
    upper := string.upper(texto)       // "HOLA MUNDO CRUEL"
    lower := string.lower(texto)       // "hola mundo cruel"
    title := string.title(texto)       // "Hola Mundo Cruel"

    // Búsqueda
    contains := string.contains(texto, "Mundo")  // true
    startsWith := string.startsWith(texto, "Hola")  // true
    endsWith := string.endsWith(texto, "!")     // false
    index := string.indexOf(texto, "Mundo")     // 5
    lastIndex := string.lastIndexOf(texto, "o") // 9

    // Modificación
    replace := string.replace(texto, "Cruel", "Hermoso")  // "Hola Mundo Hermoso"
    trim := string.trim("  hola  ")           // "hola"
    trimLeft := string.trimLeft("  hola  ")   // "hola  "
    trimRight := string.trimRight("  hola  ") // "  hola"

    // Splitting y joining
    parts := string.split(texto, " ")         // ["Hola", "Mundo", "Cruel"]
    joined := string.join(parts, "-")         // "Hola-Mundo-Cruel"

    // Substrings
    substring := string.substring(texto, 5, 10)  // "Mundo"
    left := string.left(texto, 4)              // "Hola"
    right := string.right(texto, 5)            // "Cruel"

    // Formateo
    repeat := string.repeat("ha", 3)          // "hahaha"
    padLeft := string.padLeft("42", 5, "0")   // "00042"
    padRight := string.padRight("42", 5, " ") // "42   "

    // Validación
    isAlpha := string.isAlpha("Hola")         // true
    isNumeric := string.isNumeric("123")      // true
    isAlphanumeric := string.isAlphanumeric("Hola123")  // true
    isEmail := string.isEmail("user@example.com")  // true (dependiendo de implementación)

    // Conversión
    toInt := string.toInt("42")              // 42
    toFloat := string.toFloat("3.14")        // 3.14
    toBool := string.toBool("true")          // true

    show.log("Procesamiento de strings completado")
}
```

### json.zylo

```zylo
import json

func main() {
    // Parsing JSON
    jsonString := '{"nombre": "Ana", "edad": 25, "activo": true, "etiquetas": ["developer", "web"]}'

    try {
        parsed := json.parse(jsonString)

        nombre := parsed["nombre"]       // "Ana"
        edad := parsed["edad"]           // 25
        activo := parsed["activo"]       // true
        etiquetas := parsed["etiquetas"] // ["developer", "web"]

        show.log("Nombre:", nombre)
        show.log("Edad:", edad)
        show.log("Activo:", activo)
        show.log("Etiquetas:", etiquetas)

    } catch error {
        show.log("Error parseando JSON:", error)
    }

    // Creando JSON
    persona := {
        "nombre": "Carlos",
        "edad": 30,
        "profesion": "engineer",
        "habilidades": ["Go", "Python", "JavaScript"],
        "contacto": {
            "email": "carlos@example.com",
            "telefono": "+1234567890"
        }
    }

    try {
        jsonOutput := json.stringify(persona, 2)  // 2 espacios de indentación

        show.log("JSON generado:")
        show.log(jsonOutput)

        // Opcional: escribir a archivo
        // io.writeFile("persona.json", jsonOutput)

    } catch error {
        show.log("Error generando JSON:", error)
    }

    // Validación de JSON
    invalidJson := '{"nombre": "Test", "incompleto": true'
    isValid := json.isValid(invalidJson)  // false
    show.log("JSON válido:", isValid)

    // Pretty printing
    minified := '{"a":1,"b":2}'
    pretty := json.prettyPrint(minified)
    show.log("Pretty JSON:")
    show.log(pretty)
}
```

### io.zylo

```zylo
import io
import string

func main() {
    filename := "ejemplo.txt"

    // Escribir archivo
    contenido := "Hola, este es un archivo de ejemplo.\nSegunda línea."
    try {
        io.writeFile(filename, contenido)
        show.log("Archivo escrito exitosamente")
    } catch error {
        show.log("Error escribiendo archivo:", error)
    }

    // Leer archivo completo
    try {
        data := io.readFile(filename)
        show.log("Contenido del archivo:")
        show.log(data)
    } catch error {
        show.log("Error leyendo archivo:", error)
    }

    // Leer línea por línea
    try {
        lines := io.readLines(filename)
        show.log("Número de líneas:", len(lines))
        for i, line in lines {
            show.log("Línea", i+1, ":", string.trim(line))
        }
    } catch error {
        show.log("Error leyendo líneas:", error)
    }

    // Leer entrada del usuario
    show.log("¿Cuál es tu nombre?")
    nombre := io.readLine()
    show.log("Hola,", nombre, "!")

    // Escribir/Leer archivo binario (si se implementa)
    // bytes := []byte{72, 101, 108, 108, 111}  // "Hello"
    // io.writeFileBytes("binary.dat", bytes)
    // readBytes := io.readFileBytes("binary.dat")

    // Verificar existencia
    exists := io.fileExists(filename)  // true
    show.log("Archivo existe:", exists)

    // Información de archivo
    if exists {
        size := io.fileSize(filename)
        show.log("Tamaño del archivo:", size, "bytes")
    }
}
```

### time.zylo

```zylo
import time

func main() {
    // Fecha y hora actual
    ahora := time.now()
    show.log("Ahora:", ahora)

    // Crear fechas específicas
    fecha := time.date(2024, 1, 15, 10, 30, 45)  // 2024-01-15 10:30:45
    show.log("Fecha específica:", fecha)

    // Parsing de fechas
    parsed := time.parse("2024-01-15", "2006-01-02")
    show.log("Fecha parseada:", parsed)

    // Operaciones con tiempo
    manana := ahora.addDays(1)
    ayer := ahora.addDays(-1)
    proximoMes := ahora.addMonths(1)

    show.log("Mañana:", manana)
    show.log("Ayer:", ayer)
    show.log("Próximo mes:", proximoMes)

    // Diferencia entre fechas
    diferencia := manana.diff(ayer)  // En días
    show.log("Diferencia:", diferencia, "días")

    // Formateo
    formatted := ahora.format("2006-01-02 15:04:05")
    show.log("Formateado:", formatted)

    fechaCorta := ahora.format("02/01/06")
    show.log("Fecha corta:", fechaCorta)

    // Componentes
    year := ahora.year()     // 2024
    month := ahora.month()   // 1 (enero)
    day := ahora.day()       // 15
    hour := ahora.hour()     // 10
    minute := ahora.minute() // 30
    second := ahora.second() // 45
    weekday := ahora.weekday()  // "Monday", "Tuesday", etc.

    show.log("Año:", year, "Mes:", month, "Día:", day)
    show.log("Hora:", hour, minute, second, "Día semana:", weekday)

    // Duraciones
    duracion := time.duration(0, 30, 45)  // 30 minutos 45 segundos
    show.log("Duración:", duracion)

    totalSegundos := duracion.totalSeconds()  // 1845
    show.log("Total segundos:", totalSegundos)

    // Cronómetro
    timer := time.startTimer()
    // ... hacer algo ...
    // Simulamos trabajo
    for i := 0; i < 1000000; i += 1 { }
    elapsed := timer.elapsed()
    show.log("Tiempo transcurrido:", elapsed, "segundos")

    // Sleep/Wait
    show.log("Esperando 1 segundo...")
    time.sleep(1000)  // 1000ms = 1 segundo
    show.log("¡Listo!")
}
```

## ⚡ Performance

### Optimizaciones del Compilador

1. **Type Specialisation**: Arrays homogéneos usan `[]int64` en lugar de `[]interface{}`
2. **Constant Folding**: Operaciones constantes se pre-calculan en compile-time
3. **Dead Code Elimination**: Código unreachable se elimina automáticamente
4. **Function Inlining**: Funciones pequeñas se inlined automáticamente
5. **Loop Optimisation**: Bucles se optimizan para mejor performance

### Guidelines de Performance

```zylo
// ✅ RECOMENDADO
func processLargeArray(data: []int64) {
    // Arrays grandes: pre-asignar capacidad si es posible
    result := make([]int64, len(data))

    // Evitar acceso a propiedades en loops calientes
    length := len(data)  // Cache length outside loop
    for i := 0; i < length; i += 1 {
        if data[i] > 0 {  // Direct access, no len() call
            result[i] = data[i] * 2
        }
    }
    return result
}

// ❌ NO RECOMENDADO
func slowProcessing(data: []interface{}) {
    result := []  // Empty
