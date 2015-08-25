package one

var (
	// Test unexported var
	unexportedVar bool = true

	// Test exported var
	ExportedVar bool = true
)

const (
	ExportedConst = "ExportedConst"

	unexportedConst = "unexportedConst"
)

// Types
type ExportedType string

type unexportedType string

type ExportedStruct struct {
	ExportedField   bool
	unexportedField bool
	UnusedField		bool
}

type UnusedExportedStruct struct {
	Field bool
}

func (inc ExportedStruct) Method() {

}

func (inc *ExportedStruct) UnusedMethod(s string) {

}

// Interfaces
type ExportedInterface interface {
	String() string
}

type unexportedInterface interface {
	String() string
}

// Funcs

func ExportedFunction() string {
	return "string"
}

func unexportedFunction() string {
	return "privateFuncName"
}
