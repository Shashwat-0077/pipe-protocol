package utils

import (
	"errors"
	"strconv"
	"strings"
)

func ParseSequence(sequence string) (start, end int, err error) {
	if sequence == "*" {
		return -1, -1, nil // special case: handled separately
	} else if strings.Contains(sequence, ":") {
		parts := strings.Split(sequence, ":")
		if len(parts) != 2 {
			return 0, 0, errors.New("invalid sequence range")
		}
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || start <= 0 || end < start {
			return 0, 0, errors.New("invalid sequence bounds")
		}
		return start, end, nil
	} else {
		index, err := strconv.Atoi(sequence)
		if err != nil || index <= 0 {
			return 0, 0, errors.New("invalid sequence index")
		}
		return index, index, nil
	}
}
