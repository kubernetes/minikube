# gostrgen
Random string generator in Golang.

#Installation
`go get -u github.com/elgs/gostrgen`

# Sample code
```go
package main

import (
	"fmt"
	"github.com/elgs/gostrgen"
)

func main() {

	// possible character sets are:
	// Lower, Upper, Digit, Punct, LowerUpper, LowerDigit, UpperDigit, LowerUpperDigit, All and None.
	// Any of the above can be combine by "|", e.g. LowerUpper is the same as Lower | Upper

	charsToGenerate := 20
	charSet := gostrgen.Lower | gostrgen.Digit
	includes := "[]{}<>" // optionally include some additional letters
	excludes := "Ol"     //exclude big 'O' and small 'l' to avoid confusion with zero and one.

	str, err := gostrgen.RandGen(charsToGenerate, charSet, includes, excludes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str) // zxh9[pvoxbaup32b7s0d
}
```