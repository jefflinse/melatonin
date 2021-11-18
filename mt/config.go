package mt

import (
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

const (
	OutputModeNone = iota
	OutputModeSimple
	OutputModePretty
)

var cfg = struct {
	ContinueOnFailure bool
	Debug             bool
	OutputMode        int
	Stdout            io.Writer
	WorkingDir        string
}{
	ContinueOnFailure: false,
	Debug:             false,
}

func init() {
	if os.Getenv("MELATONIN_CONTINUE_ON_FAILURE") != "" {
		cfg.ContinueOnFailure = true
	}

	cfg.Stdout = os.Stdout
	switch os.Getenv("MELATONIN_OUTPUT") {
	case "none":
		cfg.OutputMode = OutputModeNone
		cfg.Stdout = io.Discard
	case "simple":
		cfg.OutputMode = OutputModeSimple
		color.NoColor = true
	default:
		cfg.OutputMode = OutputModePretty
	}

	if os.Getenv("MELATONIN_DEBUG") != "" {
		cfg.Debug = true
	}

	if workdir := os.Getenv("MELATONIN_WORKDIR"); workdir != "" {
		cfg.WorkingDir = workdir
	} else {
		dir, err := os.Executable()
		if err != nil {
			panic(err)
		}
		cfg.WorkingDir = filepath.Dir(dir)
	}
}
