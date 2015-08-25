package two

import (
	"fmt"
	"github.com/mhrabovcin/golang-challenge/ch5-unexport/scanned/one"
)

type myIface interface {
	one.ExportedInterface
}

type myType struct {
	val bool
}

func (mt myType) String() string {
	return "myType String()"
}

func main() {

	fmt.Print(one.ExportedFunction())
	fmt.Println(one.ExportedStruct{})
	str := one.ExportedStruct{}
	fmt.Println(str.ExportedField)
	str.Method()
	var test one.ExportedType
	fmt.Println(test)

}

func TestComposedInterface(v myIface) {
	fmt.Println(v.String())
}

func Tester(test int) float64 {
	return 0.0
}
