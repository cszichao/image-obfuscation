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

// ImageType image type
type ImageType int

const (
	// ImageTypePNG png
	ImageTypePNG ImageType = iota
	// ImageTypeJPG jpeg
	ImageTypeJPG
	// ImageTypeGIF gif
	ImageTypeGIF
	// ImageTypeNotSupported not supported
	ImageTypeNotSupported
)

// ErrImageNotSupported image not supportted
var ErrImageNotSupported = errors.New("image type not supported")

// ErrAnimatedImageNotSupported animated image not supported
var ErrAnimatedImageNotSupported = errors.New("animated image not supported")

// Obfuscate makes image obfuscation on inputFileName to outputFileName
func Obfuscate(inputFileName, outputFileName string) (ImageType, error) {
	var imgType ImageType
	switch strings.ToLower(path.Ext(inputFileName)) {
	case ".png":
		imgType = ImageTypePNG
	case ".jpg":
		fallthrough
	case ".jpeg":
		imgType = ImageTypeJPG
	case ".gif":
		imgType = ImageTypeGIF
	default:
		return ImageTypeNotSupported, ErrImageNotSupported
	}
	image, err := func() (stdimage.Image, error) {
		imageFile, err := os.Open(inputFileName)
		if err != nil {
			return nil, err
		}
		defer imageFile.Close()
		var image stdimage.Image
		switch imgType {
		case ImageTypePNG:
			image, err = png.Decode(imageFile)
		case ImageTypeJPG:
			image, err = jpeg.Decode(imageFile)
		case ImageTypeGIF:
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
		return imgType, err
	}

	img := &Image{Image: image}
	defer img.Destory()

	img.FFT(false)
	min, max := img.Bounds().Min, img.Bounds().Max
	lenY, lenX := max.Y-min.Y, max.X-min.X
	rand.Seed(time.Now().UTC().UnixNano())
	for colorOrder := 0; colorOrder < 3; colorOrder++ {
		maxObsPoints := 10000
		if lenX*lenY/4 < maxObsPoints {
			maxObsPoints = lenX * lenY / 4
		}
		for i := 0; i < maxObsPoints; i++ {
			y := lenY/4 + rand.Intn(lenY/2)
			x := lenX/4 + rand.Intn(lenX/2)
			v, _ := img.GetFFT(colorOrder, x, y)
			img.SetFFT(colorOrder, x, y, complex(real(v)*1.0001, imag(v)*1.0001))
		}
	}
	newimage := img.IFFT()

	outfile, err := os.Create(outputFileName)
	if err != nil {
		return imgType, err
	}
	defer outfile.Close()
	switch imgType {
	case ImageTypePNG:
		return imgType, png.Encode(outfile, newimage.Image)
	case ImageTypeJPG:
		return imgType, jpeg.Encode(outfile, newimage.Image, &jpeg.Options{Quality: 75})
	case ImageTypeGIF:
		return imgType, gif.Encode(outfile, newimage.Image, &gif.Options{NumColors: 256})
	}
	return imgType, nil
}
