package timerounding

import (
	"time"
)

func RoundToInterval(t time.Time, invl Interval) time.Time {
	switch invl.Unit {
	case None:
		return t
	case Seconds:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()/invl.N*invl.N, 0, t.Location())
	case Minutes:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()/invl.N*invl.N, 0, 0, t.Location())
	case Hours:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()/invl.N*invl.N, 0, 0, 0, t.Location())
	case Days:
		// somewhat questionable
		return time.Date(t.Year(), t.Month(), t.Day()/invl.N*invl.N, 0, 0, 0, 0, t.Location())
	default:
		panic("Invalid unit")
	}

}

func RoundToDuration(t time.Time, d time.Duration) time.Time {
	invl, err := ToInterval(d)
	if err != nil {
		panic(err)
	}

	return RoundToInterval(t, invl)
}
