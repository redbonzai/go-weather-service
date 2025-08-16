package util

import (
	"time"
)

// IsSameLocalDay returns true if the two times occur on the same local date.
func IsSameLocalDay(a, b time.Time, loc *time.Location) bool {
	aa := a.In(loc)
	bb := b.In(loc)
	return aa.Year() == bb.Year() && aa.Month() == bb.Month() && aa.Day() == bb.Day()
}
