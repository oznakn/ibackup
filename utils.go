package main

import (
	"bufio"
	"bytes"
	"github.com/disintegration/imageorient"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"mime/multipart"
)

func compressIfPossible(file multipart.File, ext string) ([]byte, error) {
	img, _, err := imageorient.Decode(file)

	if err != nil {
		return nil, err
	}

	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		buffer, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		return buffer, nil
	}

	var buffer bytes.Buffer
	dst := bufio.NewWriter(&buffer)

	if ext == ".jpg" || ext == ".jpeg" {
		err = jpeg.Encode(dst, img, &jpeg.Options{Quality: 40})

		if err != nil {
			return nil, err
		}
	} else if ext == ".png" {
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(dst, img)

		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}
