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

func ResolveSubsamplingFactors(chromaModes []byte) (factors []uint32) {
	factors = make([]uint32, 2)

	for i := range factors {
		if chromaModes[i] == 0 {
			factors[i] = uint32(0)
		} else {
			factors[i] = uint32(LumaMode / chromaModes[i])
		}
	}
	return
}

func CalculateEncodedRowLengths(subsamplingFactors []uint32, imageWidth, bytesPerParameter uint32) (rowsLength []uint32) {
	rowsLength = make([]uint32, len(subsamplingFactors))

	for i, factor := range subsamplingFactors {
		if factor == 0 {
			rowsLength[i] = imageWidth * bytesPerParameter
		} else {
			rowsLength[i] = (imageWidth/factor)*bytesPerParameter + imageWidth*bytesPerParameter
		}
	}
	return
}

func CalculateDecodedRowLength(width uint32, bytesPerParameter uint32) uint32 {
	return ((width*bytesPerParameter*3 + 3) / 4) * 4
}
