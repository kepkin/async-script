package async_script_test

import (
	"fmt"
	"strings"

	. "github.com/kepkin/async-script"
)

func ExampleTransformer() {
	var res string
	MustRun(
		Exec("echo one two three"),
		Replace("three", "one"),
		Filter("two"),
		ToString(&res),
	)

	fmt.Print(res)
	// Output: one one
}

func Example() {
	packagesForTest := ""
	MustRun(
		Exec("go list ./..."),
		ToString(&packagesForTest),
	)

	packagesForCover := strings.ReplaceAll(packagesForTest, "\n", ",")

	MustRun(
		Exec("go test %v -coverpkg=%v -coverprofile=./coverage.out", packagesForTest, packagesForCover),
		Tee("go-test.log"),
	)
}
