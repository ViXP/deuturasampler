package deuimg

import (
	"bytes"
	"deuterasampler/internal/utils"
	"fmt"
	"math"
)

type Decoder struct {
	Height            uint32
	Width             uint32
	FileSize          uint32
	bytesPerPixel     uint32
	bytesPerParameter uint32
	maxValue          float64
	rawData           *[]byte
	decodingBuffer    *bytes.Buffer
}

func (d *Decoder) Process() []byte {
	d.write_header_to_buffer()

	bytes_padding := ((d.Width*d.bytesPerParameter*3+3)/4)*4 - d.Width*d.bytesPerParameter*3

	for row := uint32(0); row < d.Height; row++ {
		for col := uint32(0); col < d.Width; col++ {
			i := utils.FirstEncodedPixelByte + col*d.bytesPerPixel + row*d.Width*d.bytesPerPixel
			lum, cb := d.read_lum_cb((*d.rawData)[i : i+d.bytesPerPixel])
			r, g, b := d.generate_rgb(lum, cb)
			d.write_pixel(r, g, b)
		}
		d.write_padding(bytes_padding)
	}

	d.assign_file_size()

	decoded_data := d.decodingBuffer.Bytes()

	for i, b := range utils.DownsampleValue(d.FileSize, utils.Bits32) {
		decoded_data[2+i] = b
	}

	return decoded_data
}

func (d *Decoder) write_padding(bytes_padding uint32) {
	if bytes_padding == 0 {
		return
	}
	d.decodingBuffer.Write(make([]byte, bytes_padding))
}

func (d *Decoder) write_header_to_buffer() {
	header := make([]byte, utils.BmpHeaderSize)
	header[0] = byte('B')
	header[1] = byte('M')

	for i, b := range utils.DownsampleValue(utils.BmpHeaderSize, utils.Bits32) {
		header[10+i] = b
	}

	header[14] = byte(40)
	for i, b := range utils.DownsampleValue(d.Width, utils.Bits32) {
		header[18+i] = b
	}
	for i, b := range utils.DownsampleValue(d.Height, utils.Bits32) {
		header[22+i] = b
	}
	header[26] = 1
	for i, b := range utils.DownsampleValue(d.bytesPerParameter*3*8, utils.Bits16) {
		header[28+i] = b
	}
	d.decodingBuffer.Write(header)
}

func (d *Decoder) assign_file_size() {
	d.FileSize = uint32(d.decodingBuffer.Len())
}

func (d *Decoder) read_lum_cb(data []byte) (lum, cb uint32) {
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

func (d *Decoder) normalize_lum_chrome(luminance, cb uint32) (normal_lum, normal_cb float64) {
	normal_lum = float64(luminance) / d.maxValue
	normal_cb = float64(cb)/d.maxValue - 0.5

	return
}

func (d *Decoder) generate_rgb(lum, cb uint32) (r, g, b uint32) {
	var normal_b, normal_g, normal_r, normal_lum, normal_cb float64
	normal_lum, normal_cb = d.normalize_lum_chrome(lum, cb)

	normal_b = (normal_cb * 1.772) + normal_lum
	differentiator := 0.25 * (normal_b - normal_lum)
	normal_r = normal_lum - differentiator
	normal_g = normal_lum + differentiator

	r, g, b = d.scale_colors(normal_r, normal_g, normal_b)

	return
}

func (d *Decoder) scale_colors(normal_r, normal_g, normal_b float64) (r, g, b uint32) {
	r = uint32(math.Min(math.Max(normal_r*d.maxValue, 0), d.maxValue))
	g = uint32(math.Min(math.Max(normal_g*d.maxValue, 0), d.maxValue))
	b = uint32(math.Min(math.Max(normal_b*d.maxValue, 0), d.maxValue))
	return
}

func (d *Decoder) write_pixel(r, g, b uint32) {
	d.decodingBuffer.Write(utils.DownsampleValue(r, d.bytesPerParameter))
	d.decodingBuffer.Write(utils.DownsampleValue(g, d.bytesPerParameter))
	d.decodingBuffer.Write(utils.DownsampleValue(b, d.bytesPerParameter))
}

func NewDecoder(bytes_per_parameter, height, width uint32, raw_data *[]byte) *Decoder {
	bytes_per_pixel := bytes_per_parameter * 2
	fmt.Printf("Width: %v, Height: %v, Bytes per parameter: %v\n", width, height, bytes_per_parameter)
	return &Decoder{
		Height:            height,
		Width:             width,
		FileSize:          0,
		bytesPerPixel:     bytes_per_pixel,
		bytesPerParameter: bytes_per_parameter,
		maxValue:          utils.GetMaxValue(bytes_per_parameter),
		rawData:           raw_data,
		decodingBuffer:    bytes.NewBuffer(nil),
	}
}

var _ StreamProcessor = &Decoder{}
