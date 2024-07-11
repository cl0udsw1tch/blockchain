package main

import (
	"fmt"

	. "github.com/terium-project/terium/internal"
	"github.com/terium-project/terium/internal/t_error"
	"honnef.co/go/tools/config"
)

// Define an interface with a method returning interface{}
type Stringer interface {
    String() interface{}
}

// Define a type that implements the interface
type MyType struct {
    value string
}

// Implement the String method to return interface{}
func (m MyType) String() interface{} {
    return m.value
}

// Function that expects a slice of the interface type
func printValues(arr []Stringer) {
    for _, v := range arr {
        value := v.String()
        // Handle the returned value using type assertion
        switch v := value.(type) {
        case string:
            fmt.Println("String:", v)
        default:
            fmt.Println("Unknown type")
        }
    }
}

func main() {

	T_DirCtx = DirCtx{}
	if err := T_DirCtx.Config(); err != nil {
		t_error.LogErr(err)
	}


    // Create a slice of MyType
    mySlice := []MyType{{"hello"}, {"world"}}
    
    // Convert the slice of MyType to a slice of Stringer
    stringerSlice := make([]Stringer, len(mySlice))
    for i, v := range mySlice {
        stringerSlice[i] = v
    }
    
    // Pass the converted slice to the function
    printValues(stringerSlice)
}
