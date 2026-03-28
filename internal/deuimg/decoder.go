package deuimg

import (
	"bytes"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
	"runtime"
	"sync"
)

type Decoder struct {
	Height            uint32
	Width             uint32
	bytesPerParameter uint32
	chromaMode        []byte
	maxValue          float64
	encodedData       []byte
	decodingBuffer    *bytes.Buffer
}

func (d *Decoder) Process() []byte {
	var waitGroup sync.WaitGroup

	subsamplingFactors := utils.ResolveSubsamplingFactors(d.chromaMode)
	cbPerRow := d.calculateCbPerRow(subsamplingFactors)

	outputData := make([][]byte, d.Height)
	decodedRowLength := utils.CalculateDecodedRowLength(d.Width, d.bytesPerParameter)
	encodedRowsLength := utils.CalculateEncodedRowLengths(subsamplingFactors, d.Width, d.bytesPerParameter)

	routinesNum := runtime.NumCPU()
	rowsPerRoutine := d.calculateRowsPerRoutine(routinesNum)
	waitGroup.Add(routinesNum)

	for i := range routinesNum {
		go func(iteration uint32, wg *sync.WaitGroup) {
			cbBuffer := make([]uint32, cbPerRow)

			startRow := iteration * rowsPerRoutine
			endRow := min(startRow+rowsPerRoutine, d.Height)

			for row := startRow; row < endRow; row++ {
				outputData[row] = make([]byte, decodedRowLength)
				byteIndex := uint32(0)
				rowOrder := uint32(row % 2)
				bitPosition := d.findRowStartingPosition(row, encodedRowsLength)

				for pxlNum := uint32(0); pxlNum < d.Width; pxlNum++ {
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
						cb = d.readCb(d.encodedData[bitPosition+d.bytesPerParameter : bitPosition+d.bytesPerParameter*2])
						cbBuffer[cbGroup] = cb
					} else {
						cb = cbBuffer[cbGroup]
					}

					lum := d.readLum(d.encodedData[bitPosition : bitPosition+d.bytesPerParameter])
					r, g, b := d.generateRgb(lum, cb)

					utils.WriteBytes(outputData[row][byteIndex:byteIndex+d.bytesPerParameter], r, d.bytesPerParameter)
					byteIndex += d.bytesPerParameter

					utils.WriteBytes(outputData[row][byteIndex:byteIndex+d.bytesPerParameter], g, d.bytesPerParameter)
					byteIndex += d.bytesPerParameter

					utils.WriteBytes(outputData[row][byteIndex:byteIndex+d.bytesPerParameter], b, d.bytesPerParameter)
					byteIndex += d.bytesPerParameter

					bitPosition += d.bytesPerParameter
					if isSubsampled {
						bitPosition += d.bytesPerParameter
					}
				}
			}
			wg.Done()
		}(uint32(i), &waitGroup)
	}

	waitGroup.Wait()

	return d.prepareDecodedBitmap(outputData)
}

func (d *Decoder) prepareDecodedBitmap(rowsData [][]byte) []byte {
	d.writeBitmapHeader()

	for _, data := range rowsData {
		d.decodingBuffer.Write(data)
	}

	decodedData := d.decodingBuffer.Bytes()
	utils.WriteBytes(decodedData[2:6], uint32(len(decodedData)), utils.Bits32)

	return decodedData
}

func (d *Decoder) writeBitmapHeader() {
	header := make([]byte, utils.BmpHeaderSize)
	header[0] = byte('B')
	header[1] = byte('M')

	utils.WriteBytes(header[10:14], utils.BmpHeaderSize, utils.Bits32)
	header[14] = byte(40)
	utils.WriteBytes(header[18:22], d.Width, utils.Bits32)
	utils.WriteBytes(header[22:26], d.Height, utils.Bits32)
	header[26] = 1
	utils.WriteBytes(header[28:30], d.bytesPerParameter*3*8, utils.Bits16)
	d.decodingBuffer.Write(header)
}

func (d *Decoder) readCb(data []byte) (cb uint32) {
	switch d.bytesPerParameter {
	case utils.Bits8:
		cb = uint32(data[0])
	case utils.Bits16:
		cb = uint32(data[0]) | uint32(data[1])<<8
	case utils.Bits32:
		cb = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	default:
		panic(fmt.Sprintf("The %vbits/parameter is not supported!", d.bytesPerParameter*8))
	}

	return
}

func (d *Decoder) readLum(data []byte) (lum uint32) {
	switch d.bytesPerParameter {
	case utils.Bits8:
		lum = uint32(data[0])
	case utils.Bits16:
		lum = uint32(data[0]) | uint32(data[1])<<8
	case utils.Bits32:
		lum = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	default:
		panic(fmt.Sprintf("The %vbits/parameter is not supported!", d.bytesPerParameter*8))
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
	rowsPerRoutine := uint32(math.Ceil(float64(d.Height) / float64(routinesNum)))
	if rowsPerRoutine%2 == 1 {
		rowsPerRoutine += 1
	}
	return rowsPerRoutine
}

func (d *Decoder) calculateCbPerRow(subsamplingFactors []uint32) (cbPerRow uint32) {
	if subsamplingFactors[0] == 0 {
		cbPerRow = 1
	} else {
		cbPerRow = uint32(math.Ceil(float64(d.Width) / float64(subsamplingFactors[0])))
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

func NewDecoder(bytesPerParameter, height, width uint32, chromaMode []byte, encodedData []byte) *Decoder {
	fmt.Printf(
		"Width: %v, Height: %v, Bytes per parameter: %v, Subsampling mode: %v:%v:%v\n", width, height, bytesPerParameter,
		utils.LumaMode, chromaMode[0], chromaMode[1],
	)
	return &Decoder{
		Height:            height,
		Width:             width,
		bytesPerParameter: bytesPerParameter,
		chromaMode:        chromaMode,
		maxValue:          utils.GetMaxValue(bytesPerParameter),
		encodedData:       encodedData,
		decodingBuffer:    bytes.NewBuffer(nil),
	}
}

var _ StreamProcessor = &Decoder{}
