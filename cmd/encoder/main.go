package main

import (
	"deuterasampler/internal/deuimg"
	"encoding/binary"
	"os"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		panic("select image!")
	}

	encode(os.Args[1])
}

func encode(path string) {
	original, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	os.WriteFile(strings.Split(path, ".")[0]+".deuimg", encodeBmp(original), 0666)
}

func encodeBmp(data []byte) []byte {
	var compression, width, height, bytes_per_pixel uint32

	if data[0] != 'B' || data[1] != 'M' {
		panic("this is incorrect BMP file!")
	}

	compression = binary.LittleEndian.Uint32(data[30:34])

	if compression != 0 {
		panic("the file is already compressed!")
	}

	width = binary.LittleEndian.Uint32(data[18:22])
	height = binary.LittleEndian.Uint32(data[22:26])
	bytes_per_pixel = uint32(binary.LittleEndian.Uint16(data[28:30])) / 8

	encoder := deuimg.NewEncoder(bytes_per_pixel, height, width, &data)

	return encoder.Process()
}
