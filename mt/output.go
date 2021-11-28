package mt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

var (
	blueFG  = color.New(color.FgHiBlue, color.Bold).SprintFunc()
	greenFG = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	redFG   = color.New(color.FgHiRed, color.Bold).SprintFunc()
	whiteFG = color.New(color.Bold).SprintFunc()
	faintFG = color.New(color.Faint).SprintFunc()
	blueBG  = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()
)

// PrintResult prints a RunResult to stdout.
//
// By default, the output is formatted as a table and colors are used if possible.
// The behavior can be controlled by setting the MELATONIN_OUTPUT environment
// variable to "simple" to disable colors, or "none" to disable output all together.
func PrintResult(result RunResult) {
	FPrintResult(cfg.Stdout, result)
}

// FPrintResult prints a RunResult to the given io.Writer.
//
// By default, if the supplied writer is stdout or a TTY, the output is formatted as
// a table and colors are used if possible. The behavior can be controlled by setting
// the MELATONIN_OUTPUT environment variable to "simple" to disable colors, or "none"
// to disable output all together.
func FPrintResult(w io.Writer, result RunResult) {
	cw := newColumnWriter(w, 5, 2)

	if result.Group.Name != "" {
		cw.printLine(blueFG(result.Group.Name))
	}

	for i := range result.TestResults {
		if len(result.TestResults[i].Failures()) > 0 {
			cw.printTestFailure(i+1, result.TestResults[i], result.TestDurations[i])
		} else {
			cw.printTestSuccess(i+1, result.TestResults[i], result.TestDurations[i])
		}
	}

	cw.printLine("")
	cw.printLine("%d passed, %d failed, %d skipped %s", result.Passed, result.Failed, result.Skipped,
		faintFG(fmt.Sprintf("in %s", result.Duration.String())))
	cw.Flush()
}

type columnWriter struct {
	columns        int
	format         string
	dest           io.Writer
	buf            *strings.Builder
	tabWriter      *tabwriter.Writer
	currentLineNum int
	nonTableLines  map[int][]string
}

type decoratorFunc func(...interface{}) string

func newColumnWriter(output io.Writer, columns int, padding int) *columnWriter {
	buf := &strings.Builder{}
	return &columnWriter{
		columns:       columns,
		format:        strings.Repeat("%s\t", columns) + "\n",
		buf:           buf,
		dest:          output,
		tabWriter:     tabwriter.NewWriter(buf, 0, 0, padding, ' ', 0),
		nonTableLines: map[int][]string{},
	}
}

func (w *columnWriter) Flush() {
	w.tabWriter.Flush()
	s := bufio.NewScanner(strings.NewReader(w.buf.String()))
	for i := 0; s.Scan(); i++ {
		if lines, ok := w.nonTableLines[i]; ok {
			fmt.Fprintln(w.dest, strings.Join(lines, "\n"))
		}
		fmt.Fprintln(w.dest, s.Text())
	}
}

func (w *columnWriter) printColumns(decorators map[int]decoratorFunc, columns ...interface{}) {
	if len(columns) > w.columns {
		panic(fmt.Sprintf("PrintColumns() called with %d columns, expected at most %d", len(columns), w.columns))
	}

	if !color.NoColor {
		for i := 0; i < len(decorators); i++ {
			fn, ok := decorators[i]
			if ok {
				columns[i] = fn(columns[i])
			}
		}
	}

	fmt.Fprintf(w.tabWriter, w.format, columns...)
	w.currentLineNum++
}

func (w *columnWriter) printLine(str string, args ...interface{}) {
	w.nonTableLines[w.currentLineNum] = append(w.nonTableLines[w.currentLineNum], str)
}

func (w *columnWriter) printTestSuccess(testNum int, result TestResult, duration time.Duration) {
	w.printColumns(map[int]decoratorFunc{0: greenFG, 1: whiteFG, 2: blueBG, 4: faintFG},
		fmt.Sprintf("✔ %d", testNum),
		result.TestCase().Description(),
		fmt.Sprintf("%7s ", result.TestCase().Action()),
		result.TestCase().Target(),
		duration.String())
}

func (w *columnWriter) printTestFailure(testNum int, result TestResult, duration time.Duration) {
	decorators := map[int]decoratorFunc{0: redFG, 1: whiteFG, 2: blueBG, 4: faintFG}
	w.printColumns(decorators,
		fmt.Sprintf("✘ %d", testNum),
		result.TestCase().Description(),
		fmt.Sprintf("%7s ", result.TestCase().Action()),
		result.TestCase().Target(),
		duration.String())

	for _, err := range result.Failures() {
		w.printLine(fmt.Sprintf("%s", redFG(err)))
	}

	w.printLine("")
}
