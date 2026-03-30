package factories

import (
	"bytes"
	"deuterasampler/internal/parsers"
	"deuterasampler/internal/utils"
)

type BitmapFileFactory struct {
	buffer   *bytes.Buffer
	metadata *parsers.ImageMetadata
}

func (b *BitmapFileFactory) Create(rowsData [][]byte) []byte {
	b.writeBitmapHeader()

	for _, data := range rowsData {
		b.buffer.Write(data)
	}

	decodedData := b.buffer.Bytes()
	utils.WriteBytes(decodedData[2:6], uint32(len(decodedData)), utils.Bits32)

	return decodedData
}
func (b *BitmapFileFactory) writeBitmapHeader() {
	header := make([]byte, utils.BmpHeaderSize)
	header[0] = byte('B')
	header[1] = byte('M')

	utils.WriteBytes(header[10:14], utils.BmpHeaderSize, utils.Bits32)
	header[14] = byte(40)
	utils.WriteBytes(header[18:22], b.metadata.Width, utils.Bits32)
	utils.WriteBytes(header[22:26], b.metadata.Height, utils.Bits32)
	header[26] = 1
	utils.WriteBytes(header[28:30], b.metadata.BytesPerParameter*3*8, utils.Bits16)
	b.buffer.Write(header)
}

func NewBitmapFactory(metadata *parsers.ImageMetadata) *BitmapFileFactory {
	return &BitmapFileFactory{buffer: bytes.NewBuffer(nil), metadata: metadata}
}

var _ FileFactorable = &BitmapFileFactory{}
