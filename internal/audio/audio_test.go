package audio

import "testing"

func TestBoundedOutputRejectsOversizedData(t *testing.T) {
	output := boundedOutput{remaining: 3}
	if _, err := output.Write([]byte("four")); err == nil {
		t.Fatal("expected oversized command output to be rejected")
	}
	if output.Len() != 0 {
		t.Fatalf("output contains %d bytes after rejected write", output.Len())
	}
}
