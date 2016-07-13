package timerounding

import (
	"testing"
	"time"
)

var intervalTests = []struct {
	input    time.Duration
	expected string
}{
	{0, "none"},
	{1 * time.Second, "1s"},
	{5 * time.Second, "5s"},
	{60 * time.Second, "1m"},
	{1 * time.Minute, "1m"},
	{15 * time.Minute, "15m"},
	{1 * time.Hour, "1h"},
	{2 * time.Hour, "2h"},
	{23 * time.Hour, "23h"},
	{24 * time.Hour, "1d"},
}

func TestInterval(t *testing.T) {
	for _, tt := range intervalTests {
		ainvl, err := ToInterval(tt.input)
		actual := ainvl.String()
		if err != nil {
			t.Errorf("ToInterval(%q) failed: %v, expected %q", tt.input, err, tt.expected)
		} else if actual != tt.expected {
			t.Errorf("ToInterval(%q) == %q, expected %q", tt.input, actual, tt.expected)
		} else {
			t.Logf("ToInterval(%q) == %q", tt.input, actual)
		}
	}
}
