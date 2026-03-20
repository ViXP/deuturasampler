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

	decode(os.Args[1])
}

func decode(path string) {
	encoded, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	os.WriteFile(strings.Split(path, ".")[0]+"_decoded.bmp", decodeDEUImg(encoded), 0666)
}

func decodeDEUImg(data []byte) []byte {
	var width, height, bytesPerParameter uint32
	var chromaMode []byte

	if data[2] != '_' {
		panic("this file is not encoded properly!")
	}

	width = binary.LittleEndian.Uint32(data[5:9])
	height = binary.LittleEndian.Uint32(data[9:13])
	bytesPerParameter = binary.LittleEndian.Uint32(data[13:17])
	chromaMode = data[17:18]

	decoder := deuimg.NewDecoder(bytesPerParameter, height, width, chromaMode, &data)

	return decoder.Process()
}
