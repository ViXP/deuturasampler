package factories

import (
	"bytes"
	"deuterasampler/internal/parsers"
	"deuterasampler/internal/utils"
)

type DeuImgFileFactory struct {
	buffer   *bytes.Buffer
	metadata *parsers.ImageMetadata
}

func (d *DeuImgFileFactory) Create(rowsData [][]byte) []byte {
	d.writeHeader()

	for _, row := range rowsData {
		d.buffer.Write(row)
	}

	return d.buffer.Bytes()
}

func (d *DeuImgFileFactory) writeHeader() {
	header := make([]byte, utils.DeuimgHeaderSize)

	for i, sym := range "(-_-)" {
		header[i] = byte(sym)
	}

	utils.WriteBytes(header[5:9], d.metadata.Width, utils.Bits32)
	utils.WriteBytes(header[9:13], d.metadata.Height, utils.Bits32)
	utils.WriteBytes(header[13:17], d.metadata.BytesPerParameter, utils.Bits32)
	header[17] = d.metadata.ChromaMode[0]
	header[18] = d.metadata.ChromaMode[1]
	d.buffer.Write(header)
}

func NewDeuImgFileFactory(metadata *parsers.ImageMetadata) *DeuImgFileFactory {
	return &DeuImgFileFactory{buffer: bytes.NewBuffer(nil), metadata: metadata}
}

var _ FileFactorable = &DeuImgFileFactory{}
