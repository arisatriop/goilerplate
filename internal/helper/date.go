package helper

import "time"

func Now() time.Time {
	return time.Now()
}

func NowUTC() time.Time {
	return Now().UTC()
}

func NowJakarta() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return Now().In(loc)
}
