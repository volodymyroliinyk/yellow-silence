package process

import "testing"

func TestParsePID(t *testing.T) {
	for _, tc := range []struct {
		value string
		ok    bool
	}{{"1", true}, {"12345", true}, {"", false}, {"self", false}, {"12x", false}} {
		_, err := parsePID(tc.value)
		if (err == nil) != tc.ok {
			t.Errorf("parsePID(%q) error=%v, want ok=%v", tc.value, err, tc.ok)
		}
	}
}
