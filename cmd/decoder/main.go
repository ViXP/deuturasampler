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
	var width, height, bytes_per_parameter uint32

	if data[2] != '_' {
		panic("this file is not encoded properly!")
	}

	width = binary.LittleEndian.Uint32(data[5:9])
	height = binary.LittleEndian.Uint32(data[9:13])
	bytes_per_parameter = binary.LittleEndian.Uint32(data[13:17])

	decoder := deuimg.NewDecoder(bytes_per_parameter, height, width, &data)

	return decoder.Process()
}
