// internal/tests/collections_test.go

package tests

import (
	"testing"
)

func TestListOperations(t *testing.T) {
	tests := []TestCase{
		{
			Name: "List creation and access",
			Code: `
func main() {
    numeros := [1, 2, 3, 4, 5]
    show.log(numeros[0])
    show.log(numeros[2])
}`,
			ExpectedOutput: "1\n3",
			ShouldCompile: true,
		},
		{
			Name: "List with different types",
			Code: `
func main() {
    mixed := [1, "hello", 3.14, true]
    show.log(mixed[1], mixed[3])
}`,
			ExpectedOutput: "hello true",
			ShouldCompile: true,
		},
		{
			Name: "Empty list",
			Code: `
func main() {
    empty := []
    show.log("length:", len(empty))
}`,
			ExpectedOutput: "length: 0",
			ShouldCompile: true,
		},
		{
			Name: "List operations - len",
			Code: `
func main() {
    lista := [1, 2, 3, 4]
    show.log(len(lista))
}`,
			ExpectedOutput: "4",
			ShouldCompile: true,
		},
		{
			Name: "Nested lists",
			Code: `
func main() {
    nested := [[1, 2], [3, 4], [5, 6]]
    show.log(nested[1][0], nested[2][1])
}`,
			ExpectedOutput: "3 6",
			ShouldCompile: true,
		},
		{
			Name: "List indexing with variables",
			Code: `
func main() {
    arr := ["a", "b", "c", "d"]
    index := 2
    show.log(arr[index])
}`,
			ExpectedOutput: "c",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestListSlicing(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Basic list slicing",
			Code: `
func main() {
    numeros := [1, 2, 3, 4, 5]
    parte := numeros[1:4]
    show.log(parte)
}`,
			ExpectedOutput: "[2, 3, 4]",
			ShouldCompile: true,
		},
		{
			Name: "Slice from beginning",
			Code: `
func main() {
    lista := [10, 20, 30, 40, 50]
    primeros := lista[0:3]
    show.log(primeros)
}`,
			ExpectedOutput: "[10, 20, 30]",
			ShouldCompile: true,
		},
		{
			Name: "Slice to end",
			Code: `
func main() {
    data := ["a", "b", "c", "d", "e"]
    ultimos := data[2:]
    show.log(ultimos)
}`,
			ExpectedOutput: "[c, d, e]",
			ShouldCompile: true,
		},
		{
			Name: "Single element slice",
			Code: `
func main() {
    items := [1, 2, 3, 4, 5]
    single := items[1:2]
    show.log(single)
}`,
			ExpectedOutput: "[2]",
			ShouldCompile: true,
		},
		{
			Name: "Slice with variables",
			Code: `
func main() {
    arr := [10, 20, 30, 40, 50, 60, 70]
    start := 2
    end := 5
    sliced := arr[start:end]
    show.log(sliced)
}`,
			ExpectedOutput: "[30, 40, 50]",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestNegativeIndexing(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Negative index access",
			Code: `
func main() {
    lista := [1, 2, 3, 4]
    show.log(lista[-1])
}`,
			ExpectedOutput: "4",
			ShouldCompile: true,
		},
		{
			Name: "Negative index with larger list",
			Code: `
func main() {
    arr := ["a", "b", "c", "d", "e", "f"]
    show.log(arr[-2], arr[-1])
}`,
			ExpectedOutput: "e f",
			ShouldCompile: true,
		},
		{
			Name: "Negative slicing should fail syntax",
			Code: `
func main() {
    lista := [1, 2, 3]
    parte := lista[-2:2]
    show.log(parte)
}`,
			ShouldCompile: false,
		},
	}

	RunTestCases(t, tests)
}

func TestMapOperations(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Map creation and access",
			Code: `
func main() {
    persona := {"nombre": "Ana", "edad": 25}
    show.log(persona["nombre"])
    show.log(persona["edad"])
}`,
			ExpectedOutput: "Ana\n25",
			ShouldCompile: true,
		},
		{
			Name: "Map with different types",
			Code: `
func main() {
    data := {"int": 42, "float": 3.14, "string": "hello", "bool": true}
    show.log(data["float"], data["bool"])
}`,
			ExpectedOutput: "3.14 true",
			ShouldCompile: true,
		},
		{
			Name: "Empty map",
			Code: `
func main() {
    empty := {}
    show.log("empty map")
}`,
			ExpectedOutput: "empty map",
			ShouldCompile: true,
		},
		{
			Name: "Map with non-existent key",
			Code: `
func main() {
    mapa := {"a": 1}
    valor := mapa["b"]
    show.log(valor)
}`,
			ExpectedOutput: "null",
			ShouldCompile: true,
		},
		{
			Name: "Map with expression keys",
			Code: `
func main() {
    x := "key"
    data := {(x + "1"): 100, (x + "2"): 200}
    show.log(data["key1"], data["key2"])
}`,
			ExpectedOutput: "100 200",
			ShouldCompile: true,
		},
		{
			Name: "Nested maps",
			Code: `
func main() {
    nested := {
        "user": {"name": "John", "age": 30},
        "active": true
    }
    show.log(nested["user"]["name"])
}`,
			ExpectedOutput: "John",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestComplexCollections(t *testing.T) {
	tests := []TestCase{
		{
			Name: "List of maps",
			Code: `
func main() {
    users := [
        {"name": "Alice", "score": 95},
        {"name": "Bob", "score": 87},
        {"name": "Charlie", "score": 92}
    ]
    show.log(users[0]["name"], users[2]["score"])
}`,
			ExpectedOutput: "Alice 92",
			ShouldCompile: true,
		},
		{
			Name: "Map of lists",
			Code: `
func main() {
    data := {
        "numbers": [1, 2, 3, 4, 5],
        "letters": ["a", "b", "c"],
        "mixed": [42, "hello", true]
    }
    show.log(data["numbers"][2], data["letters"][1])
}`,
			ExpectedOutput: "3 b",
			ShouldCompile: true,
		},
		{
			Name: "Deep nesting",
			Code: `
func main() {
    complex := {
        "matrix": [
            [1, 2, 3],
            [4, 5, 6],
            [7, 8, 9]
        ],
        "metadata": {"size": 3, "type": "integer"}
    }
    show.log(complex["matrix"][1][2], complex["metadata"]["type"])
}`,
			ExpectedOutput: "6 integer",
			ShouldCompile: true,
		},
		{
			Name: "Collection iteration preparation",
			Code: `
func main() {
    data := [10, 20, 30, 40, 50]
    sum := 0
    for i in data {
        sum = sum + i
    }
    show.log(sum)
}`,
			ExpectedOutput: "150",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}

func TestCollectionErrors(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Index out of bounds",
			Code: `
func main() {
    lista := [1, 2, 3]
    elemento := lista[10]
}`,
			ShouldCompile: true, // Runtime error, should compile
		},
		{
			Name: "Negative index out of bounds",
			Code: `
func main() {
    lista := [1, 2, 3]
    elemento := lista[-10]
}`,
			ShouldCompile: true, // Runtime error, should compile
		},
		{
			Name: "Invalid slice bounds",
			Code: `
func main() {
    lista := [1, 2, 3]
    parte := lista[3:10]
}`,
			ExpectedOutput: "[",
			ShouldCompile: true,
		},
		{
			Name: "Access non-existent map key",
			Code: `
func main() {
    mapa := {"a": 1}
    val := mapa["nonexistent"]
    show.log(val)
}`,
			ExpectedOutput: "null",
			ShouldCompile: true,
		},
		{
			Name: "Type mismatch in collection",
			Code: `
func main() {
    mixed := [1, "2", 3]
    result := mixed[0] + mixed[1]
}`,
			ShouldCompile: false,
		},
	}

	RunTestCases(t, tests)
}

func TestCollectionFunctions(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Length function on lists",
			Code: `
func main() {
    short := [1, 2, 3]
    long := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    show.log("short:", len(short), "long:", len(long))
}`,
			ExpectedOutput: "short: 3 long: 10",
			ShouldCompile: true,
		},
		{
			Name: "Length function on maps",
			Code: `
func main() {
    map1 := {"a": 1, "b": 2}
    map2 := {}
    show.log("map1:", len(map1), "map2:", len(map2))
}`,
			ExpectedOutput: "map1: 2 map2: 0",
			ShouldCompile: true,
		},
		{
			Name: "Processing collections with functions",
			Code: `
func process_list(data) {
    total := 0
    for item in data {
        total = total + item
    }
    return total
}
func main() {
    nums := [10, 20, 30, 40, 50]
    result := process_list(nums)
    show.log(result)
}`,
			ExpectedOutput: "150",
			ShouldCompile: true,
		},
	}

	RunTestCases(t, tests)
}
