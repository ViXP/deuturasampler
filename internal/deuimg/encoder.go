package deuimg

import (
	"deuterasampler/internal/parsers"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
	"sync"
)

type Encoder struct {
	metadata      *parsers.ImageMetadata
	bytesPerPixel uint32
	maxValue      float64
	rawData       []byte
}

func (e *Encoder) Process() [][]byte {
	rowsData := make([][]byte, e.metadata.Height)
	inputRowLength := utils.CalculateDecodedRowLength(e.metadata.Width, e.metadata.BytesPerParameter)
	subsamplingFactors := utils.ResolveSubsamplingFactors(e.metadata.ChromaMode)
	encodedRowsLength := utils.CalculateEncodedRowLengths(subsamplingFactors, e.metadata.Width, e.metadata.BytesPerParameter)

	var wg sync.WaitGroup
	wg.Add(int(e.metadata.Height))

	for row := uint32(0); row < e.metadata.Height; row++ {
		go func(rowIndex uint32, waitGroup *sync.WaitGroup) {
			rowOrder := byte(rowIndex % 2)
			byteIndex := uint32(0)
			rowsData[rowIndex] = make([]byte, encodedRowsLength[rowOrder])

			for col := uint32(0); col < e.metadata.Width; col++ {
				position := utils.BmpHeaderSize + rowIndex*inputRowLength + col*e.bytesPerPixel
				pixelData := e.rawData[position : position+e.bytesPerPixel]

				r, g, b := e.readRgb(pixelData)
				lum, cb := e.generateLumChrome(r, g, b)

				utils.WriteBytes(rowsData[rowIndex][byteIndex:byteIndex+e.metadata.BytesPerParameter], lum, e.metadata.BytesPerParameter)
				byteIndex += e.metadata.BytesPerParameter

				if subsamplingFactors[rowOrder] != 0 && (col/subsamplingFactors[rowOrder])*subsamplingFactors[rowOrder] == col {
					utils.WriteBytes(rowsData[rowIndex][byteIndex:byteIndex+e.metadata.BytesPerParameter], cb, e.metadata.BytesPerParameter)
					byteIndex += e.metadata.BytesPerParameter
				}
			}

			waitGroup.Done()
		}(row, &wg)
	}

	wg.Wait()

	return rowsData
}

func (e *Encoder) generateLumChrome(r, g, b uint32) (lum, cb uint32) {
	normal_r, normal_g, normal_b := e.normalizeColors(r, g, b)
	normal_lum := 0.2126*normal_r + 0.7152*normal_g + 0.114*normal_b
	normal_cb := (normal_b - normal_lum) / 1.772

	lum, cb = e.scaleLumChrome(normal_lum, normal_cb)

	return
}

func (e *Encoder) scaleLumChrome(normal_lum, normal_cb float64) (lum, cb uint32) {
	normal_lum = math.Min(math.Max(normal_lum, 0), 1)
	normal_cb = math.Min(math.Max(normal_cb, -0.5), 0.5)

	lum = uint32(math.Round(normal_lum * e.maxValue))
	cb = uint32(math.Round((normal_cb + 0.5) * e.maxValue))
	return
}

func (e *Encoder) normalizeColors(r, g, b uint32) (normal_r, normal_g, normal_b float64) {
	normal_r = float64(r) / e.maxValue
	normal_g = float64(g) / e.maxValue
	normal_b = float64(b) / e.maxValue
	return
}

func (e *Encoder) readRgb(bytesStream []byte) (r, g, b uint32) {
	switch e.metadata.BytesPerParameter {
	case utils.Bits8:
		r = uint32(bytesStream[0])
		g = uint32(bytesStream[1])
		b = uint32(bytesStream[2])
	case utils.Bits16:
		r = uint32(bytesStream[0]) | uint32(bytesStream[1])<<8
		g = uint32(bytesStream[2]) | uint32(bytesStream[3])<<8
		b = uint32(bytesStream[4]) | uint32(bytesStream[5])<<8
	case utils.Bits32:
		r = uint32(bytesStream[0]) | uint32(bytesStream[1])<<8 | uint32(bytesStream[2])<<16 | uint32(bytesStream[3])<<24
		g = uint32(bytesStream[4]) | uint32(bytesStream[5])<<8 | uint32(bytesStream[6])<<16 | uint32(bytesStream[7])<<24
		b = uint32(bytesStream[8]) | uint32(bytesStream[9])<<8 | uint32(bytesStream[10])<<16 | uint32(bytesStream[11])<<24
	default:
		panic(fmt.Sprintf("The %vbits/color space is not supported!", e.metadata.BytesPerParameter*8))
	}

	return
}

func NewEncoder(metadata *parsers.ImageMetadata, rawData []byte) *Encoder {
	bytesPerPixel := metadata.BytesPerParameter * 3

	fmt.Printf(
		"Width: %v, Height: %v, Bytes per parameter: %v, Subsampling mode: %v:%v:%v\n",
		metadata.Width, metadata.Height, metadata.BytesPerParameter,
		utils.LumaMode, metadata.ChromaMode[0], metadata.ChromaMode[1],
	)

	return &Encoder{
		metadata:      metadata,
		bytesPerPixel: bytesPerPixel,
		maxValue:      utils.GetMaxValue(metadata.BytesPerParameter),
		rawData:       rawData,
	}
}

var _ StreamProcessor = &Encoder{}
