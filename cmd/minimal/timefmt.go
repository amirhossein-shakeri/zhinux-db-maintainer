package main

import (
	"fmt"
	"time"
)

func formatOutputTimestamp(jalali bool, now time.Time) string {
	if !jalali {
		return now.UTC().Format("20060102T150405Z")
	}

	tehran, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		tehran = time.FixedZone("Asia/Tehran", 3*3600+30*60)
	}
	local := now.In(tehran)
	jy, jm, jd := gregorianToJalali(local.Year(), int(local.Month()), local.Day())
	return fmt.Sprintf("%04d-%02d-%02d--%02d-%02d-%02d-%03d",
		jy, jm, jd, local.Hour(), local.Minute(), local.Second(), local.Nanosecond()/1e6)
}

func gregorianToJalali(gy, gm, gd int) (jy, jm, jd int) {
	gdm := [...]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	if gy > 1600 {
		jy = 979
		gy -= 1600
	} else {
		jy = 0
		gy -= 621
	}
	gy2 := gy
	if gm > 2 {
		gy2++
	}
	days := 365*gy + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 - 80 + gd + gdm[gm-1]
	jy += 33 * (days / 12053)
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	if days < 186 {
		jm = 1 + days/31
		jd = 1 + days%31
	} else {
		jm = 7 + (days-186)/30
		jd = 1 + (days-186)%30
	}
	return
}

