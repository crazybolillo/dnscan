package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

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
	timeoutFlag := flag.Int("timeout", 0, "Stop scanning after X seconds. Positive numbers enable this feature.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return 0
	}

	if *helpFlag {
		printUsage()
		return 0
	}

	workCtx := ctx
	if *timeoutFlag > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(*timeoutFlag)*time.Second)
		workCtx = timeoutCtx
		defer cancel()
	}

	args := flag.Args()
	if len(args) != 1 {
		printUsage()
		return 2
	}

	feeder := feed.New()
	scanner := scan.New(args[0], feeder.Output)

	go feeder.Start(workCtx)
	go scanner.Start(workCtx)

	export.Start(workCtx, scanner.Output, os.Stdout)

	return 0
}
