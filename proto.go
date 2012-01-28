// Author:  Steve Phillips / elimisteve
// Started: 2012.01.27
// Updated: 2012.01.27

package proto

import (
    "fmt"
	"proto"
)

func Range(values ...int) []int {
	var intSlice []int
	var min, max int
	var step int = 1

	length := len(values)
	switch length {
	case 0:
		;
	case 1:
		max = values[0]
	case 2:
		min, max = values[0], values[1]
	case 3:
		min, max, step = values[0], values[1], values[2]
	default:
		;
	}
	// Main loop. Gets executed no matter what
	for i := min; i < max; i += step {
		intSlice = append(intSlice, i)
	}
	return intSlice
}
