package main

import as "github.com/kepkin/async-script"


func main() {

	as.Run(
		as.Exec("bash print_data.sh"),
		as.Watch(5),
	)

}