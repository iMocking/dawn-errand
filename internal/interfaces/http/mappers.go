package http

import (
	"hash/fnv"
	"strconv"
	"strings"
	"time"
)

// ---- 通用 ----

// thousandSep 千分位格式化整数。
func thousandSep(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b []byte
	for i, r := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, r)
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

func moneyStr(v float64) string {
	return "¥" + thousandSep(int64(v+0.5))
}

func pctStr(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64) + "%"
}

func timeHM(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("15:04")
}

func timeMD(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("01/02")
}

func timeYMD(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("2006-01-02")
}

func deltaStr(v float64) string {
	if v >= 0 {
		return "▲ " + strconv.FormatFloat(v, 'f', 1, 64) + "%"
	}
	return "▼ " + strconv.FormatFloat(-v, 'f', 1, 64) + "%"
}

// hashIdx 稳定哈希取模（用于颜色/图标选择）。
func hashIdx(s string, n int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32() % uint32(n))
}
