//go:build !depth1

package cmd

import (
	"testing"
)

func BenchmarkRunCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := use_commit_times("../")
		if err != nil {
			b.Fatalf("failed test %#v", err)
		}
	}
}
