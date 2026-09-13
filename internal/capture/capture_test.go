package capture

import "testing"

func TestLimitedBufferRejectsOversizedOutput(t *testing.T) {
	buffer := limitedBuffer{remaining: 3}
	if _, err := buffer.Write([]byte("four")); err == nil {
		t.Fatal("expected oversized output to be rejected")
	}
	if buffer.Len() != 0 {
		t.Fatalf("buffer contains %d bytes after rejected write", buffer.Len())
	}
}

func TestSafeDimensions(t *testing.T) {
	tests := []struct {
		width, height int
		want          bool
	}{
		{1920, 1080, true},
		{15360, 6480, true},
		{10_000, 10_001, false},
		{0, 1080, false},
		{-1, 1080, false},
	}
	for _, tc := range tests {
		if got := safeDimensions(tc.width, tc.height); got != tc.want {
			t.Errorf("safeDimensions(%d, %d)=%v, want %v", tc.width, tc.height, got, tc.want)
		}
	}
}

func TestNewRejectsRelativeCustomCommand(t *testing.T) {
	if _, err := New([]string{"capture-helper"}); err == nil {
		t.Fatal("expected relative custom command to be rejected")
	}
}
