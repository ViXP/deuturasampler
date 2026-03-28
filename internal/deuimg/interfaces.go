package deuimg

type StreamProcessor interface {
	Process() []byte
}

type HeaderParser interface {
	Parse() (width, height, bytesPerParameter uint32)
}
