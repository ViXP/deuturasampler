# Deuturasampler: Experimental Codec for Deuteranopia-Inspired Image Compression

## Project Idea

Deuturasampler is an experimental image codec inspired by the visual perception of people with deuteranopia. The core idea is to exploit the fact that, for individuals with deuteranopia, the red and green color components are less distinguishable. The averaging of these components, can reduce the image data size without significantly impacting perceived image quality for this audience (sort of, but the difference is still visible though).

This solution is written in **Go** programming language.

**Note:** This codec is not intended to be a scientifically accurate model for color remapping or removal as experienced by people with deuteranopia. Instead, it serves as a proof of concept to demonstrate that simple color averaging can lead to meaningful compression.

## How It Works

The codec works by averaging the red and green channels of an image, effectively reducing the amount of unique color information stored.

### Compression Modes

- **Default mode (4:4:4):**
  - All color channels are preserved, but red and green are averaged to reduce redundancy.
  - Example: Input file size = 1.1 MB, Encoded size = 707 KB

- **4:4:2 mode:**
  - Further reduces color information by subsampling the averaged red-green channel.
  - Example: Encoded size = 618 KB

- **4:4:1 mode:**
  - More aggressive subsampling of the averaged red-green channel.
  - Example: Encoded size = 574 KB

- **4:4:0 mode:**
  - Only the blue channel and a heavily subsampled red-green channel are stored.
  - Example: Encoded size = 530 KB

- **4:2:2 mode:**
  - Both the averaged red-green and blue channels are subsampled horizontally.
  - Example: Encoded size = 530 KB

- **4:2:1 mode:**
  - Even more subsampling of both channels.
  - Example: Encoded size = 486 KB

- **4:2:0 mode:**
  - Both channels are subsampled horizontally and vertically.
  - Example: Encoded size = 442 KB

- **4:1:1 mode:**
  - Maximum horizontal subsampling of both channels.
  - Example: Encoded size = 442 KB

- **4:1:0 mode:**
  - Maximum horizontal and vertical subsampling of both channels.
  - Example: Encoded size = 397 KB

## How to Use

There are two command-line applications:

- **encode**: Compresses a BMP image and saves it into the file with the same name and \*.deuimg file format.
- **decode**: Decompresses a .deuimg file back to a BMP image.

### Encoding

Run the encoder with:

    ./encode input_file.bmp Cb Crg

Where:

- `input_file.bmp` is your source BMP image.
- `Cb` and `Crg` specify the chroma subsampling mode (e.g., 4:4:4, 4:4:2, etc.). By default it is 4 4.

The encoded file (with extension `.deuimg`) will be saved in the same directory as the input file.

### Decoding

Run the decoder with:

    ./decode encoded_file.deuimg

The decoded BMP image will be saved in the same directory as the input `.deuimg` file, with the same name + `_decoded`.

### The examples

Here are the original file and the decoded version of the same picture, with the different levels
of subsampling during encoding:

| Input image                                                     | 4:4:4                                                                  | 4:4:2                                                                  | 4:4:0                                                                  | 4:2:2                                                                  | 4:2:0                                                                  | 4:1:1                                                                  | 4:1:0                                                                  |
| --------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| <img src="docs/input_file.bmp" title="Input file" width="120"/> | <img src="docs/input_file_decoded_444.bmp" title="4:4:4" width="120"/> | <img src="docs/input_file_decoded_442.bmp" title="4:4:2" width="120"/> | <img src="docs/input_file_decoded_440.bmp" title="4:4:0" width="120"/> | <img src="docs/input_file_decoded_422.bmp" title="4:2:2" width="120"/> | <img src="docs/input_file_decoded_420.bmp" title="4:2:0" width="120"/> | <img src="docs/input_file_decoded_411.bmp" title="4:1:1" width="120"/> | <img src="docs/input_file_decoded_410.bmp" title="4:1:0" width="120"/> |

---

## Algorithm Overview

1. **Input:** RGB bitmap image (BMP).
2. **Averaging:** For each pixel, computes the average of the red and green components.
3. **Encoding:**
   - In 4:4:4 mode, store the averaged red-green value and the blue channel for each pixel.
   - In 4:4:2 mode, subsample the averaged red-green channel (e.g., halve the horizontal resolution of this channel).
4. **Output:** Compressed image file with reduced color information and the extension \*.deuimg.

## Disclaimer

This codec is a minimal, experimental demonstration. It is not intended for production use or as a replacement for
scientifically validated color vision deficiency simulators or compressors. Its purpose is to explore the potential for
data reduction by leveraging perceptual redundancies in color vision.

---

**Author:** Cyril ViXP
**License:** MIT
