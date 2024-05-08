package scan

import (
	"context"
	"fmt"
	"net"
	"testing"
)

type fakeResolver struct {
	resolve func(host string) ([]net.IPAddr, error)
}

func (r *fakeResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	return r.resolve(host)
}

var notFoundTests = map[string]func(host string) ([]net.IPAddr, error){
	"Empty": func(host string) ([]net.IPAddr, error) {
		return []net.IPAddr{}, nil
	},
	"Error": func(host string) ([]net.IPAddr, error) {
		return []net.IPAddr{}, fmt.Errorf("kaboom")
	},
}

func TestFound(t *testing.T) {
	records := map[string][]net.IPAddr{
		"blue.pen.com": {net.IPAddr{IP: net.ParseIP("3.120.5.12")}},
		"red.pen.com": {
			net.IPAddr{IP: net.ParseIP("9.60.5.12")},
			net.IPAddr{IP: net.ParseIP("5.7.8.9")},
		},
	}
	resolve := func(host string) ([]net.IPAddr, error) {
		return records[host], nil
	}

	input := make(chan string, 2)
	input <- "blue"
	input <- "red"
	close(input)

	s := New("pen.com", input)
	s.Resolver = &fakeResolver{resolve: resolve}
	s.Start(context.Background())

	for res := range s.Output {
		ips := records[res.Host]
		if len(ips) != len(res.IPs) {
			t.Errorf("expected %d ip addresses, got %d", len(ips), len(res.IPs))
		}
		delete(records, res.Host)
	}

	for domain := range records {
		t.Errorf("expected %s to be resolved", domain)
	}
}

func TestNotFound(t *testing.T) {
	for name, fn := range notFoundTests {
		t.Run(name, func(t *testing.T) {
			input := make(chan string, 1)
			input <- "hello"
			close(input)

			s := New("kiwi.com", input)
			s.Resolver = &fakeResolver{resolve: fn}
			s.Start(context.Background())

			_, ok := <-s.Output
			if ok {
				t.Errorf("expected channel to be closed")
			}
		})
	}
}
