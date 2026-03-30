package main

import (
	"deuterasampler/internal/deuimg"
	"deuterasampler/internal/factories"
	"deuterasampler/internal/parsers"
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
	headerParser := parsers.BitmapHeaderParser{Data: data}
	imageMetadata := headerParser.Parse()
	imageMetadata.ChromaMode = chromaMode

	return factories.NewDeuImgFileFactory(imageMetadata).Create(deuimg.NewEncoder(imageMetadata, data).Process())
}
