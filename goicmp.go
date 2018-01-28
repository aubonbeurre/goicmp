package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/aubonbeurre/goicmp/utils"
	gostats "github.com/aubonbeurre/gostats/utils"
	"github.com/bndr/gojenkins"
	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
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
	Jenkins bool   `short:"j" long:"jenkins" description:"Jenkins mode"`
}

func goJenkins() (err error) {
	gostats.GPrefs.Load()
	var jenkins = gostats.NewJenkinsClient()
	var jobs []gojenkins.InnerJob
	if jobs, err = jenkins.GetAllJobNames(); err != nil {
		return err
	}
	var data = make([][]string, 0)
	var idx = 1
	for _, job := range jobs {
		var row = make([]string, 0)
		if job.Color == "notbuilt" || job.Color == "disabled" {
			continue
		}
		row = append(row, fmt.Sprintf("%d", idx))
		row = append(row, job.Name)
		row = append(row, job.Color)
		row = append(row, job.Url)
		data = append(data, row)
		idx++
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Name", "Color", "Url"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
	return nil
}

func main() {
	var parser = flags.NewParser(&gOpts, flags.Default)

	var err error
	var args []string
	if args, err = parser.Parse(); err != nil {
		os.Exit(1)
	}

	if gOpts.Jenkins {
		if err = goJenkins(); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

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
