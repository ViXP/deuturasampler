package main

import (
	"deuterasampler/internal/deuimg"
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
	headerParser := deuimg.DeuImgHeaderParser{Data: data}
	width, height, bytesPerParameter := headerParser.Parse()
	decoder := deuimg.NewDecoder(bytesPerParameter, height, width, headerParser.ParseChromaMode(), data)

	return decoder.Process()
}
