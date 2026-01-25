//go:build !depth1

package cmd

import (
	"testing"
)

func BenchmarkRunCommand(b *testing.B) {
	b.ResetTimer()
	err := use_commit_times("../")
	if err != nil {
		b.Fatalf("failed test %#v", err)
	}
}
