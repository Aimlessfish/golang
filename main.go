// Example of using argument flags in Go.
// Get, parse, output parse.

// Example of structured logging also included.

package main

import (
	"flag"
	"fmt"
)

func main() {
	// Get
	flag1 := flag.String("flag1", "", "string")
	flag2 := flag.Int("flag2", 0, "integer")

	// Parse
	flag.Parse()

	// Output
	fmt.Print("flag1 = ", *flag1, " flag2 = ", *flag2)
}
