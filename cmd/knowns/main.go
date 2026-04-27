package main

import (
	"os"

	"github.com/datran93/knowns/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
