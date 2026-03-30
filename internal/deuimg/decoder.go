package deuimg

import (
	"deuterasampler/internal/parsers"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
	"runtime"
	"sync"
)

type Decoder struct {
	metadata    *parsers.ImageMetadata
	maxValue    float64
	encodedData []byte
}

func (d *Decoder) Process() [][]byte {
	var waitGroup sync.WaitGroup

	subsamplingFactors := utils.ResolveSubsamplingFactors(d.metadata.ChromaMode)
	cbPerRow := d.calculateCbPerRow(subsamplingFactors)

	outputData := make([][]byte, d.metadata.Height)
	decodedRowLength := utils.CalculateDecodedRowLength(d.metadata.Width, d.metadata.BytesPerParameter)
	encodedRowsLength :=
		utils.CalculateEncodedRowLengths(subsamplingFactors, d.metadata.Width, d.metadata.BytesPerParameter)

	routinesNum := runtime.NumCPU()
	rowsPerRoutine := d.calculateRowsPerRoutine(routinesNum)
	waitGroup.Add(routinesNum)

	for i := range routinesNum {
		go func(iteration uint32, wg *sync.WaitGroup) {
			cbBuffer := make([]uint32, cbPerRow)

			startRow := iteration * rowsPerRoutine
			endRow := min(startRow+rowsPerRoutine, d.metadata.Height)

			for row := startRow; row < endRow; row++ {
				outputData[row] = make([]byte, decodedRowLength)
				byteIndex := uint32(0)
				rowOrder := uint32(row % 2)
				bitPosition := d.findRowStartingPosition(row, encodedRowsLength)

				for pxlNum := uint32(0); pxlNum < d.metadata.Width; pxlNum++ {
					var cbGroup uint32
					if subsamplingFactors[0] == 0 {
						cbGroup = 0
					} else {
						cbGroup = pxlNum / subsamplingFactors[0]
					}

					isSubsampled := false
					if subsamplingFactors[rowOrder] != 0 {
						isSubsampled = pxlNum%subsamplingFactors[rowOrder] == 0
					}

					var cb uint32
					if isSubsampled {
						cb = d.readCb(d.encodedData[bitPosition+d.metadata.BytesPerParameter : bitPosition+d.metadata.BytesPerParameter*2])
						cbBuffer[cbGroup] = cb
					} else {
						cb = cbBuffer[cbGroup]
					}

					lum := d.readLum(d.encodedData[bitPosition : bitPosition+d.metadata.BytesPerParameter])
					r, g, b := d.generateRgb(lum, cb)

					utils.WriteBytes(outputData[row][byteIndex:byteIndex+d.metadata.BytesPerParameter], r, d.metadata.BytesPerParameter)
					byteIndex += d.metadata.BytesPerParameter

					utils.WriteBytes(outputData[row][byteIndex:byteIndex+d.metadata.BytesPerParameter], g, d.metadata.BytesPerParameter)
					byteIndex += d.metadata.BytesPerParameter

					utils.WriteBytes(outputData[row][byteIndex:byteIndex+d.metadata.BytesPerParameter], b, d.metadata.BytesPerParameter)
					byteIndex += d.metadata.BytesPerParameter

					bitPosition += d.metadata.BytesPerParameter
					if isSubsampled {
						bitPosition += d.metadata.BytesPerParameter
					}
				}
			}
			wg.Done()
		}(uint32(i), &waitGroup)
	}

	waitGroup.Wait()

	return outputData
}

func (d *Decoder) readCb(data []byte) (cb uint32) {
	switch d.metadata.BytesPerParameter {
	case utils.Bits8:
		cb = uint32(data[0])
	case utils.Bits16:
		cb = uint32(data[0]) | uint32(data[1])<<8
	case utils.Bits32:
		cb = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	default:
		panic(fmt.Sprintf("The %vbits/parameter is not supported!", d.metadata.BytesPerParameter*8))
	}

	return
}

func (d *Decoder) readLum(data []byte) (lum uint32) {
	switch d.metadata.BytesPerParameter {
	case utils.Bits8:
		lum = uint32(data[0])
	case utils.Bits16:
		lum = uint32(data[0]) | uint32(data[1])<<8
	case utils.Bits32:
		lum = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	default:
		panic(fmt.Sprintf("The %vbits/parameter is not supported!", d.metadata.BytesPerParameter*8))
	}

	return
}

func (d *Decoder) normalizeLumChrome(luminance, cb uint32) (normal_lum, normal_cb float64) {
	normal_lum = float64(luminance) / d.maxValue
	normal_cb = float64(cb)/d.maxValue - 0.5

	return
}

func (d *Decoder) generateRgb(lum, cb uint32) (r, g, b uint32) {
	var normal_b, normal_g, normal_r, normal_lum, normal_cb float64
	normal_lum, normal_cb = d.normalizeLumChrome(lum, cb)

	normal_b = (normal_cb * 1.772) + normal_lum
	differentiator := 0.25 * (normal_b - normal_lum)
	normal_r = normal_lum - differentiator
	normal_g = normal_lum + differentiator

	r, g, b = d.scalePixelValues(normal_r, normal_g, normal_b)

	return
}

func (d *Decoder) scalePixelValues(normal_r, normal_g, normal_b float64) (r, g, b uint32) {
	r = uint32(math.Min(math.Max(normal_r*d.maxValue, 0), d.maxValue))
	g = uint32(math.Min(math.Max(normal_g*d.maxValue, 0), d.maxValue))
	b = uint32(math.Min(math.Max(normal_b*d.maxValue, 0), d.maxValue))
	return
}

func (d *Decoder) calculateRowsPerRoutine(routinesNum int) uint32 {
	rowsPerRoutine := uint32(math.Ceil(float64(d.metadata.Height) / float64(routinesNum)))
	if rowsPerRoutine%2 == 1 {
		rowsPerRoutine += 1
	}
	return rowsPerRoutine
}

func (d *Decoder) calculateCbPerRow(subsamplingFactors []uint32) (cbPerRow uint32) {
	if subsamplingFactors[0] == 0 {
		cbPerRow = 1
	} else {
		cbPerRow = uint32(math.Ceil(float64(d.metadata.Width) / float64(subsamplingFactors[0])))
	}
	return
}

func (d *Decoder) findRowStartingPosition(row uint32, encodedRowsLength []uint32) (rowStartPosition uint32) {
	rowStartPosition = utils.DeuimgHeaderSize

	if row%2 == 0 {
		rowStartPosition += (row / 2) * (encodedRowsLength[0] + encodedRowsLength[1])
	} else {
		rowStartPosition += (row-1)/2*(encodedRowsLength[0]+encodedRowsLength[1]) + encodedRowsLength[0]
	}

	return
}

func NewDecoder(metadata *parsers.ImageMetadata, encodedData []byte) *Decoder {
	fmt.Printf(
		"Width: %v, Height: %v, Bytes per parameter: %v, Subsampling mode: %v:%v:%v\n",
		metadata.Width, metadata.Height, metadata.BytesPerParameter, utils.LumaMode, metadata.ChromaMode[0],
		metadata.ChromaMode[1],
	)
	return &Decoder{
		metadata:    metadata,
		maxValue:    utils.GetMaxValue(metadata.BytesPerParameter),
		encodedData: encodedData,
	}
}

var _ StreamProcessor = &Decoder{}
