package main

import (
	"deuterasampler/internal/deuimg"
	"deuterasampler/internal/factories"
	"deuterasampler/internal/parsers"
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
	headerParser := parsers.DeuImgHeaderParser{Data: data}
	imageMetadata := headerParser.Parse()

	return factories.NewBitmapFactory(imageMetadata).Create(deuimg.NewDecoder(imageMetadata, data).Process())
}
