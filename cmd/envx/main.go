package main

import (
	"fmt"
	"os"

	"github.com/posit-dev/envx"
)

func main() {
	if err := envx.Run(os.Args, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		os.Exit(1)
	}
}
