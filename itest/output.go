package itest

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var Verbose bool

var (
	underlineText = color.New(color.Underline).SprintFunc()
)

func fatal(format string, a ...interface{}) {
	problem(format, a...)
	os.Exit(1)
}

func problem(format string, a ...interface{}) {
	color.Red(format, a...)
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
		color.Blue(format, a...)
	}
}
