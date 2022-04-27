package golden

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

// Golden represents the contents of a golden file.
type Golden struct {
	// WantStatus is the expected response status code.
	WantStatus int

	// WantHeaders are the expected response headers.
	//
	// If MatchHeadersExactly is set to true, any unexpected headers will cause
	// the test utilizing this golden file to fail.
	WantHeaders http.Header

	// WantBody is the expected response body. If the body is a JSON object
	// and MatchBodyJSONExactly is set to true, the JSON is expected to exactly
	// match exactly, and any unexpected JSON keys or values will cause the test
	// utilizing this golden file to fail.
	WantBody any

	// MatchHeadersExactly determines whether or not unexpected headers will cause
	// a test utilizing this golden file to fail.
	MatchHeadersExactly bool

	// MatchBodyJSONExactly determines whether or not unexpected JSON keys or values
	// will cause a test utilizing this golden file to fail.
	MatchBodyJSONExactly bool
}

const (
	headersLinePrefix = "--- headers"
	bodyLinePrefix    = "--- body"
)

// AppFS is the filesystem used by the golden package.
var AppFS = afero.NewOsFs()

// LoadFile loads a golden file from the given path.
func LoadFile(path string) (*Golden, error) {
	if exists, err := afero.Exists(AppFS, path); err != nil {
		return nil, newGoldenFileError(path, err)
	} else if !exists {
		return nil, fmt.Errorf("golden file %q: not found", path)
	}

	f, err := AppFS.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, newGoldenFileError(path, err)
	}
	defer f.Close()

	golden := &Golden{}
	var headersLines, bodyLines []string
	var target *[]string
	var foundHeaders, foundBody, bodyIsJSON bool

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
				return nil, newGoldenFileError(path, err)
			}
			continue
		}

		if start, _ := matcher.IndexString(line, headersLinePrefix); start != -1 {
			if foundHeaders {
				return nil, newGoldenFileError(path, fmt.Errorf("duplicate headers directive"))
			} else if foundBody {
				return nil, newGoldenFileError(path, fmt.Errorf("headers directive must come before body directive"))
			}

			if err := golden.parseHeaderDirectives(line[2:]); err != nil {
				return nil, newGoldenFileError(path, err)
			}

			foundHeaders = true
			target = &headersLines
			continue
		} else if start, _ := matcher.IndexString(line, bodyLinePrefix); start != -1 {
			if foundBody {
				return nil, newGoldenFileError(path, fmt.Errorf("duplicate body directive"))
			}

			if bodyIsJSON, err = golden.parseBodyDirectives(line[2:]); err != nil {
				return nil, newGoldenFileError(path, err)
			}

			foundBody = true
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
		return nil, newGoldenFileError(path, err)
	}

	if err := golden.parseHeaderLines(headersLines); err != nil {
		return nil, newGoldenFileError(path, err)
	}

	if err := golden.parseBodyLines(bodyLines, bodyIsJSON); err != nil {
		return nil, newGoldenFileError(path, err)
	}

	if golden.WantStatus == 0 && golden.WantHeaders == nil && golden.WantBody == nil {
		return nil, newGoldenFileError(path, fmt.Errorf("no expected status, headers, or body specified"))
	}

	return golden, nil
}

// SaveFile saves a golden file to the given path.
func (g *Golden) SaveFile(path string) error {
	if g.WantStatus == 0 {
		return newGoldenFileError(path, fmt.Errorf("expected status is required"))
	}

	lines := []string{fmt.Sprintf("%d", g.WantStatus)}

	if g.WantHeaders != nil {
		headersDirectives := []string{headersLinePrefix}
		if g.MatchHeadersExactly {
			headersDirectives = append(headersDirectives, "exact")
		}
		lines = append(lines, strings.Join(headersDirectives, " "))
		for key, values := range g.WantHeaders {
			for _, value := range values {
				lines = append(lines, fmt.Sprintf("%s: %s", key, value))
			}
		}
	}

	if g.WantBody != nil {
		bodyDirectives := []string{bodyLinePrefix}
		var content string
		switch bodyVal := g.WantBody.(type) {
		case string:
			content = bodyVal
		case map[string]any, []any:
			bodyDirectives = append(bodyDirectives, "json")
			if g.MatchBodyJSONExactly {
				bodyDirectives = append(bodyDirectives, "exact")
			}
			var err error
			content, err = bodyContentToString(bodyVal)
			if err != nil {
				return newGoldenFileError(path, err)
			}
		case float64, float32, uint64, int64, uint32, int32, uint, int, bool:
			var err error
			content, err = bodyContentToString(bodyVal)
			if err != nil {
				return newGoldenFileError(path, err)
			}
		default:
			return newGoldenFileError(path, fmt.Errorf("unable to marshal body of type %T", bodyVal))
		}

		lines = append(lines, strings.Join(bodyDirectives, " "))
		lines = append(lines, content)
	}

	content := strings.Join(lines, "\n")
	if err := afero.WriteFile(AppFS, path, []byte(content), 0644); err != nil {
		return newGoldenFileError(path, err)
	}

	return nil
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
		case "exact":
			g.MatchHeadersExactly = true
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
	bodyIsJSON := false
	for _, directive := range bodyDirectives[2:] {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		switch directive {
		case "json":
			bodyIsJSON = true
		case "exact":
			g.MatchBodyJSONExactly = true
		default:
			return false, fmt.Errorf("unknown body directive %q", directive)
		}
	}

	if !bodyIsJSON && g.MatchBodyJSONExactly {
		return false, fmt.Errorf("body directive %q requires %q directive", "exact", "json")
	}

	return bodyIsJSON, nil
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

func bodyContentToString(body any) (string, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("unable to marshal body content: %w", err)
	}

	return string(b), nil
}

func newGoldenFileError(path string, err error) error {
	return fmt.Errorf("golden file %q: %w", path, err)
}
