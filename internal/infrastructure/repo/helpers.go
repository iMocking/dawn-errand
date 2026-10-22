package repo

import (
	"time"

	"dawn-errand/internal/pkg/ulidgen"
)

func newULID() string { return ulidgen.New() }

func nowSec() int64 { return time.Now().Unix() }

// startOfDayUTC 当日 0 点 UTC。
func startOfDayUTC(now int64) int64 {
	t := time.Unix(now, 0).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()
}

// daysAgoStart 距今 n 天前的 0 点 UTC。
func daysAgoStart(now int64, days int) int64 {
	return startOfDayUTC(now) - int64(days)*86400
}
