package itest

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

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
	greenFG = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	redFG   = color.New(color.FgHiRed, color.Bold).SprintFunc()
	whiteFG = color.New(color.Bold).SprintFunc()
	faintFG = color.New(color.Faint).SprintFunc()

	blueBG = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()

	underline = color.New(color.Underline).SprintFunc()
)

type ColumnWriter struct {
	columns   int
	format    string
	tabWriter *tabwriter.Writer
}

func NewColumnWriter(output io.Writer, columns int, padding int) *ColumnWriter {
	format := ""
	for i := 0; i < columns; i++ {
		format += "%s\t"
	}
	format += "\n"

	return &ColumnWriter{
		columns:   columns,
		format:    format,
		tabWriter: tabwriter.NewWriter(output, 0, 0, padding, ' ', 0),
	}
}

func (w *ColumnWriter) PrintColumns(columns ...interface{}) {
	if len(columns) > w.columns {
		panic(fmt.Sprintf("PrintColumns() called with %d columns, expected at most %d", len(columns), w.columns))
	}

	fmt.Fprintf(w.tabWriter, w.format, columns...)
}

func (w *ColumnWriter) Flush() {
	w.tabWriter.Flush()
}

func debug(format string, a ...interface{}) {
	if Verbose {
		color.Cyan(format, a...)
	}
}
