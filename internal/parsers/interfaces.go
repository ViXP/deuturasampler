package parsers

type HeaderParser interface {
	Parse() *ImageMetadata
}

type ImageMetadata struct {
	Height            uint32
	Width             uint32
	BytesPerParameter uint32
	ChromaMode        []byte
}
