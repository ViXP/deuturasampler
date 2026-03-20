package main

import (
	"deuterasampler/internal/deuimg"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

func main() {
	var chromaMode []byte

	switch len(os.Args) {
	case 1:
		panic("select image!")
	case 2, 3:
		chromaMode = []byte{4, 4}
	case 4:
		i, err := strconv.ParseInt(os.Args[2], 10, 8)

		if err != nil {
			panic(err)
		}

		j, err := strconv.ParseInt(os.Args[3], 10, 8)

		if err != nil {
			panic(err)
		}

		chromaMode = []byte{byte(i), byte(j)}
	}

	encode(os.Args[1], chromaMode)
}

func encode(path string, chromaMode []byte) {
	original, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	os.WriteFile(strings.Split(path, ".")[0]+".deuimg", encodeBmp(original, chromaMode), 0666)
}

func encodeBmp(data []byte, chromaMode []byte) []byte {
	var compression, width, height, bytesPerPixel uint32

	if data[0] != 'B' || data[1] != 'M' {
		panic("this is incorrect BMP file!")
	}

	compression = binary.LittleEndian.Uint32(data[30:34])

	if compression != 0 {
		panic("the file is already compressed!")
	}

	width = binary.LittleEndian.Uint32(data[18:22])
	height = binary.LittleEndian.Uint32(data[22:26])
	bytesPerPixel = uint32(binary.LittleEndian.Uint16(data[28:30])) / 8

	encoder := deuimg.NewEncoder(bytesPerPixel, height, width, chromaMode, &data)

	return encoder.Process()
}
