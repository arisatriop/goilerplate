package utils

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func NowJakarta() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc)
}
