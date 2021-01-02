package util

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	B float64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

var unitMap = map[string]float64{
	"B": B,
	"K": KB,
	"M": MB,
	"G": GB,
	"T": TB,
	"P": PB,

	"KB": KB,
	"MB": MB,
	"GB": GB,
	"TB": TB,
	"PB": PB,
}

func StringToBytes(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if value, err := strconv.ParseInt(s, 10, 64); err == nil {
		return float64(value), nil
	}

	split := make([]string, 0)
	for i, r := range s {
		if !unicode.IsDigit(r) {
			split = append(split, strings.TrimSpace(string(s[:i])))
			split = append(split, strings.TrimSpace(string(s[i:])))
			break
		}
	}

	if len(split) != 2 {
		return 0, errors.New("unknown format " + s)
	}

	unit, ok := unitMap[strings.ToUpper(split[1])]
	if !ok {
		return 0, errors.New("unknown suffix " + split[1])
	}

	value, err := strconv.ParseFloat(split[0], 64)
	if err != nil {
		return 0, err
	}

	return value * unit, nil
}

// convert bytes to string. reserved two decimals
func BytesToString(bytes float64) string {
	unit := ""
	var value float64

	switch {
	case bytes >= PB:
		unit = "PB"
		value = bytes / PB
	case bytes >= TB:
		unit = "TB"
		value = bytes / TB
	case bytes >= GB:
		unit = "GB"
		value = bytes / GB
	case bytes >= MB:
		unit = "MB"
		value = bytes / MB
	case bytes >= KB:
		unit = "KB"
		value = bytes / KB
	case bytes >= B:
		unit = "B"
		value = bytes
	default:
		unit = ""
		value = bytes
	}

	return strconv.FormatFloat(round(value, 2), 'f', -1, 64) + unit
}

func round(n float64, decimal uint32) float64 {
	return math.Round(n*math.Pow(10, float64(decimal))) / math.Pow(10, float64(decimal))
}
