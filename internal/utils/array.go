package utils

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

func ExistIn(el string, items []string) bool {
	for _, v := range items {
		if v == el {
			return true
		}
	}
	return false
}

func ArrayToString[T constraints.Integer](array []T) string {
	str := []string{}
	for _, v := range array {
		str = append(str, fmt.Sprintf("%d", v))
	}
	return fmt.Sprintf("[%s]", strings.Join(str, ","))
}
