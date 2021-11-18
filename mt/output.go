package mt

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var Verbose bool

func init() {
	if os.Getenv("MELATONIN_VERBOSE") != "" {
		Verbose = true
	}

	if os.Getenv("MELATONIN_NOCOLOR") != "" {
		color.NoColor = true
	}
}

var (
	greenFG = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	redFG   = color.New(color.FgHiRed, color.Bold).SprintFunc()
	whiteFG = color.New(color.Bold).SprintFunc()
	faintFG = color.New(color.Faint).SprintFunc()

	blueBG = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()
)

type columnWriter struct {
	columns   int
	format    string
	stdout    io.Writer
	tabWriter *tabwriter.Writer
}

func newColumnWriter(output io.Writer, columns int, padding int) *columnWriter {
	return &columnWriter{
		columns:   columns,
		format:    strings.Repeat("%s\t", columns) + "\n",
		stdout:    output,
		tabWriter: tabwriter.NewWriter(output, 0, 0, padding, ' ', 0),
	}
}

func (w *columnWriter) PrintColumns(columns ...interface{}) {
	if len(columns) > w.columns {
		panic(fmt.Sprintf("PrintColumns() called with %d columns, expected at most %d", len(columns), w.columns))
	}

	fmt.Fprintf(w.tabWriter, w.format, columns...)
}

func (w *columnWriter) Flush() {
	w.tabWriter.Flush()
	if w.stdout != io.Discard {
		fmt.Println()
	}
}

func debug(format string, a ...interface{}) {
	if Verbose {
		color.Cyan(format, a...)
	}
}
