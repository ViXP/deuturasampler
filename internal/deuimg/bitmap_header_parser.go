package deuimg

import "encoding/binary"

type BitmapHeaderParser struct {
	Data []byte
}

func (b *BitmapHeaderParser) Parse() (width uint32, height uint32, bytesPerParameter uint32) {
	if b.Data[0] != 'B' || b.Data[1] != 'M' {
		panic("this is incorrect BMP file!")
	}

	if binary.LittleEndian.Uint32(b.Data[30:34]) != 0 {
		panic("the BMP file is already compressed!")
	}

	width = binary.LittleEndian.Uint32(b.Data[18:22])
	height = binary.LittleEndian.Uint32(b.Data[22:26])
	bytesPerParameter = uint32(binary.LittleEndian.Uint16(b.Data[28:30])) / 24
	return
}

var _ HeaderParser = &BitmapHeaderParser{}
