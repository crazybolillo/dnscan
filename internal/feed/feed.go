package feed

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"io"
	"strings"
)

//go:embed data/subdomains-top1million-110000.txt
var subdomains string

type Feeder struct {
	channel chan string
	reader  io.Reader

	Output <-chan string
}

func (f *Feeder) Start(ctx context.Context) {
	feed(ctx, f.reader, f.channel)
}

func feed(ctx context.Context, reader io.Reader, out chan<- string) {
	defer close(out)

	buffer := bufio.NewReader(reader)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			domain, err := buffer.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				// TODO HANDLE ME
				return
			}
			domain = strings.TrimSpace(domain)
			if domain != "" {
				out <- domain[:]
			}
			if errors.Is(err, io.EOF) {
				return
			}
		}
	}
}

func NewWithReader(reader io.Reader) Feeder {
	channel := make(chan string)

	return Feeder{
		channel: channel,
		reader:  reader,
		Output:  channel,
	}
}

func New() Feeder {
	return NewWithReader(strings.NewReader(subdomains))
}
