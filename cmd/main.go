package main

import (
	"fmt"

	obfuscation "github.com/cszichao/image-obfuscation"
)

func main() {
	fmt.Println(obfuscation.Obfuscate(
		"/Users/zichao/Desktop/images/1.jpg",
		"/Users/zichao/Desktop/images/2.jpg"))
}
