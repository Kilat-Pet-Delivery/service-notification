package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestCursorRoundTrip ensures encodeCursor + decodeCursor are inverses.
// Lives in a non-integration file so plain `go test ./...` exercises it.
func TestCursorRoundTrip(t *testing.T) {
	original := cursorPayload{
		CreatedAt: time.Date(2026, 5, 16, 12, 0, 0, 123_456_000, time.UTC),
		ID:        uuid.New(),
	}
	enc := encodeCursor(original)
	if enc == "" {
		t.Fatal("encodeCursor produced empty string")
	}
	decoded, err := decodeCursor(enc)
	if err != nil {
		t.Fatalf("decodeCursor failed: %v", err)
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) || decoded.ID != original.ID {
		t.Fatalf("round-trip mismatch: got %+v want %+v", decoded, original)
	}
}

// TestDecodeCursor_InvalidInput rejects garbage cursors so the handler can
// translate them into 400 Bad Request.
func TestDecodeCursor_InvalidInput(t *testing.T) {
	cases := []string{
		"not-base64!!!", // invalid base64
		"Zm9vYmFy",      // valid base64, invalid JSON
		"",              // empty -- callers should not pass this to decodeCursor
	}
	for _, c := range cases {
		if _, err := decodeCursor(c); err == nil {
			t.Errorf("decodeCursor(%q) expected error, got nil", c)
		}
	}
}

// TestClampLimit checks the default / max / passthrough behaviour of the
// limit clamp helper used by NotificationListRepo.List.
func TestClampLimit(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{in: 0, want: 20},
		{in: -3, want: 20},
		{in: 1, want: 1},
		{in: 20, want: 20},
		{in: 50, want: 50},
		{in: 51, want: 50},
		{in: 9999, want: 50},
	}
	for _, c := range cases {
		if got := clampLimit(c.in); got != c.want {
			t.Errorf("clampLimit(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
