package itest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

type Golden struct {
	WantStatus  int
	WantHeaders http.Header
	WantBody    interface{}
}

const (
	headersLinePrefix = "--- headers"
	bodyLinePrefix    = "--- body"
)

func NewGoldenFromFile(fs afero.Fs, path string) (*Golden, error) {
	if exists, err := afero.Exists(fs, path); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("golden file %q: not found", path)
	}

	f, err := fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	golden := &Golden{}
	var headersLines, bodyLines []string
	var target *[]string
	bodyIsJSON := false

	scanner := bufio.NewScanner(f)
	matcher := search.New(language.English, search.IgnoreCase)
	for scanner.Scan() {
		line := scanner.Text()

		// skip any empty lines that aren't part of the headers or body content
		if target == nil && len(line) == 0 {
			continue
		}

		// status must be the first non-empty line encountered
		if golden.WantStatus == 0 {
			if err := golden.parseStatusLine(line); err != nil {
				return nil, err
			}
			continue
		}

		if start, _ := matcher.IndexString(line, headersLinePrefix); start != -1 {
			if err := golden.parseHeaderDirectives(line[2:]); err != nil {
				return nil, err
			}

			target = &headersLines
			continue
		} else if start, _ := matcher.IndexString(line, bodyLinePrefix); start != -1 {
			if bodyIsJSON, err = golden.parseBodyDirectives(line[2:]); err != nil {
				return nil, err
			}

			target = &bodyLines
			continue
		} else {
			if target == nil {
				return nil, newGoldenFileError(path, fmt.Errorf("unexpected line %q", line))
			}

			*target = append(*target, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if err := golden.parseHeaderLines(headersLines); err != nil {
		return nil, err
	}

	if err := golden.parseBodyLines(bodyLines, bodyIsJSON); err != nil {
		return nil, err
	}

	return golden, nil
}

func (g *Golden) parseStatusLine(line string) error {
	var err error
	g.WantStatus, err = strconv.Atoi(line)
	if err != nil {
		return fmt.Errorf("invalid status %q", line)
	}

	return nil
}

func (g *Golden) parseHeaderDirectives(line string) error {
	headersDirectives := strings.Split(line, " ")
	for _, directive := range headersDirectives[2:] {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		switch directive {
		default:
			return fmt.Errorf("unknown headers directive %q", directive)
		}
	}

	return nil
}

func (g *Golden) parseHeaderLines(lines []string) error {
	if len(lines) > 0 {
		g.WantHeaders = http.Header{}
		for _, line := range lines {
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid header %q", line)
			}

			key := strings.TrimSpace(parts[0])
			if key == "" {
				return fmt.Errorf("invalid header key %q", line)
			}

			value := strings.TrimSpace(parts[1])

			g.WantHeaders.Add(key, value)
		}
	}

	return nil
}

func (g *Golden) parseBodyDirectives(line string) (bool, error) {
	bodyDirectives := strings.Split(line, " ")
	for _, directive := range bodyDirectives[2:] {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		switch directive {
		case "json":
			return true, nil
		default:
			return false, fmt.Errorf("unknown body directive %q", directive)
		}
	}

	return false, nil
}

func (g *Golden) parseBodyLines(lines []string, asJSON bool) error {
	if len(lines) > 0 {
		body := strings.Join(lines, "\n")
		if asJSON {
			if err := json.Unmarshal([]byte(body), &g.WantBody); err != nil {
				return fmt.Errorf("invalid JSON body: %s\n---\n%s\n---", err, body)
			}
		} else {
			g.WantBody = body
		}
	}

	return nil
}

func newGoldenFileError(path string, err error) error {
	return fmt.Errorf("golden file %q: %w", path, err)
}
