package export

import (
	"context"
	"fmt"
	"io"

	"github.com/crazybolillo/dnscan/internal/scan"
)

func Start(ctx context.Context, input <-chan scan.Result, writer io.Writer) {
	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-input:
			if !ok {
				return
			}
			msg := fmt.Sprintf("%s: %v\n", result.Host, result.IPs)
			writer.Write([]byte(msg))
		}
	}
}
