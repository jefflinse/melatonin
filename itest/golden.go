package itest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

const (
	MatchModeUnset = iota
	MatchModeExact
	MatchModeContains
)

type Golden struct {
	WantStatus  int
	WantHeaders http.Header
	WantBody    interface{}

	headerMatchMode int
	bodyIsJSON      bool
}

func ParseGoldenFileFs(fs afero.Fs, filename string) (*Golden, error) {
	if exists, err := afero.Exists(fs, filename); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("golden file %q does not exist", filename)
	}

	b, err := afero.ReadFile(fs, filename)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, fmt.Errorf("golden file %q is empty", filename)
	}

	content := string(b)
	golden := &Golden{}

	matcher := search.New(language.English, search.IgnoreCase)

	headersLineStart, _ := matcher.IndexString(content, "--- headers")
	var headersLineEnd int
	if headersLineStart != -1 {
		headersLineEnd, _ = matcher.IndexString(content[headersLineStart:], "\n")
		if headersLineEnd == -1 {
			return nil, fmt.Errorf("golden file %q contains invalid header directive", filename)
		}

		headersLine := content[headersLineStart : headersLineStart+headersLineEnd]
		fmt.Println("HEADERS LINE:", headersLine)
		headerDirectives := strings.Split(headersLine, " ")
		for _, directive := range headerDirectives[2:] {
			switch directive {
			case "exact":
				if golden.headerMatchMode != MatchModeUnset {
					return nil, fmt.Errorf("golden file %q contains conflicting header directives", filename)
				}
				golden.headerMatchMode = MatchModeExact
			case "contains":
				if golden.headerMatchMode != MatchModeUnset {
					return nil, fmt.Errorf("golden file %q contains conflicting header directives", filename)
				}
				golden.headerMatchMode = MatchModeContains
			default:
				return nil, fmt.Errorf("golden file %q contains invalid header directive %q", filename, directive)
			}
		}
	} else {
		fmt.Println("no header directive")
	}

	bodyLineStart, _ := matcher.IndexString(content, "--- body")
	var bodyLineEnd int
	if bodyLineStart != -1 {
		bodyLineEnd, _ = matcher.IndexString(content[bodyLineStart:], "\n")
		if bodyLineEnd == -1 {
			return nil, fmt.Errorf("golden file %q contains invalid body directive", filename)
		}

		bodyLine := content[bodyLineStart : bodyLineStart+bodyLineEnd]
		fmt.Println("BODY LINE:", bodyLine)
		bodyDirectives := strings.Split(bodyLine, " ")
		for _, directive := range bodyDirectives[2:] {
			switch directive {
			case "json":
				golden.bodyIsJSON = true
			default:
				return nil, fmt.Errorf("golden file %q: invalid body directive %q", filename, directive)
			}
		}
	} else {
		fmt.Println("no body directive")
	}

	headersContent := content[headersLineStart+headersLineEnd+1 : bodyLineStart-1]
	bodyContent := content[bodyLineStart+bodyLineEnd+1:]

	fmt.Printf("HEADERS CONTENT:\n[%s]\n", headersContent)
	fmt.Printf("BODY CONTENT\n[%s]\n", bodyContent)

	golden.WantHeaders = http.Header{}
	headerLines := strings.Split(headersContent, "\n")
	for _, headerLine := range headerLines {
		if len(headerLine) == 0 {
			continue
		}

		parts := strings.SplitN(headerLine, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("golden file %q: invalid header line %q", filename, headerLine)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		golden.WantHeaders.Add(key, value)
	}

	if golden.bodyIsJSON {
		if json.Unmarshal([]byte(bodyContent), &golden.WantBody) != nil {
			return nil, fmt.Errorf("golden file %q: invalid JSON body", filename)
		}
	} else {
		golden.WantBody = bodyContent
	}

	return golden, nil
}
