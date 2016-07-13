package timerounding

import (
	"time"
)

//Mon Jan 2 15:04:05 -0700 MST 2006
func FormatToUnit(t time.Time, u Unit) string {
	switch u {
	case None, Seconds:
		return t.Format("20060102-150405")
	case Minutes:
		return t.Format("20060102-1504")
	case Hours:
		return t.Format("20060102-15")
	case Days:
		return t.Format("20060102")
	default:
		panic("Invalid unit")
	}

}

func FormatRoundedToInterval(t time.Time, invl Interval) string {
	return FormatToUnit(RoundToInterval(t, invl), invl.Unit)
}

func FormatRoundedToDuration(t time.Time, d time.Duration) string {
	invl, err := ToInterval(d)
	if err != nil {
		panic(err)
	}

	return FormatRoundedToInterval(t, invl)
}
