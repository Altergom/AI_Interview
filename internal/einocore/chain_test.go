package einocore

import (
	"context"
	"testing"
)

func TestNewPassthroughChain(t *testing.T) {
	ctx := context.Background()
	r, err := NewPassthroughChain(ctx)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Invoke(ctx, "ping")
	if err != nil {
		t.Fatal(err)
	}
	if out != "ping" {
		t.Fatalf("got %q want ping", out)
	}
}
