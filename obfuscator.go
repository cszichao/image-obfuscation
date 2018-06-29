package imageobfuscation

import (
	stdimage "image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"v2ray.com/core/common/errors"
)

type imageType int

const (
	imageTypePNG imageType = iota
	imageTypeJPG
	imageTypeGIF
)

// ErrImageNotSupported image not supportted
var ErrImageNotSupported = errors.New("image type not supported")

// ErrAnimatedImageNotSupported animated image not supported
var ErrAnimatedImageNotSupported = errors.New("animated image not supported")

// Obfuscate makes image obfuscation on inputFileName to outputFileName
func Obfuscate(inputFileName, outputFileName string) error {
	var imgType imageType
	switch strings.ToLower(path.Ext(inputFileName)) {
	case ".png":
		imgType = imageTypePNG
	case ".jpg":
		fallthrough
	case ".jpeg":
		imgType = imageTypeJPG
	case ".gif":
		imgType = imageTypeGIF
	default:
		return ErrImageNotSupported
	}
	image, err := func() (stdimage.Image, error) {
		imageFile, err := os.Open(inputFileName)
		if err != nil {
			return nil, err
		}
		defer imageFile.Close()
		var image stdimage.Image
		switch imgType {
		case imageTypePNG:
			image, err = png.Decode(imageFile)
		case imageTypeJPG:
			image, err = jpeg.Decode(imageFile)
		case imageTypeGIF:
			gifs, err2 := gif.DecodeAll(imageFile)
			if err2 != nil {
				err = err2
			} else if len(gifs.Image) != 1 {
				err = ErrAnimatedImageNotSupported
			} else {
				image = gifs.Image[0]
			}

		}

		if err != nil {
			return nil, err
		}
		return image, err
	}()
	if err != nil {
		return err
	}

	img := &Image{Image: image}
	defer img.Destory()

	img.FFT(false)
	min, max := img.Bounds().Min, img.Bounds().Max
	lenY, lenX := max.Y-min.Y, max.X-min.X
	rand.Seed(time.Now().UTC().UnixNano())
	for colorOrder := 0; colorOrder < 3; colorOrder++ {
		for i := 0; i < 10000; i++ {
			y := rand.Intn(lenY)
			x := rand.Intn(lenX)
			v, _ := img.GetFFT(colorOrder, x, y)
			img.SetFFT(colorOrder, x, y, complex(real(v)*1.001, imag(v)*1.001))
		}
	}
	newimage := img.IFFT()

	outfile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outfile.Close()
	switch imgType {
	case imageTypePNG:
		return png.Encode(outfile, newimage.Image)
	case imageTypeJPG:
		return jpeg.Encode(outfile, newimage.Image, &jpeg.Options{Quality: 75})
	case imageTypeGIF:
		return gif.Encode(outfile, newimage.Image, &gif.Options{NumColors: 256})
	}
	return nil
}
