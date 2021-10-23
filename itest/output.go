package itest

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var Verbose bool

func init() {
	if os.Getenv("ITEST_VERBOSE") != "" {
		Verbose = true
	}

	if os.Getenv("ITEST_NOCOLOR") != "" {
		color.NoColor = true
	}
}

var (
	greenText = color.New(color.FgGreen).SprintFunc()
	redText   = color.New(color.FgHiRed).SprintFunc()
)

func fatal(format string, a ...interface{}) {
	problem(format, a...)
	os.Exit(1)
}

func problem(format string, a ...interface{}) {
	color.HiRed(format, a...)
}

func warn(format string, a ...interface{}) {
	color.Yellow(format, a...)
}

func info(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Println()
}

func debug(format string, a ...interface{}) {
	if Verbose {
		color.Cyan(format, a...)
	}
}
