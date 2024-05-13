package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
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
	os.Exit(run(context.Background(), os.Args, os.Stdout))
}

func printUsage(flg *flag.FlagSet) {
	fmt.Print(usage)
	fmt.Println()
	flg.PrintDefaults()
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

func work(ctx context.Context, zone string, resolver *net.Resolver, w io.Writer) {
	feeder := feed.New()
	scanner := scan.New(zone, feeder.Output)
	if resolver != nil {
		scanner.Resolver = resolver
	}

	go feeder.Start(ctx)
	go scanner.Start(ctx)

	export.Start(ctx, scanner.Output, w)
}

func run(ctx context.Context, args []string, w io.Writer) int {
	flg := flag.NewFlagSet(args[0], flag.ContinueOnError)

	versionFlag := flg.Bool("version", false, "Print the program's version and exit.")
	helpFlag := flg.Bool("help", false, "Print this help and exit.")
	timeoutFlag := flg.Int("timeout", 0, "Stop scanning after X seconds. Positive numbers enable this feature.")
	resolverFlag := flg.String("resolver", "", "Use the given server to resolve DNS instead. Format: ipaddr:port")

	err := flg.Parse(args[1:])
	if err != nil {
		printUsage(flg)
		return 2
	}

	if *versionFlag {
		fmt.Println(version)
		return 0
	}

	if *helpFlag {
		printUsage(flg)
		return 0
	}

	workCtx := ctx
	if *timeoutFlag > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(*timeoutFlag)*time.Second)
		workCtx = timeoutCtx
		defer cancel()
	}

	var resolver *net.Resolver = net.DefaultResolver
	if len(*resolverFlag) != 0 {
		_, _, err := net.SplitHostPort(*resolverFlag)
		if err != nil {
			fmt.Println(err)
			return 2
		}

		dialer := net.Dialer{}
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, *resolverFlag)
			},
		}
	}

	posArgs := flg.Args()
	if len(posArgs) != 1 {
		printUsage(flg)
		return 2
	}
	zone := posArgs[0]

	setupMetrics(zone)
	work(workCtx, zone, resolver, w)

	return 0
}
