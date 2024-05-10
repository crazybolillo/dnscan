package main

import (
	"context"
	"fmt"
	"os"

	"github.com/crazybolillo/dnscan/internal/export"
	"github.com/crazybolillo/dnscan/internal/feed"
	"github.com/crazybolillo/dnscan/internal/scan"
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

	feeder := feed.New()
	scanner := scan.New(args[0], feeder.Output)

	go feeder.Start(ctx)
	go scanner.Start(ctx)

	export.Start(ctx, scanner.Output, os.Stdout)

	return 0
}
