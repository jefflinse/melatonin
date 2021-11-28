package mt

import (
	"io"
	"os"

	"github.com/fatih/color"
)

var cfg = struct {
	ContinueOnFailure bool
	Stdout            io.Writer
	WorkingDir        string
}{}

func init() {
	if os.Getenv("MELATONIN_CONTINUE_ON_FAILURE") != "" {
		cfg.ContinueOnFailure = true
	}

	cfg.Stdout = os.Stdout
	switch os.Getenv("MELATONIN_OUTPUT") {
	case "none":
		cfg.Stdout = io.Discard
	case "simple":
		color.NoColor = true
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
