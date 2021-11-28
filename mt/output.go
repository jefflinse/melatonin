package mt

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

var (
	greenFG = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	redFG   = color.New(color.FgHiRed, color.Bold).SprintFunc()
	whiteFG = color.New(color.Bold).SprintFunc()
	faintFG = color.New(color.Faint).SprintFunc()
	blueBG  = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()
)

type columnWriter struct {
	columns   int
	format    string
	dest      io.Writer
	tabWriter *tabwriter.Writer
}

func newColumnWriter(output io.Writer, columns int, padding int) *columnWriter {
	return &columnWriter{
		columns:   columns,
		format:    strings.Repeat("%s\t", columns) + "\n",
		dest:      output,
		tabWriter: tabwriter.NewWriter(output, 0, 0, padding, ' ', 0),
	}
}

func (w *columnWriter) printLine(str string, args ...interface{}) {
	fmt.Fprintf(w.dest, str+"\n", args...)
}

func (w *columnWriter) printColumns(columns ...interface{}) {
	if len(columns) > w.columns {
		panic(fmt.Sprintf("PrintColumns() called with %d columns, expected at most %d", len(columns), w.columns))
	}

	fmt.Fprintf(w.tabWriter, w.format, columns...)
}

func debug(format string, a ...interface{}) {
	if cfg.Debug {
		color.Cyan(format, a...)
	}
}

func PrintRunResult(result RunResult) {
	w := newColumnWriter(cfg.Stdout, 5, 2)
	for i := range result.TestResults {
		if len(result.TestResults[i].Errors()) > 0 {
			w.printTestFailure(result.TestResults[i], result.TestDurations[i])
		} else {
			w.printTestSuccess(result.TestResults[i], result.TestDurations[i])
		}
	}

	w.printLine("")
	w.printLine("%d passed, %d failed, %d skipped %s", result.Passed, result.Failed, result.Skipped,
		faintFG(fmt.Sprintf("in %s", result.Duration.String())))
	w.tabWriter.Flush()
}

func (w *columnWriter) printTestFailure(result TestResult, duration time.Duration) {
	w.printColumns(
		redFG(" ✘"),
		whiteFG(result.TestCase().Description()),
		blueBG(fmt.Sprintf("%7s ", result.TestCase().Action())),
		result.TestCase().Target(),
		faintFG(duration.String()))

	for _, err := range result.Errors() {
		w.printColumns(redFG(""), redFG(fmt.Sprintf("  %s", err)), blueBG(""), "", faintFG(""))
	}
}

func (w *columnWriter) printTestSuccess(result TestResult, duration time.Duration) {
	w.printColumns(
		greenFG(" ✔"),
		whiteFG(result.TestCase().Description()),
		blueBG(fmt.Sprintf("%7s ", result.TestCase().Action())),
		result.TestCase().Target(),
		faintFG(duration.String()))
}
