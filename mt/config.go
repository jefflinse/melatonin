package mt

import (
	"io"
	"os"
)

const (
	outputTypeNone = iota
	outputTypeFormattedTable
	outputTypeJSON
)

var cfg = struct {
	ContinueOnFailure bool
	OutputType        int
	Stdout            io.Writer
	WorkingDir        string
}{
	ContinueOnFailure: false,
	OutputType:        outputTypeFormattedTable,
	Stdout:            os.Stdout,
	WorkingDir:        "",
}

func init() {
	if os.Getenv("MELATONIN_CONTINUE_ON_FAILURE") != "" {
		cfg.ContinueOnFailure = true
	}

	cfg.Stdout = os.Stdout
	switch os.Getenv("MELATONIN_OUTPUT") {
	case "none":
		cfg.OutputType = outputTypeNone
		cfg.Stdout = io.Discard
	case "json":
		cfg.OutputType = outputTypeJSON
	default:
		cfg.OutputType = outputTypeFormattedTable
	}

	if workdir := os.Getenv("MELATONIN_WORKDIR"); workdir != "" {
		cfg.WorkingDir = workdir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		cfg.WorkingDir = dir
	}
}
