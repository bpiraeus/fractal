package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Configuration
const (
	xCenter = -0.75
	yCenter =  0
	zoom = 1

	// Quality
	imgWidth     = 1024
	imgHeight    = 1024
	maxIter      = 1500
	samples      = 50
	linearMixing = true

	showProgress = true
	profileCpu   = false
)

const (
	ratio = float64(imgWidth) / float64(imgHeight)
)

func main() {
	log.Println("Allocating image...")
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	log.Println("Rendering...")
	start := time.Now()
	render(img)
	end := time.Now()

	log.Println("Done rendering in", end.Sub(start))

	log.Println("Encoding image...")
	f, err := os.Create("result.png")
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
	log.Println("Done!")
}

func render(img *image.RGBA) {
	if profileCpu {
		f, err := os.Create("profile.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	jobs := make(chan int)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func () {
			for y := range jobs {
				for x := 0; x < imgWidth; x++ {
					var r, g, b int
					for i := 0; i < samples; i++ {
						nx := 3*(1/zoom)*ratio*((float64(x)+RandFloat64())/float64(imgWidth)-0.5) + xCenter
						ny := 3*(1/zoom)*((float64(y)+RandFloat64())/float64(imgHeight)-0.5) - yCenter
						c := paint(mandelbrotIterComplex(nx, ny, maxIter))
						if linearMixing {
							r += int(RGBToLinear(c.R))
							g += int(RGBToLinear(c.G))
							b += int(RGBToLinear(c.B))
						} else {
							r += int(c.R)
							g += int(c.G)
							b += int(c.B)
						}
					}
					var cr, cg, cb uint8
					if linearMixing {
						cr = LinearToRGB(uint16(float64(r) / float64(samples)))
						cg = LinearToRGB(uint16(float64(g) / float64(samples)))
						cb = LinearToRGB(uint16(float64(b) / float64(samples)))
					} else {
						cr = uint8(float64(r) / float64(samples))
						cg = uint8(float64(g) / float64(samples))
						cb = uint8(float64(b) / float64(samples))
					}
					img.SetRGBA(x, y, color.RGBA{ R: cr, G: cg, B: cb, A: 255 })
				}
			}
		}()
	}

	for y := 0; y < imgHeight; y++ {
		jobs <- y
		if showProgress {
			fmt.Printf("\r%d/%d (%d%%)", y, imgHeight, int(100*(float64(y) / float64(imgHeight))))
		}
	}
	if showProgress {
		fmt.Printf("\r%d/%[1]d (100%%)\n", imgHeight)
	}
}

func paint(r float64, n int) color.RGBA {
	insideSet := color.RGBA{ R: 255, G: 255, B: 255, A: 255 }

	if r > 4 {
		return hslToRGB(float64(n) / 800 * r, 1, 0.5)
	}

	return insideSet
}

func mandelbrotIterComplex(px, py float64, maxIter int) (float64, int) {
	var current complex128
	pxpy := complex(px, py)

	for i := 0; i < maxIter; i++ {
		magnitude := cmplx.Abs(current)
		if magnitude > 2 {
			return magnitude * magnitude, i
		}
		current = current * current + pxpy
	}

	magnitude := cmplx.Abs(current)
	return magnitude * magnitude, maxIter
}