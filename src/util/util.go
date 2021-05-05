package util

import (
	"strings"
)

func ParseStringToSlice(input string) []string {
	return strings.Split(strings.Replace(input, " ", "", -1), ",")
}
