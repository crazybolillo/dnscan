package main

import (
	"context"
	"fmt"
	"os"
)

const usage = `dnscan is a tool for scanning DNS domains.

Usage:
	dnscan <domain>

`

func main() {
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Print(usage)
		return 2
	}

	return 0
}
