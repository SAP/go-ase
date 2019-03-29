package libase

import (
	"time"
)

// TimeToMicroseconds converts a time.Time into the number of
// microseconds elapsed since time.Time{}.
func TimeToMicroseconds(date time.Time) uint64 {
	year, month, day := date.Year(), int(date.Month()), date.Day()

	// Calculate JDN, formula from Calendars by Doggett
	jd := (1461*(year+4800+(month-14)/12))/4 + (367*(month-2-12*((month-14)/12)))/12 - (3*((year+4900+(month-14)/12)/100))/4 + day - 32075

	// Calculate Rata Die from JDN and convert to microseconds
	rataDie := uint64(jd-1721425) * (uint64(time.Duration(24)*time.Hour) / 1000)
	// While Sybase uses Rata Die it still seems to count year 0
	rataDie = rataDie + uint64((time.Hour*24)*365/1000)

	// Get hours, minutes and seconds as microseconds
	hours := uint64(time.Duration(date.Hour())*time.Hour) / 1000
	minutes := uint64(time.Duration(date.Minute())*time.Minute) / 1000
	seconds := uint64(time.Duration(date.Second())*time.Second) / 1000
	nanoseconds := uint64(date.Nanosecond()) / 1000

	// Count hours, minutes and seconds as nanoseconds, sum and calcuate
	// microseconds
	return rataDie + hours + minutes + seconds + nanoseconds
}

// MicrosecondsToTime takes the number of microseconds since time.Time{}
// and calculates the appropriate time.Time.
func MicrosecondsToTime(microseconds uint64) time.Time {
	// Formula from CACM, Fliegel / Van Flandern
	// Implementation from github.com/thda/tds

	// Calculate small offsets
	nanoseconds := int(microseconds%1000000) * 1000
	seconds := int(microseconds/1000000) % 86400

	// Calculate julian day
	jD := int(microseconds/1000000)/86400 - 693961

	// Convert julian day to date
	l := jD + 68569 + 2415021
	n := 4 * l / 146097
	l = l - (146097*n+3)/4
	y := 4000 * (l + 1) / 1461001
	l = l - 1461*y/4 + 31
	m := 80 * l / 2447
	d := l - 2447*m/80
	l = m / 11
	m = m + 2 - 12*l
	y = 100*(n-49) + y + l
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)

	if nanoseconds != 0 {
		t = t.Add(time.Duration(nanoseconds) * time.Nanosecond)
	}

	if seconds != 0 {
		t = t.Add(time.Duration(seconds) * time.Second)
	}

	return t
}
