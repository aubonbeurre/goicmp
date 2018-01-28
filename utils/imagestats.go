package utils

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// DiffStats ...
type DiffStats struct {
	NumPixels  int64 `json:"numpixels"`
	DiffPixels int64 `json:"diffpixels"`
	ExactSame  bool  `json:"exactsame"`
}

// Report ...
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

// DownloadImage ...
func DownloadImage(url string) (path string, err error) {
	var f *os.File
	if f, err = ioutil.TempFile("", ""); err != nil {
		return "", err
	}
	defer f.Close()

	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	return f.Name(), nil
}

// NewImage ...
func NewImage(file string) (rgba *image.RGBA64, err error) {
	var imgFile *os.File
	if imgFile, err = os.Open(file); err != nil {
		return nil, err
	}
	defer imgFile.Close()

	var img image.Image
	if img, _, err = image.Decode(imgFile); err != nil {
		return nil, err
	}

	rgba = image.NewRGBA64(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return rgba, nil
}

// DownloadOrLoadImage ...
func DownloadOrLoadImage(image1path string) (rgba *image.RGBA64, err error) {

	if strings.HasPrefix(image1path, "http") {
		if image1path, err = DownloadImage(image1path); err != nil {
			return nil, fmt.Errorf("Error reading '%s' (%v)", image1path, err)
		}
	}

	if rgba, err = NewImage(image1path); err != nil {
		return nil, fmt.Errorf("Error reading '%s' (%v)", image1path, err)
	}
	return rgba, err
}
