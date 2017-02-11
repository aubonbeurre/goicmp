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
	"image/png"
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
	Output  string `short:"o" long:"output" description:"Output diff image" optional:"yes" optional-value:""`
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

type highlight16_func func(c1 color.RGBA64, c2 color.RGBA64) (same bool, r color.RGBA64)

func compare16(pic1 *image.RGBA64, pic2 *image.RGBA64, highlight highlight16_func) (stats DiffStats, result *image.RGBA64, err error) {
	stats.ExactSame = true
	stats.NumPixels = int64(pic1.Bounds().Dx()) * int64(pic1.Bounds().Dy())
	result = image.NewRGBA64(pic1.Bounds())
	for y := 0; y < pic1.Bounds().Dy(); y++ {
		for x := 0; x < pic1.Bounds().Dx(); x++ {
			var c1 color.RGBA64 = pic1.RGBA64At(x, y);
			var c2 color.RGBA64 = pic2.RGBA64At(x, y);

			same, r := highlight(c1, c2)

			if !same {
				stats.ExactSame = false
				stats.DiffPixels += 1
			}
			result.SetRGBA64(x, y, r);
		}
	}

	return stats, result, nil
}

func delta16(x uint16, y uint16) uint {
	if x < y {
		return uint(y - x)
	}
	return uint(x - y)
}

func samePixel16(c1 color.RGBA64, c2 color.RGBA64) (same bool, delta uint16) {
	if c1.A == 0 && c2.A == 0 {
		return true, 0
	}
	same = c1.R == c2.R && c1.G == c2.G && c1.B == c2.B && c1.A == c2.A
	if !same {
		delta = uint16((delta16(c1.R, c2.R) + delta16(c1.G, c2.G) + delta16(c1.B, c2.B) + delta16(c1.A, c2.A)) / 4)
		//fmt.Printf("delta: %d\n", delta)
		if delta <= 255 {
			same = true
		}
	}
	return same, delta
}

func highlight_distance16(c1 color.RGBA64, c2 color.RGBA64) (same bool, r color.RGBA64) {
	same, _ = samePixel16(c1, c2)

	if !same {
		r.R = 65535
		r.G = 0
		r.B = 0
		r.A = 65535
	} else {
		r.R = 0
		r.G = 0
		r.B = 0
		r.A = 0
	}
	return same, r
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
		var result image.Image
		if stats, result, err = compare16(gImage1, gImage2, highlight_distance16); err != nil {
			panic(err.Error())
		}

		if len(gOpts.Output) > 0 {
			var f *os.File
			if f, err = os.OpenFile(gOpts.Output, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
				panic(err.Error())
			}
			if err = png.Encode(f, result); err != nil {
				panic(err.Error())
			}
			if len(gOpts.Verbose) > 0 {
				fmt.Printf("Output file created: %s\n", gOpts.Output)
			}
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
