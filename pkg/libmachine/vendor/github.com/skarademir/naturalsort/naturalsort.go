package naturalsort

import (
	"regexp"
	"strings"
)

type NaturalSort []string

var r = regexp.MustCompile(`[^0-9]+|[0-9]+`)

func (s NaturalSort) Len() int {
	return len(s)
}
func (s NaturalSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NaturalSort) Less(i, j int) bool {

	spliti := r.FindAllString(strings.Replace(s[i], " ", "", -1), -1)
	splitj := r.FindAllString(strings.Replace(s[j], " ", "", -1), -1)

	for index := 0; index < len(spliti) && index < len(splitj); index++ {
		if spliti[index] != splitj[index] {
			// Both slices are numbers
			if isNumber(spliti[index][0]) && isNumber(splitj[index][0]) {
				// Remove Leading Zeroes
				stringi := strings.TrimLeft(spliti[index], "0")
				stringj := strings.TrimLeft(splitj[index], "0")
				if len(stringi) == len(stringj) {
					for indexchar := 0; indexchar < len(stringi); indexchar++ {
						if stringi[indexchar] != stringj[indexchar] {
							return stringi[indexchar] < stringj[indexchar]
						}
					}
					return len(spliti[index]) < len(splitj[index])
				}
				return len(stringi) < len(stringj)
			}
			// One of the slices is a number (we give precedence to numbers regardless of ASCII table position)
			if isNumber(spliti[index][0]) || isNumber(splitj[index][0]) {
				return isNumber(spliti[index][0])
			}
			// Both slices are not numbers
			return spliti[index] < splitj[index]
		}

	}
	// Fall back for cases where space characters have been annihliated by the replacment call
	// Here we iterate over the unmolsested string and prioritize numbers over
	for index := 0; index < len(s[i]) && index < len(s[j]); index++ {
		if isNumber(s[i][index]) || isNumber(s[j][index]) {
			return isNumber(s[i][index])
		}
	}
	return s[i] < s[j]
}
func isNumber(input uint8) bool {
	return input >= '0' && input <= '9'
}
