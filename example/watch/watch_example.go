package main

import . "github.com/kepkin/async-script"

func main() {

	MustRun(
		Exec("bash print_data.sh"),
		Watch(5),
	)

}
