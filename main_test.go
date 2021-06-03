package main

import (
	"flag"
	"testing"
)

var target = flag.String("target", "", "https://host:port of serivce under test")

func TestRandomEndpoint(t *testing.T) {
	if *target == "" {
		t.Skip("missing `target` parameter")
	}
	t.Logf("Attacking %q", *target)
}
