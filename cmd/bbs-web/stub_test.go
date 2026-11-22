//go:build !(js && wasm)

package main

import "testing"

func TestStub(t *testing.T) {
	t.Parallel()

	// Simple sanity test for non-wasm stub package compilation.
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
}
