package imageobfuscation

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/cpmech/gosl/fun/fftw"

	"github.com/mjibson/go-dsp/dsputils"
)

// Image image class with FFT / DFT ability
type Image struct {
	rgbPlan2d [3]*fftw.Plan2d

	rgbMatrixes    [3]*dsputils.Matrix
	rgbFFTMatrixes [3]*dsputils.Matrix
	sync.RWMutex
	image.Image
}

const maxUint = math.MaxUint16

// FFT do FFT on image
func (m *Image) FFT(force bool) {
	m.Lock()
	defer m.Unlock()
	if !force && m.rgbPlan2d[0] != nil {
		return
	}

	img := m.Image
	min, max := img.Bounds().Min, img.Bounds().Max
	lenY, lenX := max.Y-min.Y, max.X-min.X
	scale := 1.0

	var imgColorData [3][]complex128
	for i := range imgColorData {
		imgColorData[i] = make([]complex128, lenX*lenY)
	}

	for i := 0; i < lenX; i++ {
		for j := 0; j < lenY; j++ {
			r, g, b, _ := img.At(i+min.X, j+min.Y).RGBA()
			imgColorData[0][i*lenY+j] = complex(scale*float64(r), 0)
			imgColorData[1][i*lenY+j] = complex(scale*float64(g), 0)
			imgColorData[2][i*lenY+j] = complex(scale*float64(b), 0)
		}
	}

	for i := range m.rgbPlan2d {
		if m.rgbPlan2d[i] != nil {
			m.rgbPlan2d[i].Free()
		}
		m.rgbPlan2d[i] = fftw.NewPlan2d(lenX, lenY, imgColorData[i], false, true)
		m.rgbPlan2d[i].Execute()
	}
}

// IFFT do inverse FFT
func (m *Image) IFFT() *Image {
	m.RLock()
	defer m.RUnlock()
	if m.rgbPlan2d[0] == nil {
		return m // do nothing when fft have not performed
	}
	newImage := image.NewRGBA64(m.Bounds())
	min, max := newImage.Bounds().Min, newImage.Bounds().Max
	lenY, lenX := max.Y-min.Y, max.X-min.X

	fftComplexData := make([]complex128, lenX*lenY)
	for colorOrder := 0; colorOrder < 3; colorOrder++ {
		// fill with fft data
		for i := range fftComplexData {
			fftComplexData[i] = m.rgbPlan2d[colorOrder].Get(i/lenY, i%lenY)
		}
		ifftPlan := fftw.NewPlan2d(lenX, lenY, fftComplexData, true, true)
		ifftPlan.Execute()
		for i := range fftComplexData {
			part := real(ifftPlan.Get(i/lenY, i%lenY)) / float64(lenX*lenY)
			if part > float64(maxUint) {
				fmt.Println(part)
				part = maxUint

			} else if part < 0 {
				part = 0
			}
			v := uint16(part)
			switch colorOrder {
			case 0: //red
				newImage.SetRGBA64(i/lenY, i%lenY, color.RGBA64{v, 0, 0, 0})
			case 1: //green
				rgba64 := newImage.RGBA64At(i/lenY, i%lenY)
				newImage.SetRGBA64(i/lenY, i%lenY, color.RGBA64{rgba64.R, v, 0, 0})
			case 2: //blue
				rgba64 := newImage.RGBA64At(i/lenY, i%lenY)
				_, _, _, alpha := m.Image.At(i/lenY, i%lenY).RGBA()
				newImage.SetRGBA64(i/lenY, i%lenY, color.RGBA64{rgba64.R, rgba64.G, v, uint16(alpha)})
			}
		}
		ifftPlan.Free()
	}
	return &Image{Image: newImage}
}
