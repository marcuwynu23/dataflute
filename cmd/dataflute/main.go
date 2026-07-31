package main

import (
	"fmt"
	"os"

	"github.com/marcuwynu23/dataflute/cmd/dataflute/serve"
)

func main() {
	if err := serve.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
