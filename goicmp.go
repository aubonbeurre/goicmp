package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"image"
	"image/draw"
	"image/color"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	_ "image/png"
	"math"
)

var (
	gImage1 *image.RGBA64
	gImage2 *image.RGBA64

	gDiffFlag bool = false
)

var gOpts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

type DiffStats struct {
	NumPixels  int64
	DiffPixels int64
	ExactSame  bool
}

func (p *DiffStats) Report() {
	fmt.Println("Diff report:")
	if p.ExactSame {
		fmt.Println("\tExact match")
	} else {
		fmt.Println("\tImages differ")
	}
	fmt.Printf("\tDifferent Pixels %%: %.2f%%\n", float64(p.DiffPixels)/float64(p.NumPixels)*100.0)
	fmt.Printf("\tDifferent Pixels #: %d\n", p.DiffPixels)
}

func compare16(pic1 *image.RGBA64, pic2 *image.RGBA64) (stats DiffStats, err error) {
	stats.ExactSame = true
	stats.NumPixels = int64(pic1.Bounds().Dx()) * int64(pic1.Bounds().Dy())
	var KERNEL int = 3
	var Threshold uint32 = 16
	for y := 0; y < (pic1.Bounds().Dy() - KERNEL); y++ {
		for x := 0; x < (pic1.Bounds().Dx() - KERNEL); x++ {
			var r1, g1, b1, a1 uint32
			r1, g1, b1, a1 = 0, 0, 0, 0

			for i := 0; i <= KERNEL; i++ {
				for j := 0; j <= KERNEL; j++ {
					var c color.RGBA64 = pic1.RGBA64At(x + i, y + i);
					r0, g0, b0, a0 := c.RGBA()
					r1 += r0
					g1 += g0
					b1 += b0
					a1 += a0
				}
			}

			r1 >>= 12
			g1 >>= 12
			b1 >>= 12
			a1 >>= 12

			var r2, g2, b2, a2 uint32
			r2, g2, b2, a2 = 0, 0, 0, 0

			for i := 0; i <= KERNEL; i++ {
				for j := 0; j <= KERNEL; j++ {
					var c color.RGBA64 = pic2.RGBA64At(x + i, y + i);
					r0, g0, b0, a0 := c.RGBA()
					r2 += r0
					g2 += g0
					b2 += b0
					a2 += a0
				}
			}

			r2 >>= 12
			g2 >>= 12
			b2 >>= 12
			a2 >>= 12

			var diff uint32 = uint32(math.Abs(float64(r1) - float64(r2)))
			diff += uint32(math.Abs(float64(g1) - float64(g2)))
			diff += uint32(math.Abs(float64(b1) - float64(b2)))
			diff += uint32(math.Abs(float64(a1) - float64(a2)))

			diff >>= 2
			same := diff <= Threshold

			if !same {
				stats.ExactSame = false
				stats.DiffPixels += 1
			}
		}
	}

	return stats, nil
}

func downloadImage(url string) (err error, path string) {
	var f *os.File
	if f, err = ioutil.TempFile("", ""); err != nil {
		return err, ""
	}

	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return err, ""
	}
	defer resp.Body.Close()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err, ""
	}

	f.Close()
	return nil, f.Name()
}

func NewTexture(file string) (err error, rgba *image.RGBA64) {
	var imgFile *os.File
	if imgFile, err = os.Open(file); err != nil {
		return err, nil
	}
	defer imgFile.Close()

	var img image.Image
	if img, _, err = image.Decode(imgFile); err != nil {
		return err, nil
	}

	rgba = image.NewRGBA64(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return nil, rgba
}

func main() {
	var parser = flags.NewParser(&gOpts, flags.Default)

	var err error
	var args []string
	if args, err = parser.Parse(); err != nil {
		os.Exit(1)
	}

	if len(args) < 1 || len(args) > 2 {
		//panic(fmt.Errorf("Too many or not enough arguments"))
	}

	gDiffFlag = len(args) == 2

	var image1_path string = args[0]
	if strings.HasPrefix(image1_path, "http") {
		if err, image1_path = downloadImage(image1_path); err != nil {
			panic(err)
		}
	}

	if err, gImage1 = NewTexture(image1_path); err != nil {
		panic(err)
	}

	if gDiffFlag {
		var image2_path string = args[1]
		if strings.HasPrefix(image2_path, "http") {
			if err, image2_path = downloadImage(image2_path); err != nil {
				panic(err)
			}
		}

		if err, gImage2 = NewTexture(image2_path); err != nil {
			panic(err)
		}

		if gImage1.Bounds().Size().X != gImage2.Bounds().Size().X || gImage1.Bounds().Size().Y != gImage2.Bounds().Size().Y {
			fmt.Println("WARNING: image dimensions differ!")
		} else {
			if len(gOpts.Verbose) > 0 {
				fmt.Printf("image dimensions: %dx%d\n", gImage1.Bounds().Size().X, gImage1.Bounds().Size().Y)
			}
		}

		var stats DiffStats
		if stats, err = compare16(gImage1, gImage2); err != nil {
			panic(err.Error())
		}

		if len(gOpts.Verbose) > 0 {
			stats.Report()
		}

		if stats.ExactSame {
			os.Exit(0)
		} else {
			os.Exit(98)
		}
	} else {
		if len(gOpts.Verbose) > 0 {
			fmt.Printf("image dimensions: %dx%d\n", gImage1.Bounds().Size().X, gImage1.Bounds().Size().Y)
		}
	}
}
