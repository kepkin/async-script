package main

import (
	"fmt"

	. "github.com/kepkin/async-script"
)

func main() {

	var res string
	MustRun(
		Exec("echo one two three"),
		Replace("three", "one"),
		Filter("two"),
		ToString(&res),
	)

	fmt.Print(res)
}
