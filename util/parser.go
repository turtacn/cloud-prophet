package util

import "strconv"

func ParserString2Float(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}
