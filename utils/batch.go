package utils

import (
	"encoding/json"
	"image"
	"log"
	"os"
	"runtime"
	"time"
)

// DiffBatch ...
type DiffBatch struct {
	Diffs []DiffImage `json:"diffs"`
}

// DiffImage ...
type DiffImage struct {
	Image     string   `json:"img"`
	Baselines []string `json:"baselines"`
	AEP       string   `json:"aep"`
	CompInfo  string   `json:"compinfo"`
	TC        string   `json:"tc"`
	UUID      string   `json:"uuid"`
}

// DiffResults ...
type DiffResults struct {
	Results []DiffResult `json:"results"`
	Elapsed string       `json:"elapsed"`
}

// DiffResult ...
type DiffResult struct {
	TC    string    `json:"tc"`
	UUID  string    `json:"uuid"`
	Stats DiffStats `json:"stats"`
	Err   string    `json:"err"`
}

// SetError ...
func (b *DiffResult) SetError(err error, img string) {
	log.Printf("Error diffing image '%s' (%v)\n", img, err)
	b.Err = err.Error()
}

func workerDiffImages(id int, jobs <-chan DiffImage, results chan<- DiffResult) {
	for j := range jobs {
		var result = DiffResult{TC: j.TC, UUID: j.UUID}
		var err error

		var rgba *image.RGBA64
		//log.Printf("Now diffing %s (%d)", j.Image, id)
		if rgba, err = DownloadOrLoadImage(j.Image); err != nil {
			result.SetError(err, j.Image)
			results <- result
			continue
		}

		for _, baseline := range j.Baselines {
			var rgba2 *image.RGBA64
			if rgba2, err = DownloadOrLoadImage(baseline); err != nil {
				result.SetError(err, baseline)
				break
			}

			if result.Stats, err = Compare16(rgba, rgba2); err != nil {
				result.SetError(err, j.Image)
				break
			}

			if result.Stats.ExactSame {
				break
			}
		}
		if len(result.Err) > 0 {
			results <- result
			continue
		}
		if !result.Stats.ExactSame {
			//log.Printf("Found diff for %s %s", j.CompInfo, j.AEP)
		}
		results <- result
	}
}

// RunDiffBatch ...
func RunDiffBatch(batchpath string, output string) (exitcode int, err error) {
	var bachFile *os.File
	if bachFile, err = os.Open(batchpath); err != nil {
		return -1, err
	}
	defer bachFile.Close()

	var batch DiffBatch
	json.NewDecoder(bachFile).Decode(&batch)

	start := time.Now()
	var jobs = make(chan DiffImage, len(batch.Diffs))
	var results = make(chan DiffResult, len(batch.Diffs))

	for w := 1; w <= runtime.NumCPU(); w++ {
		go workerDiffImages(w, jobs, results)
	}

	for _, job := range batch.Diffs {
		jobs <- job
	}
	close(jobs)

	var res DiffResults
	res.Results = make([]DiffResult, 0)
	for range batch.Diffs {
		aresult := <-results
		if !aresult.Stats.ExactSame || len(aresult.Err) > 0 {
			res.Results = append(res.Results, aresult)
		}
	}

	res.Elapsed = time.Since(start).Round(time.Duration(time.Millisecond)).String()
	log.Printf("Diff %d images (in %s)\n", len(batch.Diffs), res.Elapsed)

	if len(output) > 0 {
		fd, _ := os.Create(output)
		defer fd.Close()
		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "\t")
		encoder.Encode(res)
	} else {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "\t")
		encoder.Encode(res)
	}

	return exitcode, nil
}
