//go:build integration

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var buffer bytes.Buffer
	expected := map[string]bool{
		"www.test.com: [{1.2.3.4 }]":                                true,
		"mail.test.com: [{5.6.7.8 } {9.10.11.12 }]":                 true,
		"ftp.test.com: [{fd81:fae1:95e6:1a32:3eb:693c:8c8c:db9c }]": true,
	}

	run(context.Background(), []string{"dnscan", "-timeout=1", "-resolver=127.0.0.1:5353", "test.com"}, &buffer)
	for {
		line, err := buffer.ReadString('\n')
		result := strings.TrimSpace(line)
		if err != nil {
			break
		}
		delete(expected, result)
	}

	for missing, _ := range expected {
		t.Errorf("expected to find %s in output", missing)
	}
}
