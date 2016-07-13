package timerounding

import (
	"errors"
	"fmt"
	"time"
)

type Interval struct {
	N    int
	Unit Unit
}

var (
	Interval1m  = Interval{1, Minutes}
	Interval5m  = Interval{5, Minutes}
	Interval15m = Interval{15, Minutes}
)

var (
	ErrNegativeDuration = errors.New("negative duration")
	ErrSmallDuration    = errors.New("duration too small")
	ErrMultipleUnits    = errors.New("duration is a mix of multiple units")
)

var (
	units = []Unit{Days, Hours, Minutes, Seconds}
)

func ToInterval(d time.Duration) (Interval, error) {
	if d == 0 {
		return Interval{0, None}, nil
	}
	if d < 0 {
		return Interval{0, None}, ErrNegativeDuration
	}

	for _, u := range units {
		invl := u.Duration()

		if d >= invl {
			v := d / invl

			var err error
			if v*invl != d {
				err = ErrMultipleUnits
			}

			return Interval{int(v), u}, err
		}
	}

	return Interval{0, None}, ErrSmallDuration
}

func (invl Interval) IsNone() bool {
	return invl.Unit == None
}

func (invl Interval) String() string {
	if invl.IsNone() {
		return "none"
	}
	return fmt.Sprintf("%d%s", invl.N, invl.Unit.String())
}
