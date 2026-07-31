package main

import (
	"fmt"
	"os"

	"github.com/marcuwynu23/dataserve/cmd/dataserve/serve"
)

func main() {
	if err := serve.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
