package deuimg

import (
	"bytes"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
)

type Encoder struct {
	Height           uint32
	Width            uint32
	FileSize         uint32
	bytesPerPixel    uint32
	bytesPerColor    uint32
	maxValue         float64
	rawData          *[]byte
	processingBuffer *bytes.Buffer
}

func (e *Encoder) Process() []byte {
	e.write_header()

	bytesPerWidth := ((e.Width*e.bytesPerPixel + 3) / 4) * 4

	for row := uint32(0); row < e.Height; row++ {
		for col := uint32(0); col < e.Width; col++ {
			position := utils.BmpHeaderSize + row*bytesPerWidth + col*e.bytesPerPixel
			pixel_data := (*e.rawData)[position : position+e.bytesPerPixel]

			r, g, b := e.read_rgb(pixel_data)
			lum, cb := e.generate_lum_chrome(r, g, b)

			e.processingBuffer.Write(utils.DownsampleValue(lum, e.bytesPerColor))
			e.processingBuffer.Write(utils.DownsampleValue(cb, e.bytesPerColor))
		}
	}

	return e.processingBuffer.Bytes()
}

func (e *Encoder) generate_lum_chrome(r, g, b uint32) (lum, cb uint32) {
	normal_r, normal_g, normal_b := e.normalize_colors(r, g, b)
	normal_lum := 0.2126*normal_r + 0.7152*normal_g + 0.114*normal_b
	normal_cb := (normal_b - normal_lum) / 1.772

	lum, cb = e.scale_lum_chrome(normal_lum, normal_cb)

	return
}

func (e *Encoder) scale_lum_chrome(normal_lum, normal_cb float64) (lum, cb uint32) {
	normal_lum = math.Min(math.Max(normal_lum, 0), 1)
	normal_cb = math.Min(math.Max(normal_cb, -0.5), 0.5)

	lum = uint32(math.Round(normal_lum * e.maxValue))
	cb = uint32(math.Round((normal_cb + 0.5) * e.maxValue))
	return
}

func (e *Encoder) normalize_colors(r, g, b uint32) (normal_r, normal_g, normal_b float64) {
	normal_r = float64(r) / e.maxValue
	normal_g = float64(g) / e.maxValue
	normal_b = float64(b) / e.maxValue
	return
}

func (e *Encoder) read_rgb(bytes_stream []byte) (r, g, b uint32) {
	switch e.bytesPerColor {
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
		panic(fmt.Sprintf("The %vbits/color space is not supported!", e.bytesPerColor*8))
	}

	return
}

func (e *Encoder) write_header() {
	e.processingBuffer.WriteString("(-_-)")
	e.processingBuffer.Write(utils.DownsampleValue(e.Width, utils.Bits32))
	e.processingBuffer.Write(utils.DownsampleValue(e.Height, utils.Bits32))
	e.processingBuffer.Write(utils.DownsampleValue(e.bytesPerColor, utils.Bits32))
}

func NewEncoder(bytes_per_pixel, height, width uint32, raw_data *[]byte) *Encoder {
	bytes_per_color := bytes_per_pixel / 3

	fmt.Printf("Width: %v, Height: %v, Bytes per parameter: %v\n", width, height, bytes_per_color)
	return &Encoder{
		bytesPerPixel:    bytes_per_pixel,
		bytesPerColor:    bytes_per_color,
		Height:           height,
		Width:            width,
		FileSize:         uint32(len(*raw_data)),
		maxValue:         utils.GetMaxValue(bytes_per_color),
		rawData:          raw_data,
		processingBuffer: bytes.NewBuffer(nil),
	}
}

var _ StreamProcessor = &Encoder{}
