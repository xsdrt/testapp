package main

import (
	"fmt"

	"github.com/xsdrt/hispeed2"
)

func main() {
	result := hispeed2.TestFunc(1, 1)
	fmt.Println(result)

	result = hispeed2.TestFunc2(2, 1)
	fmt.Println(result)

	result = hispeed2.TestFunc3(2, 2)
	fmt.Println(result)

}
