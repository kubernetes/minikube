// test_results converts JSON test results to HTML
//
// usage:
//   json_tests_to_html -in <results.json> -out <results.html>

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"
)

var (
	inPath  = flag.String("in", "", "path to JSON input file")
	outPath = flag.String("out", "", "path to HTML output file")
)

type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string

	EmbeddedLog []string
}

type TestGroup struct {
	Test   string
	Hidden bool
	Status string
	Start  time.Time
	End    time.Time
	Events []TestEvent
}

// parseJSON is a very forgiving JSON parser.
func parseJSON(path string) ([]TestEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	events := []TestEvent{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Go's -json output is line-by-line JSON events
		b := scanner.Bytes()
		if b[0] == '{' {
			ev := TestEvent{}
			err = json.Unmarshal(b, &ev)
			if err != nil {
				fmt.Println("ERROR: %v", err)
				continue
			}
			events = append(events, ev)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, err
}

// group events by their test name
func processEvents(evs []TestEvent) []TestGroup {
	gm := map[string]int{}
	groups := []TestGroup{}
	for _, e := range evs {
		if e.Test == "" {
			continue
		}
		index, ok := gm[e.Test]
		if !ok {
			index = len(groups)
			groups = append(groups, TestGroup{
				Test:  e.Test,
				Start: e.Time,
			})
			gm[e.Test] = index
		}
		groups[index].Events = append(groups[index].Events, e)
		groups[index].Status = e.Action
	}

	// Hide ancestors
	for k, v := range gm {
		for k2 := range gm {
			if strings.HasPrefix(k2, fmt.Sprintf("%s/", k)) {
				groups[v].Hidden = true
			}
		}
	}

	return groups
}

type Content struct {
	Groups []TestGroup
}

func generateHTML(groups []TestGroup) ([]byte, error) {
	t, err := template.New("out").Parse(`
<html><head><style>
body { font-family: "Arial"; }
.pass { background-color: #9f9; }
.fail { background-color: #f99; }
</style></head><body><ul>
{{ range .Groups }}{{ if not .Hidden }}
<li><strong>{{ .Test }}: <span class="{{ .Status }}">{{ .Status }}</span></strong>{{ if eq .Status "fail" }}<pre>{{ range .Events }}{{ .Output }}{{ end }}</pre>{{ end }}</li>
{{ end }}{{ end }}
</ul></body></html>`)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := t.ExecuteTemplate(&b, "out", &Content{Groups: groups}); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func main() {
	flag.Parse()
	if *inPath == "" {
		panic("must provide path to JSON input file")
	}
	if *outPath == "" {
		panic("must provide path to HTML output file")
	}

	events, err := parseJSON(*inPath)
	if err != nil {
		panic(fmt.Sprintf("json: %v", err))
	}
	groups := processEvents(events)
	html, err := generateHTML(groups)
	if err != nil {
		panic(fmt.Sprintf("html: %v", err))
	}
	if err := ioutil.WriteFile(*outPath, html, 0644); err != nil {
		panic(fmt.Sprintf("write: %v", err))
	}
}
