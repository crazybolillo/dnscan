package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

func setupMetrics(zone string) {
	promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "dnscan_info",
		Help:        "Information about the running dnscan process",
		ConstLabels: prometheus.Labels{"zone": zone},
	}).Set(1)
	http.Handle("/metrics", promhttp.Handler())

	go http.ListenAndServe(":9990", nil)
}

func work(ctx context.Context, zone string) {
	feeder := feed.New()
	scanner := scan.New(zone, feeder.Output)

	go feeder.Start(ctx)
	go scanner.Start(ctx)

	export.Start(ctx, scanner.Output, os.Stdout)
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
	zone := args[0]

	setupMetrics(zone)
	work(workCtx, zone)

	return 0
}
