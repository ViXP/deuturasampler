package parsers

import (
	"encoding/binary"
)

type DeuImgHeaderParser struct {
	Data []byte
}

func (d *DeuImgHeaderParser) Parse() *ImageMetadata {
	if d.Data[2] != '_' {
		panic("this file is not encoded properly!")
	}

	return &ImageMetadata{
		Width:             binary.LittleEndian.Uint32(d.Data[5:9]),
		Height:            binary.LittleEndian.Uint32(d.Data[9:13]),
		BytesPerParameter: binary.LittleEndian.Uint32(d.Data[13:17]),
		ChromaMode:        d.Data[17:19],
	}
}

var _ HeaderParser = &DeuImgHeaderParser{}
