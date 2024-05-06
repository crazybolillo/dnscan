package feed

import (
	"context"
	"strings"
	"testing"
)

var newTests = map[string]struct {
	values   []string
	expected []string
}{
	"trimmed": {
		values:   []string{"www", "blog", "api"},
		expected: []string{"www", "blog", "api"},
	},
	"untrimmed": {
		values:   []string{"  voip ", "mail  ", "  meter\n"},
		expected: []string{"voip", "mail", "meter"},
	},
	"empty": {
		values:   []string{" ", "me", ""},
		expected: []string{"me"},
	},
}

func TestFeed(t *testing.T) {
	for name, data := range newTests {
		t.Run(name, func(t *testing.T) {
			reader := strings.NewReader(strings.Join(data.values, "\n"))
			f := NewWithReader(reader)
			go f.Start(context.Background())

			missing := make(map[string]bool)
			for _, val := range data.expected {
				missing[val] = true
			}

			for domain := range f.Output {
				delete(missing, domain)
			}

			for missed := range missing {
				t.Errorf("missing %s from output", missed)
			}
		})
	}
}
