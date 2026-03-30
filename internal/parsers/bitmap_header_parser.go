package parsers

import "encoding/binary"

type BitmapHeaderParser struct {
	Data []byte
}

func (b *BitmapHeaderParser) Parse() *ImageMetadata {
	if b.Data[0] != 'B' || b.Data[1] != 'M' {
		panic("this is incorrect BMP file!")
	}

	if binary.LittleEndian.Uint32(b.Data[30:34]) != 0 {
		panic("the BMP file is already compressed!")
	}

	return &ImageMetadata{
		Width:             binary.LittleEndian.Uint32(b.Data[18:22]),
		Height:            binary.LittleEndian.Uint32(b.Data[22:26]),
		BytesPerParameter: uint32(binary.LittleEndian.Uint16(b.Data[28:30])) / 24,
	}
}

var _ HeaderParser = &BitmapHeaderParser{}
