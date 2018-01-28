package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"strings"

	"github.com/aubonbeurre/goicmp/utils"
	"github.com/jessevdk/go-flags"
)

var (
	gImage1 *image.RGBA64
	gImage2 *image.RGBA64

	gDiffFlag = false
)

var gOpts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	// Example of a value name
	Batch string `short:"b" long:"batch" description:"A json file with a list of diff" value-name:"BATCHFILE"`
}

func main() {
	var parser = flags.NewParser(&gOpts, flags.Default)

	var err error
	var args []string
	if args, err = parser.Parse(); err != nil {
		os.Exit(1)
	}

	if len(gOpts.Batch) > 0 {
		var exitcode int
		if exitcode, err = utils.RunDiffBatch(gOpts.Batch); err != nil {
			panic(err)
		}
		os.Exit(exitcode)
	}

	if len(args) < 1 || len(args) > 2 {
		//panic(fmt.Errorf("Too many or not enough arguments"))
	}

	gDiffFlag = len(args) == 2

	var image1path = args[0]
	if strings.HasPrefix(image1path, "http") {
		if image1path, err = utils.DownloadImage(image1path); err != nil {
			panic(err)
		}
	}

	if gImage1, err = utils.NewImage(image1path); err != nil {
		panic(err)
	}

	if gDiffFlag {
		var image2path = args[1]
		if strings.HasPrefix(image2path, "http") {
			if image2path, err = utils.DownloadImage(image2path); err != nil {
				panic(err)
			}
		}

		if gImage2, err = utils.NewImage(image2path); err != nil {
			panic(err)
		}

		if gImage1.Bounds().Size().X != gImage2.Bounds().Size().X || gImage1.Bounds().Size().Y != gImage2.Bounds().Size().Y {
			fmt.Println("WARNING: image dimensions differ!")
		} else {
			if len(gOpts.Verbose) > 0 {
				fmt.Printf("image dimensions: %dx%d\n", gImage1.Bounds().Size().X, gImage1.Bounds().Size().Y)
			}
		}

		var stats utils.DiffStats
		if stats, err = utils.Compare16(gImage1, gImage2); err != nil {
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
