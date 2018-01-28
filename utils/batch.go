package utils

import (
	"encoding/json"
	"os"
)

// DiffBatch ...
type DiffBatch struct {
	Diffs []DiffImages `json:"diffs"`
}

// DiffImages ...
type DiffImages struct {
	Image1 string `json:"img1"`
	Image2 string `json:"img2"`
}

// RunDiffBatch ...
func RunDiffBatch(batchpath string) (exitcode int, err error) {
	var bachFile *os.File
	if bachFile, err = os.Open(batchpath); err != nil {
		return -1, err
	}
	defer bachFile.Close()

	var batch DiffBatch
	json.NewDecoder(bachFile).Decode(&batch)

	return exitcode, nil
}
