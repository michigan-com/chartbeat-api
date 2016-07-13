package timerounding

import (
	"testing"
	"time"
)

var periodFormatTests = []struct {
	time     time.Time
	invl     Interval
	expected string
}{
	{time.Date(2016, 1, 2, 13, 4, 00, 0, time.UTC), Interval5m, "20160102-1300"},
	{time.Date(2016, 1, 2, 13, 0, 00, 0, time.UTC), Interval5m, "20160102-1300"},
	{time.Date(2016, 1, 2, 13, 5, 00, 0, time.UTC), Interval5m, "20160102-1305"},
	{time.Date(2016, 1, 2, 13, 6, 00, 0, time.UTC), Interval5m, "20160102-1305"},
	{time.Date(2016, 1, 2, 13, 9, 00, 0, time.UTC), Interval5m, "20160102-1305"},
	{time.Date(2016, 1, 2, 13, 59, 00, 0, time.UTC), Interval5m, "20160102-1355"},

	{time.Date(2016, 1, 2, 13, 9, 00, 0, time.UTC), Interval15m, "20160102-1300"},
	{time.Date(2016, 1, 2, 13, 40, 00, 0, time.UTC), Interval15m, "20160102-1330"},
}

func TestPeriodFormat(t *testing.T) {
	for _, tt := range periodFormatTests {
		actual := FormatRoundedToInterval(tt.time, tt.invl)
		if actual != tt.expected {
			t.Errorf("Format(%q, %q) == %q, expected %q", tt.time, tt.invl, actual, tt.expected)
		} else {
			t.Logf("Format(%q, %q) == %q", tt.time, tt.invl, actual)
		}
	}
}
