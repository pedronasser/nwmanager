package helpers

import "time"

func GetCurrentTimeAsUTC() time.Time {
	now := time.Now()
	date := now.Format("02/01/2006")
	hour := now.Format("15:04")
	utcTime, _ := time.Parse("02/01/2006 15:04", date+" "+hour)
	return utcTime
}
