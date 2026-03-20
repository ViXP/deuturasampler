package deuimg

import (
	"bytes"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
	"sync"
)

type Decoder struct {
	Height            uint32
	Width             uint32
	FileSize          uint32
	bytesPerPixel     uint32
	bytesPerParameter uint32
	chromaMode        []byte
	maxValue          float64
	rawData           *[]byte
	decodingBuffer    *bytes.Buffer
}

func (d *Decoder) Process() []byte {
	d.writeHeader()

	var wg sync.WaitGroup

	fullRowWidth := ((d.Width*d.bytesPerParameter*3 + 3) / 4) * 4
	rowData := make([][]byte, d.Height)
	wg.Add(int(d.Height))

	for row := uint32(0); row < d.Height; row++ {
		go func(rowIndex uint32, waitGroup *sync.WaitGroup) {
			rowData[rowIndex] = make([]byte, fullRowWidth)
			byteIndex := uint32(0)

			for col := uint32(0); col < d.Width; col++ {
				position := utils.DeuimgHeaderSize + col*d.bytesPerPixel + rowIndex*d.Width*d.bytesPerPixel
				lum, cb := d.readLumCb((*d.rawData)[position : position+d.bytesPerPixel])
				r, g, b := d.generateRgb(lum, cb)

				utils.WriteBytes(rowData[rowIndex][byteIndex:byteIndex+d.bytesPerParameter], r, d.bytesPerParameter)
				byteIndex += d.bytesPerParameter

				utils.WriteBytes(rowData[rowIndex][byteIndex:byteIndex+d.bytesPerParameter], g, d.bytesPerParameter)
				byteIndex += d.bytesPerParameter

				utils.WriteBytes(rowData[rowIndex][byteIndex:byteIndex+d.bytesPerParameter], b, d.bytesPerParameter)
				byteIndex += d.bytesPerParameter
			}

			waitGroup.Done()
		}(row, &wg)
	}

	wg.Wait()

	for _, data := range rowData {
		d.decodingBuffer.Write(data)
	}

	d.assignFileSize()

	decoded_data := d.decodingBuffer.Bytes()

	utils.WriteBytes(decoded_data[2:6], d.FileSize, utils.Bits32)

	return decoded_data
}

func (d *Decoder) writeHeader() {
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

func (d *Decoder) assignFileSize() {
	d.FileSize = uint32(d.decodingBuffer.Len())
}

func (d *Decoder) readLumCb(data []byte) (lum, cb uint32) {
	switch d.bytesPerParameter {
	case utils.Bits8:
		lum = uint32(data[0])
		cb = uint32(data[1])
	case utils.Bits16:
		lum = uint32(data[0]) | uint32(data[1])<<8
		cb = uint32(data[2]) | uint32(data[3])<<8
	case utils.Bits32:
		lum = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		cb = uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
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

	r, g, b = d.scaleColors(normal_r, normal_g, normal_b)

	return
}

func (d *Decoder) scaleColors(normal_r, normal_g, normal_b float64) (r, g, b uint32) {
	r = uint32(math.Min(math.Max(normal_r*d.maxValue, 0), d.maxValue))
	g = uint32(math.Min(math.Max(normal_g*d.maxValue, 0), d.maxValue))
	b = uint32(math.Min(math.Max(normal_b*d.maxValue, 0), d.maxValue))
	return
}

func NewDecoder(bytes_per_parameter, height, width uint32, chromaMode []byte, rawData *[]byte) *Decoder {
	bytesPerPixel := bytes_per_parameter * 2
	fmt.Printf("Width: %v, Height: %v, Bytes per parameter: %v\n", width, height, bytes_per_parameter)
	return &Decoder{
		Height:            height,
		Width:             width,
		FileSize:          0,
		bytesPerPixel:     bytesPerPixel,
		bytesPerParameter: bytes_per_parameter,
		chromaMode:        chromaMode,
		maxValue:          utils.GetMaxValue(bytes_per_parameter),
		rawData:           rawData,
		decodingBuffer:    bytes.NewBuffer(nil),
	}
}

var _ StreamProcessor = &Decoder{}
