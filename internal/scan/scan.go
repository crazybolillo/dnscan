package scan

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var scannedCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "dnscan_scanned_total",
	Help: "The total number of subdomains that have been scanned",
})
var foundCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "dnscan_found_total",
	Help: "The total number of subdomains with at least one IP address found.",
})

type HostResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type Scanner struct {
	channel chan Result
	group   sync.WaitGroup

	Zone     string
	Input    <-chan string
	Output   <-chan Result
	Resolver HostResolver
}

type Result struct {
	Host string
	IPs  []net.IPAddr
}

func New(zone string, input <-chan string) Scanner {
	channel := make(chan Result, runtime.GOMAXPROCS(0))

	return Scanner{
		channel:  channel,
		group:    sync.WaitGroup{},
		Zone:     zone,
		Input:    input,
		Output:   channel,
		Resolver: net.DefaultResolver,
	}
}

func (s *Scanner) Start(ctx context.Context) {
	defer close(s.channel)

	s.group.Add(runtime.GOMAXPROCS(0))
	for range runtime.GOMAXPROCS(0) {
		go scan(ctx, s)
	}
	s.group.Wait()
}

func scan(ctx context.Context, scanner *Scanner) {
	defer scanner.group.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case subdomain, ok := <-scanner.Input:
			if !ok {
				return
			}
			scannedCount.Inc()

			host := fmt.Sprintf("%s.%s", subdomain, scanner.Zone)
			ips, err := scanner.Resolver.LookupIPAddr(ctx, host)
			if err != nil {
				continue
			}
			if len(ips) == 0 {
				continue
			}

			foundCount.Inc()
			scanner.channel <- Result{Host: host, IPs: ips}
		}
	}
}
