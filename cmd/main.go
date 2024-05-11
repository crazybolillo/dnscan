package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/crazybolillo/dnscan/internal/export"
	"github.com/crazybolillo/dnscan/internal/feed"
	"github.com/crazybolillo/dnscan/internal/scan"
)

var version = "dev"

const usage = `dnscan is a tool for scanning DNS domains.

Usage:
	dnscan [options] <domain>
`

func main() {
	os.Exit(run(context.Background()))
}

func printUsage() {
	fmt.Print(usage)
	fmt.Println()
	flag.PrintDefaults()
	fmt.Println()
}

func run(ctx context.Context) int {
	versionFlag := flag.Bool("version", false, "Print the program's version and exit.")
	helpFlag := flag.Bool("help", false, "Print this help and exit.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return 0
	}

	if *helpFlag {
		printUsage()
		return 0
	}

	args := os.Args[1:]
	if len(args) != 1 {
		printUsage()
		return 2
	}

	feeder := feed.New()
	scanner := scan.New(args[0], feeder.Output)

	go feeder.Start(ctx)
	go scanner.Start(ctx)

	export.Start(ctx, scanner.Output, os.Stdout)

	return 0
}
