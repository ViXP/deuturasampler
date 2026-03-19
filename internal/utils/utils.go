package utils

import (
	"fmt"
	"math"
)

func GetMaxValue(number_of_bytes uint32) float64 {
	return math.Pow(2, float64(number_of_bytes)*8) - 1.0
}

func DownsampleValue(value uint32, precision uint32) []byte {
	var downsampled []byte = make([]byte, precision)

	switch precision {
	case Bits8:
		downsampled[0] = byte(value)
	case Bits16:
		downsampled[0] = byte(value)
		downsampled[1] = byte(value >> 8)
	case Bits32:
		downsampled[0] = byte(value)
		downsampled[1] = byte(value >> 8)
		downsampled[2] = byte(value >> 16)
		downsampled[3] = byte(value >> 24)
	default:
		panic(fmt.Sprintf("The %vbits/parameter precision is not supported!", precision*8))
	}

	return downsampled
}
