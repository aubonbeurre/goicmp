package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

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
	Batch   string `short:"b" long:"batch" description:"A json file with a list of diff" value-name:"BATCHFILE"`
	Out     string `short:"o" long:"output" description:"A json output with the results" value-name:"OUTPUT"`
}

func main() {
	var parser = flags.NewParser(&gOpts, flags.Default)

	var err error
	var args []string
	if args, err = parser.Parse(); err != nil {
		os.Exit(1)
	}

	gOpts.Batch = "/Users/aparente/test.json"
	if len(gOpts.Batch) > 0 {
		var exitcode int
		if exitcode, err = utils.RunDiffBatch(gOpts.Batch, gOpts.Out); err != nil {
			panic(err)
		}
		os.Exit(exitcode)
	}

	if len(args) < 1 || len(args) > 2 {
		//panic(fmt.Errorf("Too many or not enough arguments"))
	}

	gDiffFlag = len(args) == 2

	var image1path = args[0]
	if gImage1, err = utils.DownloadOrLoadImage(image1path); err != nil {
		panic(fmt.Errorf("Error reading '%s' (%v)", image1path, err))
	}

	if gDiffFlag {
		var image2path = args[1]

		if gImage2, err = utils.DownloadOrLoadImage(image2path); err != nil {
			panic(fmt.Errorf("Error reading '%s' (%v)", image2path, err))
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
