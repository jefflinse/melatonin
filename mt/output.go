package mt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	cyanFG  = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	greenFG = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	redFG   = color.New(color.FgHiRed, color.Bold).SprintFunc()
	whiteFG = color.New(color.Bold).SprintFunc()
	faintFG = color.New(color.Faint).SprintFunc()
	blueBG  = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()
)

// PrintResults prints the results of a group run to stdout.
//
// By default, the output is formatted as a table and colors are used if possible.
// The behavior can be controlled by setting the MELATONIN_OUTPUT environment
// variable to "json" to produce JSON output, or "none" to disable output all together.
func PrintResults(results GroupRunResult) {
	FPrintResults(cfg.Stdout, results)
}

// FPrintResults prints the results of a group run to the given io.Writer.
//
// By default, the output is formatted as a table and colors are used if possible.
// The behavior can be controlled by setting the MELATONIN_OUTPUT environment
// variable to "json" to produce JSON output, or "none" to disable output all together.
func FPrintResults(w io.Writer, results GroupRunResult) {
	switch cfg.OutputType {
	case outputTypeNone:
		return
	case outputTypeJSON:
		FPrintJSONResults(w, results, false)
	default:
		FPrintFormattedResults(w, results)
	}
}

// PrintFormattedResults prints the results of a group run as a formatted table to stdout.
func PrintFormattedResults(results GroupRunResult) {
	FPrintFormattedResults(cfg.Stdout, results)
}

// FPrintFormattedResults prints the results of a group run as a formatted table to the given io.Writer.
func FPrintFormattedResults(w io.Writer, groupResult GroupRunResult) {
	cw := newColumnWriter(w, 5, 1)

	if groupResult.Group.Name != "" {
		cw.printLine(cyanFG(groupResult.Group.Name))
	}

	for i := range groupResult.Results {
		if len(groupResult.Results[i].TestResult.Failures()) > 0 {
			cw.printTestFailure(i+1, groupResult.Results[i])
		} else {
			cw.printTestSuccess(i+1, groupResult.Results[i])
		}
	}

	cw.printLine("")
	cw.printLine(fmt.Sprintf("%d passed, %d failed, %d skipped %s", groupResult.Passed, groupResult.Failed, groupResult.Skipped,
		faintFG(fmt.Sprintf("in %s", groupResult.Duration.String()))))
	cw.Flush()
}

type jsonOutputObj struct {
	Groups []jsonGroupRunResult `json:"groups"`
}

type jsonGroupRunResult struct {
	Name     string              `json:"name"`
	Duration time.Duration       `json:"duration"`
	Results  []jsonTestRunResult `json:"results"`
}

type jsonTestRunResult struct {
	Test      jsonTest      `json:"test"`
	Result    jsonResult    `json:"result"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at"`
	Duration  time.Duration `json:"duration"`
}

type jsonTest struct {
	Description string   `json:"description"`
	Action      string   `json:"action"`
	Target      string   `json:"target"`
	Data        TestCase `json:"data,omitempty"`
}

type jsonResult struct {
	Failures []error    `json:"failures"`
	Data     TestResult `json:"data,omitempty"`
}

// PrintJSONResults prints the results of a group run as JSON to the given io.Writer.
func PrintJSONResults(results GroupRunResult, deep bool) error {
	return FPrintJSONResults(cfg.Stdout, results, deep)
}

// FPrintJSONResults prints the results of a group run as JSON to the given io.Writer.
func FPrintJSONResults(w io.Writer, result GroupRunResult, deep bool) error {
	groupResultObj := jsonGroupRunResult{
		Name:     result.Group.Name,
		Duration: result.Duration,
		Results:  make([]jsonTestRunResult, len(result.Results)),
	}

	for i := range result.Results {
		testRunResult := jsonTestRunResult{
			Test: jsonTest{
				Description: result.Results[i].TestCase.Description(),
				Action:      result.Results[i].TestCase.Action(),
				Target:      result.Results[i].TestCase.Target(),
			},
			Result: jsonResult{
				Failures: result.Results[i].TestResult.Failures(),
			},
			StartedAt: result.Results[i].StartedAt,
			EndedAt:   result.Results[i].EndedAt,
			Duration:  result.Results[i].Duration,
		}

		if deep {
			testRunResult.Test.Data = result.Results[i].TestCase
			testRunResult.Result.Data = result.Results[i].TestResult
		}

		groupResultObj.Results[i] = testRunResult
	}

	return json.NewEncoder(w).Encode(jsonOutputObj{
		Groups: []jsonGroupRunResult{groupResultObj},
	})
}

type columnWriter struct {
	columns            int
	elasticColumnIndex int
	format             string
	dest               io.Writer
	buf                *strings.Builder
	tabWriter          *tabwriter.Writer
	currentLineNum     int
	nonTableLines      map[int][]string
	termWidth          int
}

type decoratorFunc func(...interface{}) string

func newColumnWriter(output io.Writer, columns int, elasticColumnIndex int) *columnWriter {
	buf := &strings.Builder{}
	return &columnWriter{
		columns:            columns,
		elasticColumnIndex: elasticColumnIndex,
		format:             strings.Repeat("%s\t", columns) + "\n",
		buf:                buf,
		dest:               output,
		tabWriter:          tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0),
		currentLineNum:     -1,
		nonTableLines:      map[int][]string{},
		termWidth:          getTerminalWidth() - 2,
	}
}

func (w *columnWriter) Flush() {
	w.tabWriter.Flush()
	if lines, ok := w.nonTableLines[-1]; ok {
		fmt.Fprintln(w.dest, strings.Join(lines, "\n"))
	}

	s := bufio.NewScanner(strings.NewReader(w.buf.String()))
	i := 0
	for ; s.Scan(); i++ {
		fmt.Fprintln(w.dest, s.Text())
		if lines, ok := w.nonTableLines[i]; ok {
			fmt.Fprintln(w.dest, strings.Join(lines, "\n"))
		}
	}

	if lines, ok := w.nonTableLines[i]; ok {
		fmt.Fprintln(w.dest, strings.Join(lines, "\n"))
	}
}

func (w *columnWriter) printColumns(decorators map[int]decoratorFunc, columns ...interface{}) {
	if len(columns) > w.columns {
		panic(fmt.Sprintf("PrintColumns() called with %d columns, expected at most %d", len(columns), w.columns))
	}

	noColor := color.NoColor
	color.NoColor = true
	buf := &strings.Builder{}
	temp := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	n, _ := fmt.Fprintf(temp, w.format, columns...)
	color.NoColor = noColor

	if diff := n - w.termWidth; w.termWidth > 0 && diff > 0 {
		str := columns[w.elasticColumnIndex].(string)
		newLength := len(str) - diff - 5
		if newLength > 0 {
			str = str[:newLength]
		} else {
			str = ""
		}

		columns[w.elasticColumnIndex] = str + "..."
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

func (w *columnWriter) printTestSuccess(testNum int, result TestRunResult) {
	w.printColumns(map[int]decoratorFunc{0: greenFG, 1: whiteFG, 2: blueBG, 4: faintFG},
		fmt.Sprintf("✔ %d", testNum),
		result.TestCase.Description(),
		fmt.Sprintf("%7s ", result.TestCase.Action()),
		result.TestCase.Target(),
		result.Duration.String())
}

func (w *columnWriter) printTestFailure(testNum int, result TestRunResult) {
	decorators := map[int]decoratorFunc{0: redFG, 1: whiteFG, 2: blueBG, 4: faintFG}
	w.printColumns(decorators,
		fmt.Sprintf("✘ %d", testNum),
		result.TestCase.Description(),
		fmt.Sprintf("%7s ", result.TestCase.Action()),
		result.TestCase.Target(),
		result.Duration.String())

	failures := result.TestResult.Failures()
	for i := 0; i < len(failures)-1; i++ {
		w.printLine(redFG(fmt.Sprintf("├╴ %s", failures[i])))
	}

	w.printLine(redFG(fmt.Sprintf("└╴ %s", failures[len(failures)-1])))
	w.printLine("")
}

func getTerminalWidth() int {
	if !term.IsTerminal(0) {
		return 0
	}

	width, _, err := term.GetSize(0)
	if err != nil {
		return 0
	}

	return width
}
