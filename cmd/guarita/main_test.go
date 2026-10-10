package main

import (
	"bytes"
	"testing"
)

func TestRunWritesStartupMessageToProvidedWriter(t *testing.T) {
	var stderr bytes.Buffer

	if err := run(&stderr); err != nil {
		t.Fatalf("run() returned an error: %v", err)
	}

	if got, want := stderr.String(), startupMessage+"\n"; got != want {
		t.Fatalf("run() wrote %q; want %q", got, want)
	}
}
