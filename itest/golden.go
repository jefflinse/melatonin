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

	bodyIsJSON bool
}

func NewParseGoldenFileFs(fs afero.Fs, path string) (*Golden, error) {
	if exists, err := afero.Exists(fs, path); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("golden file %q does not exist", path)
	}

	f, err := fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	golden := &Golden{}
	statusLine := ""
	var headersLines, bodyLines []string
	var target *[]string
	scanner := bufio.NewScanner(f)
	matcher := search.New(language.English, search.IgnoreCase)
	for scanner.Scan() {
		line := scanner.Text()

		// skip any empty lines that aren't part of the headers or body content
		if target == nil && len(line) == 0 {
			continue
		}

		if start, _ := matcher.IndexString(line, "--- headers"); start != -1 {
			if statusLine == "" {
				return nil, fmt.Errorf("no status found before encountering headers")
			}

			headersDirectives := strings.Split(line, " ")
			for _, directive := range headersDirectives[2:] {
				directive = strings.TrimSpace(directive)
				if directive == "" {
					continue
				}

				switch directive {
				default:
					return nil, fmt.Errorf("unknown headers directive %q", directive)
				}
			}

			target = &headersLines
			continue
		} else if start, _ := matcher.IndexString(line, "--- body"); start != -1 {
			if statusLine == "" {
				return nil, fmt.Errorf("no status found before encountering body %q", statusLine)
			}

			bodyDirectives := strings.Split(line, " ")
			for _, directive := range bodyDirectives[2:] {
				directive = strings.TrimSpace(directive)
				if directive == "" {
					continue
				}

				switch directive {
				case "json":
					golden.bodyIsJSON = true
				default:
					return nil, fmt.Errorf("unknown body directive %q", directive)
				}
			}

			target = &bodyLines
			continue
		} else if statusLine == "" {
			statusLine = line
			continue
		} else {
			*target = append(*target, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	golden.WantStatus, err = strconv.Atoi(statusLine)
	if err != nil {
		return nil, fmt.Errorf("invalid status %q: %s", golden.WantStatus, err)
	}

	if len(headersLines) > 0 {
		golden.WantHeaders = http.Header{}
		for _, line := range headersLines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid header %q", line)
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			golden.WantHeaders.Add(key, value)
		}
	}

	if len(bodyLines) > 0 {
		body := strings.Join(bodyLines, "\n")
		if golden.bodyIsJSON {
			if err := json.Unmarshal([]byte(body), &golden.WantBody); err != nil {
				return nil, fmt.Errorf("invalid JSON body %q: %s", body, err)
			}
		} else {
			golden.WantBody = body
		}
	}

	return golden, nil
}
