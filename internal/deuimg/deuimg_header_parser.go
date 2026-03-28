package deuimg

import (
	"encoding/binary"
)

type DeuImgHeaderParser struct {
	Data []byte
}

func (d *DeuImgHeaderParser) Parse() (width uint32, height uint32, bytesPerParameter uint32) {
	if d.Data[2] != '_' {
		panic("this file is not encoded properly!")
	}

	width = binary.LittleEndian.Uint32(d.Data[5:9])
	height = binary.LittleEndian.Uint32(d.Data[9:13])
	bytesPerParameter = binary.LittleEndian.Uint32(d.Data[13:17])
	return
}

func (d *DeuImgHeaderParser) ParseChromaMode() []byte {
	return d.Data[17:19]
}

var _ HeaderParser = &DeuImgHeaderParser{}
