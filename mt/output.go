package mt

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jefflinse/tablecloth"
)

var (
	cyanFG              = color.New(color.FgHiCyan).SprintFunc()
	cyanFGBold          = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	cyanFGUnderline     = color.New(color.FgHiCyan, color.Underline).SprintFunc()
	cyanFGBoldUnderline = color.New(color.FgHiCyan, color.Bold, color.Underline).SprintFunc()
	greenFG             = color.New(color.FgHiGreen).SprintFunc()
	greenFGBold         = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	redFG               = color.New(color.FgHiRed).SprintFunc()
	redFGBold           = color.New(color.FgHiRed, color.Bold).SprintFunc()
	whiteFG             = color.New(color.FgWhite).SprintFunc()
	whiteFGBold         = color.New(color.FgWhite, color.Bold).SprintFunc()
	faintFG             = color.New(color.Faint).SprintFunc()
	blueBG              = color.New(color.BgBlue, color.FgHiWhite).SprintFunc()
)

var (
	groupHeaderColor = cyanFG
)

const (
	indentationPrefix = "│ "
	// indentationPrefix = "  "
	groupFooterPrefix = "└╴ "
	// groupFooterPrefix = "  "
)

// PrintResults prints the results of a group run to stdout.
//
// By default, the output is formatted as a table and colors are used if possible.
// The behavior can be controlled by setting the MELATONIN_OUTPUT environment
// variable to "json" to produce JSON output, or "none" to disable output all together.
func PrintResults(results *GroupRunResult) {
	FPrintResults(cfg.Stdout, results)
}

// FPrintResults prints the results of a group run to the given io.Writer.
//
// By default, the output is formatted as a table and colors are used if possible.
// The behavior can be controlled by setting the MELATONIN_OUTPUT environment
// variable to "json" to produce JSON output, or "none" to disable output all together.
func FPrintResults(w io.Writer, results *GroupRunResult) {
	switch cfg.OutputType {
	case outputTypeJSON:
		fprintJSONResults(w, results, false)
	default:
		table := tablecloth.NewTable(4)
		fprintFormattedResults(table, results, 0)
	}
}

// printFormattedResults prints the results of a group run as a formatted table to stdout.
func printFormattedResults(results *GroupRunResult) {
	table := tablecloth.NewTable(4)
	fprintFormattedResults(table, results, 0)
}

// fprintFormattedResults prints the results of a group run as a formatted table to the given io.Writer.
func fprintFormattedResults(table *tablecloth.Table, groupResult *GroupRunResult, depth int) {
	printGroupHeader(table, groupResult.Group.Name, depth)

	for i := range groupResult.TestResults {
		if len(groupResult.TestResults[i].TestResult.Failures()) > 0 {
			printTestFailure(table, i+1, groupResult.TestResults[i], depth)
		} else {
			printTestSuccess(table, i+1, groupResult.TestResults[i], depth)
		}
	}

	// print a newline between last test result and first group result
	if len(groupResult.TestResults) > 0 {
		printLine(table, depth+1, "")
	}
	for i := range groupResult.SubgroupResults {
		fprintFormattedResults(table, groupResult.SubgroupResults[i], depth+1)
		// print a newline after each subgroup
		printLine(table, depth+1, "")
	}

	printGroupFooter(table, groupResult.Group.Name, depth, fmt.Sprintf(
		"%d passed, %d failed, %d skipped %s",
		groupResult.Passed,
		groupResult.Failed,
		groupResult.Skipped,
		faintFG(fmt.Sprintf("in %s", groupResult.Duration.String()))))

	if depth == 0 {
		table.Write(os.Stdout)
	}
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
func PrintJSONResults(results *GroupRunResult, deep bool) error {
	return fprintJSONResults(cfg.Stdout, results, deep)
}

// fprintJSONResults prints the results of a group run as JSON to the given io.Writer.
func fprintJSONResults(w io.Writer, result *GroupRunResult, deep bool) error {
	groupResultObj := jsonGroupRunResult{
		Name:     result.Group.Name,
		Duration: result.Duration,
		Results:  make([]jsonTestRunResult, len(result.TestResults)),
	}

	for i := range result.TestResults {
		testRunResult := jsonTestRunResult{
			Test: jsonTest{
				Description: result.TestResults[i].TestCase.Description(),
				Action:      result.TestResults[i].TestCase.Action(),
				Target:      result.TestResults[i].TestCase.Target(),
			},
			Result: jsonResult{
				Failures: result.TestResults[i].TestResult.Failures(),
			},
			StartedAt: result.TestResults[i].StartedAt,
			EndedAt:   result.TestResults[i].EndedAt,
			Duration:  result.TestResults[i].Duration,
		}

		if deep {
			testRunResult.Test.Data = result.TestResults[i].TestCase
			testRunResult.Result.Data = result.TestResults[i].TestResult
		}

		groupResultObj.Results[i] = testRunResult
	}

	return json.NewEncoder(w).Encode(jsonOutputObj{
		Groups: []jsonGroupRunResult{groupResultObj},
	})
}

func printGroupHeader(table *tablecloth.Table, groupName string, depth int) {
	if groupName == "" {
		return
	}
	line := fmt.Sprintf("%s", strings.Repeat(indentationPrefix, depth))
	line = fmt.Sprintf("%s%s", faintFG(line), groupHeaderColor(groupName))
	table.AddLine(line)
}

func printGroupFooter(table *tablecloth.Table, groupName string, depth int, stats string) {
	line := fmt.Sprintf("%s%s", faintFG(groupFooterPrefix), stats)
	printLine(table, depth, line)
}

func printLine(table *tablecloth.Table, depth int, str string, args ...interface{}) {
	line := faintFG(strings.Repeat(indentationPrefix, depth)) + str
	table.AddLine(line)
}

func printTestSuccess(table *tablecloth.Table, testNum int, result TestRunResult, depth int) {

	table.AddRow(
		tablecloth.Cell{
			Format: "%s%s %s %s",
			Values: []tablecloth.FormattableCellValue{
				{Value: strings.Repeat(indentationPrefix, depth+1), Format: faintFG},
				{Value: "✔", Format: greenFG},
				{Value: testNum, Format: greenFG},
				{Value: result.TestCase.Description(), Format: whiteFG},
			},
		},
		tablecloth.Cell{
			Format: "%s",
			Values: []tablecloth.FormattableCellValue{
				{Value: fmt.Sprintf("%7s ", result.TestCase.Action()), Format: blueBG},
			},
		},
		tablecloth.Cell{
			Format: result.TestCase.Target(),
		},
		tablecloth.Cell{
			Format: "%7s ",
			Values: []tablecloth.FormattableCellValue{
				{Value: result.Duration.String(), Format: faintFG},
			},
		},
	)

	// w.printColumns(
	// 	fmt.Sprintf("%s%s %s",
	// 		faintFG(strings.Repeat(indentationPrefix, depth+1)),
	// 		fmt.Sprintf("%s %s", greenFG("✔"), whiteFG(testNum)),
	// 		whiteFG(result.TestCase.Description())),
	// 	blueBG(fmt.Sprintf("%7s ", result.TestCase.Action())),
	// 	result.TestCase.Target(),
	// 	faintFG(result.Duration.String()))
}

func printTestFailure(table *tablecloth.Table, testNum int, result TestRunResult, depth int) {

	table.AddRow(
		tablecloth.Cell{
			Format: "%s%s %s %s",
			Values: []tablecloth.FormattableCellValue{
				{Value: strings.Repeat(indentationPrefix, depth+1), Format: faintFG},
				{Value: "✘", Format: redFGBold},
				{Value: testNum, Format: redFGBold},
				{Value: result.TestCase.Description(), Format: whiteFGBold},
			},
		},
		tablecloth.Cell{
			Format: "%s",
			Values: []tablecloth.FormattableCellValue{
				{Value: fmt.Sprintf("%7s ", result.TestCase.Action()), Format: blueBG},
			},
		},
		tablecloth.Cell{
			Format: result.TestCase.Target(),
		},
		tablecloth.Cell{
			Format: "%s",
			Values: []tablecloth.FormattableCellValue{
				{Value: result.Duration.String(), Format: faintFG},
			},
		},
	)

	// w.printColumns(
	// 	fmt.Sprintf("%s%s %s",
	// 		faintFG(strings.Repeat(indentationPrefix, depth+1)),
	// 		fmt.Sprintf("%s %s", redFGBold("✘"), whiteFG(testNum)),
	// 		whiteFGBold(result.TestCase.Description())),
	// 	blueBG(fmt.Sprintf("%7s ", result.TestCase.Action())),
	// 	result.TestCase.Target(),
	// 	faintFG(result.Duration.String()))

	failures := result.TestResult.Failures()
	for i := 0; i < len(failures)-1; i++ {
		// w.printLine(depth+1, redFG(fmt.Sprintf("├╴  %s", failures[i])))
		printLine(table, depth+1, redFG(fmt.Sprintf("  %s", failures[i])))
	}

	printLine(table, depth+1, redFG(fmt.Sprintf("  %s", failures[len(failures)-1])))
	// w.printLine(depth+1, redFG(fmt.Sprintf("└╴  %s", failures[len(failures)-1])))
}
