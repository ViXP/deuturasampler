package deuimg

import (
	"bytes"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
	"sync"
)

type Encoder struct {
	Height            uint32
	Width             uint32
	FileSize          uint32
	chromaMode        []byte
	bytesPerPixel     uint32
	bytesPerParameter uint32
	maxValue          float64
	rawData           *[]byte
	processingBuffer  *bytes.Buffer
}

func (e *Encoder) Process() []byte {
	rowsData := make([][]byte, e.Height)
	inputRowLength := utils.CalculateDecodedRowLength(e.Width, e.bytesPerParameter)
	subsamplingFactors := utils.ResolveSubsamplingFactors(e.chromaMode)
	encodedRowsLength := utils.CalculateEncodedRowLengths(subsamplingFactors, e.Width, e.bytesPerParameter)

	var wg sync.WaitGroup
	wg.Add(int(e.Height))

	for row := uint32(0); row < e.Height; row++ {
		go func(rowIndex uint32, waitGroup *sync.WaitGroup) {
			rowOrder := byte(rowIndex % 2)
			byteIndex := uint32(0)
			rowsData[rowIndex] = make([]byte, encodedRowsLength[rowOrder])

			for col := uint32(0); col < e.Width; col++ {
				position := utils.BmpHeaderSize + rowIndex*inputRowLength + col*e.bytesPerPixel
				pixelData := (*e.rawData)[position : position+e.bytesPerPixel]

				r, g, b := e.readRgb(pixelData)
				lum, cb := e.generateLumChrome(r, g, b)

				utils.WriteBytes(rowsData[rowIndex][byteIndex:byteIndex+e.bytesPerParameter], lum, e.bytesPerParameter)
				byteIndex += e.bytesPerParameter

				if subsamplingFactors[rowOrder] != 0 && (col/subsamplingFactors[rowOrder])*subsamplingFactors[rowOrder] == col {
					utils.WriteBytes(rowsData[rowIndex][byteIndex:byteIndex+e.bytesPerParameter], cb, e.bytesPerParameter)
					byteIndex += e.bytesPerParameter
				}
			}

			waitGroup.Done()
		}(row, &wg)
	}

	wg.Wait()

	return e.prepareEncodedDeuimg(rowsData)
}

func (e *Encoder) prepareEncodedDeuimg(rowsData [][]byte) []byte {
	e.writeHeader()

	for _, row := range rowsData {
		e.processingBuffer.Write(row)
	}

	return e.processingBuffer.Bytes()
}

func (e *Encoder) writeHeader() {
	header := make([]byte, utils.DeuimgHeaderSize)

	for i, sym := range "(-_-)" {
		header[i] = byte(sym)
	}

	utils.WriteBytes(header[5:9], e.Width, utils.Bits32)
	utils.WriteBytes(header[9:13], e.Height, utils.Bits32)
	utils.WriteBytes(header[13:17], e.bytesPerParameter, utils.Bits32)
	header[17] = e.chromaMode[0]
	header[18] = e.chromaMode[1]
	e.processingBuffer.Write(header)
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

func (e *Encoder) readRgb(bytes_stream []byte) (r, g, b uint32) {
	switch e.bytesPerParameter {
	case utils.Bits8:
		r = uint32(bytes_stream[0])
		g = uint32(bytes_stream[1])
		b = uint32(bytes_stream[2])
	case utils.Bits16:
		r = uint32(bytes_stream[0]) | uint32(bytes_stream[1])<<8
		g = uint32(bytes_stream[2]) | uint32(bytes_stream[3])<<8
		b = uint32(bytes_stream[4]) | uint32(bytes_stream[5])<<8
	case utils.Bits32:
		r = uint32(bytes_stream[0]) | uint32(bytes_stream[1])<<8 | uint32(bytes_stream[2])<<16 | uint32(bytes_stream[3])<<24
		g = uint32(bytes_stream[4]) | uint32(bytes_stream[5])<<8 | uint32(bytes_stream[6])<<16 | uint32(bytes_stream[7])<<24
		b = uint32(bytes_stream[8]) | uint32(bytes_stream[9])<<8 | uint32(bytes_stream[10])<<16 | uint32(bytes_stream[11])<<24
	default:
		panic(fmt.Sprintf("The %vbits/color space is not supported!", e.bytesPerParameter*8))
	}

	return
}

func NewEncoder(bytesPerPixel, height, width uint32, chromaMode []byte, rawData *[]byte) *Encoder {
	bytesPerParameter := bytesPerPixel / 3

	fmt.Printf(
		"Width: %v, Height: %v, Bytes per parameter: %v, Subsampling mode: %v:%v:%v\n", width, height, bytesPerParameter,
		utils.LumaMode, chromaMode[0], chromaMode[1],
	)

	return &Encoder{
		bytesPerPixel:     bytesPerPixel,
		bytesPerParameter: bytesPerParameter,
		Height:            height,
		Width:             width,
		FileSize:          uint32(len(*rawData)),
		chromaMode:        chromaMode,
		maxValue:          utils.GetMaxValue(bytesPerParameter),
		rawData:           rawData,
		processingBuffer:  bytes.NewBuffer(nil),
	}
}

var _ StreamProcessor = &Encoder{}
