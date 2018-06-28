package imageobfuscation

import (
	stdimage "image"
	"image/png"
	"os"
)

// Obfuscate makes image obfuscation on inputFileName to outputFileName
func Obfuscate(inputFileName, outputFileName string) error {
	imageFile, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer imageFile.Close()

	image, _, err := stdimage.Decode(imageFile)
	if err != nil {
		return err
	}

	img := &Image{Image: image}
	img.FFT(false)
	// ffts := img.GetFFT(false)
	// for _, fft := range ffts {
	// 	dim := fft.Dimensions()
	// 	height := dim[0]
	// 	width := dim[1]
	// 	for i := 0; i < 100000; i++ {
	// 		y := rand.Intn(height)
	// 		x := rand.Intn(width)
	// 		point := []int{y, x}
	// 		point2 := []int{height - 1 - y, width - 1 - x}
	// 		fft.SetValue(complex(real(fft.Value(point))*10.0000001, imag(fft.Value(point))), point)
	// 		fft.SetValue(complex(real(fft.Value(point2))*10.0000001, imag(fft.Value(point2))), point2)
	// 	}
	// }

	newimage := img.IFFT()

	outfile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return png.Encode(outfile, newimage.Image)
}
