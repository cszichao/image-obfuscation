package main

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	obfuscation "github.com/cszichao/image-obfuscation"
)

func main() {
	fmt.Println(obfuscation.Obfuscate(
		"/Users/zichao/Desktop/images/1.png",
		"/Users/zichao/Desktop/images/2.png"))
}
