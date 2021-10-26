package golden_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jefflinse/go-itest/golden"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	for _, test := range []struct {
		name       string
		content    string
		wantGolden *golden.Golden
		wantError  string
	}{
		{
			name:    "success, only status specified",
			content: "200",
			wantGolden: &golden.Golden{
				WantStatus:  200,
				WantHeaders: nil,
				WantBody:    nil,
			},
			wantError: "",
		},
		{
			name:    "success, status and headers specified",
			content: "200\n--- headers\nSome-Header: foo\nContent-Type: application/json\nSome-Header: bar",
			wantGolden: &golden.Golden{
				WantStatus: 200,
				WantHeaders: http.Header{
					"Content-Type": []string{"application/json"},
					"Some-Header":  []string{"foo", "bar"},
				},
				WantBody: nil,
			},
			wantError: "",
		},
		{
			name:    "success, status and string body specified",
			content: "200\n--- body\nbody content\nmore content",
			wantGolden: &golden.Golden{
				WantStatus:  200,
				WantHeaders: nil,
				WantBody:    "body content\nmore content",
			},
			wantError: "",
		},
		{
			name:    "success, status headers and body specified with empty lines discarded",
			content: "\n\n\n200\n\n\n--- headers\n\n\nSome-Header: foo\n\n\nAnother-Header: bar\n\n\n--- body\n\n\nbody content\n\n\nmore content",
			wantGolden: &golden.Golden{
				WantStatus: 200,
				WantHeaders: http.Header{
					"Some-Header":    []string{"foo"},
					"Another-Header": []string{"bar"},
				},
				WantBody: "\n\nbody content\n\n\nmore content",
			},
			wantError: "",
		},
		{
			name:    "success, header directives separated by multiple spaces",
			content: "200\n--- headers     \nSome-Header: foo",
			wantGolden: &golden.Golden{
				WantStatus: 200,
				WantHeaders: http.Header{
					"Some-Header": []string{"foo"},
				},
				WantBody: nil,
			},
			wantError: "",
		},
		{
			name:    "success, body directives separated by multiple spaces",
			content: "200\n--- body     json\n{}",
			wantGolden: &golden.Golden{
				WantStatus:  200,
				WantHeaders: nil,
				WantBody:    map[string]interface{}{},
			},
			wantError: "",
		},
		{
			name:    "success, status and JSON body specified",
			content: "200\n--- body json\n{\"foo\":[\"bar\"]}",
			wantGolden: &golden.Golden{
				WantStatus:  200,
				WantHeaders: nil,
				WantBody:    map[string]interface{}{"foo": []interface{}{"bar"}},
			},
			wantError: "",
		},
		{
			name:      "failure, file not found",
			content:   "",
			wantError: "not found",
		},
		{
			name:      "failure, empty file or no expectations defined",
			content:   "",
			wantError: "no expected status, headers, or body specified",
		},
		{
			name:      "failure, invalid status",
			content:   "foo",
			wantError: `invalid status "foo"`,
		},
		{
			name:      "failure, headers before status",
			content:   "--- headers\nContent-Type: application/xml",
			wantError: `invalid status "--- headers"`,
		},
		{
			name:      "failure, body before status",
			content:   "--- body\nfoo",
			wantError: `invalid status "--- body"`,
		},
		{
			name:      "failure, duplicate headers section",
			content:   "200\n--- headers\nContent-Type: application/xml\n--- headers\nContent-Type: application/xml",
			wantError: `duplicate headers directive`,
		},
		{
			name:      "failure, duplicate body section",
			content:   "200\n--- body\nfoo\n--- body\nfoo",
			wantError: `duplicate body directive`,
		},
		{
			name:      "failure, headers after body",
			content:   "200\n--- body\nfoo\n--- headers\nSome-Header: foo",
			wantError: `headers directive must come before body directive`,
		},
		{
			name:      "failure, unknown headers directive",
			content:   "200\n--- headers foo",
			wantError: `unknown headers directive "foo"`,
		},
		{
			name:      "failure, unknown body directive",
			content:   "200\n--- body foo",
			wantError: `unknown body directive "foo"`,
		},
		{
			name:      "failure, unexpected line after status",
			content:   "200\nfoo",
			wantError: `unexpected line "foo"`,
		},
		{
			name:      "failure, invalid header linie",
			content:   "200\n--- headers\nfoo",
			wantError: `invalid header "foo"`,
		},
		{
			name:      "failure, invalid header key",
			content:   "200\n--- headers\n: foo",
			wantError: `invalid header key ": foo"`,
		},
		{
			name:      "failure, invalid body JSON",
			content:   "200\n--- body json\n{foo",
			wantError: "invalid JSON body: invalid character 'f' looking for beginning of object key string\n---\n{foo\n---",
		},
	} {
		path := "/test.golden"
		t.Run(test.name, func(t *testing.T) {
			golden.AppFS = afero.NewMemMapFs()

			// special case
			if test.name != "failure, file not found" {
				if err := afero.WriteFile(golden.AppFS, path, []byte(test.content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			g, err := golden.LoadFile(path)
			if test.wantError != "" {
				assert.EqualError(t, err, fmt.Sprintf("golden file %q: %s", path, test.wantError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantGolden, g)
			}
		})
	}
}

func TestSaveFile(t *testing.T) {
	for _, test := range []struct {
		name      string
		g         *golden.Golden
		wantError string
	}{
		{
			name: "success, status only",
			g: &golden.Golden{
				WantStatus:  200,
				WantHeaders: nil,
				WantBody:    nil,
			},
		},
		{
			name: "success, status and headers",
			g: &golden.Golden{
				WantStatus: 200,
				WantHeaders: http.Header{
					"Content-Type": []string{"application/json"},
					"Some-Header":  []string{"foo", "bar"},
				},
			},
		},
		{
			name: "success, status and string body",
			g: &golden.Golden{
				WantStatus: 200,
				WantBody:   "foo\nbar\nbaz",
			},
		},
		{
			name: "success, status and string body",
			g: &golden.Golden{
				WantStatus: 200,
				WantBody:   "foo\nbar\nbaz",
			},
		},
		{
			name: "success, status and JSON body",
			g: &golden.Golden{
				WantStatus: 200,
				WantBody:   map[string]interface{}{"foo": []interface{}{"bar"}},
			},
		},
		{
			name: "success, status headers and body",
			g: &golden.Golden{
				WantStatus: 200,
				WantHeaders: http.Header{
					"Content-Type": []string{"application/json"},
					"Some-Header":  []string{"foo", "bar"},
				},
				WantBody: "foo\nbar\nbaz",
			},
		},
		{
			name:      "failure, no status",
			g:         &golden.Golden{},
			wantError: `expected status is required`,
		},
		{
			name: "failure, invalid body JSON",
			g: &golden.Golden{
				WantStatus: 200,
				WantBody:   make(chan int),
			},
			wantError: "unable to marshal body of type chan int",
		},
		{
			name: "failure, invalid body JSON element",
			g: &golden.Golden{
				WantStatus: 200,
				WantBody:   map[string]interface{}{"foo": make(chan int)},
			},
			wantError: "unable to marshal body content: json: unsupported type: chan int",
		},
		{
			name: "failure, can't write file",
			g: &golden.Golden{
				WantStatus: 200,
			},
			wantError: "operation not permitted",
		},
	} {
		path := "/test.golden"
		t.Run(test.name, func(t *testing.T) {
			golden.AppFS = afero.NewMemMapFs()

			// special case
			if test.name == "failure, can't write file" {
				golden.AppFS = afero.NewReadOnlyFs(golden.AppFS)
			}

			err := test.g.SaveFile(path)
			if test.wantError != "" {
				assert.EqualError(t, err, fmt.Sprintf("golden file %q: %s", path, test.wantError))
			} else {
				assert.NoError(t, err)
				g, err := golden.LoadFile(path)
				assert.NoError(t, err)
				assert.Equal(t, test.g, g)
			}
		})
	}
}
