package utils

import (
	"fmt"
	"math"
)

func GetMaxValue(number_of_bytes uint32) float64 {
	return math.Pow(2, float64(number_of_bytes)*8) - 1.0
}

func WriteBytes(slice []byte, value uint32, precision uint32) {
	switch precision {
	case Bits8:
		slice[0] = byte(value)
	case Bits16:
		slice[0] = byte(value)
		slice[1] = byte(value >> 8)
	case Bits32:
		slice[0] = byte(value)
		slice[1] = byte(value >> 8)
		slice[2] = byte(value >> 16)
		slice[3] = byte(value >> 24)
	default:
		panic(fmt.Sprintf("The %vbits/parameter precision is not supported!", precision*8))
	}
}
