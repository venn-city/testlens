package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var errShowHelp = errors.New("help")

func main() {
	inPath, outPath, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, errShowHelp) {
			printHelp(os.Stdout)
			return
		}
		printHelp(os.Stdout)
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(2)
	}

	events, err := ParseTest2JSON(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testlens: %v\n", err)
		os.Exit(1)
	}

	gantt := EventsToGantt(events, fmt.Sprintf("Test lens — %s", inPath))

	if err := RenderHTML(outPath, gantt); err != nil {
		fmt.Fprintf(os.Stderr, "testlens: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d tasks)\n", outPath, len(gantt.Tasks))
}

func progName() string {
	return filepath.Base(os.Args[0])
}

func printHelp(w io.Writer) {
	name := progName()
	fmt.Fprintf(w, `testlens — turn a Go test2json log into an interactive HTML Gantt timeline.

Each line of the input file must be one JSON object as produced by Go's test2json
converter (see "go doc cmd/test2json"). The output is a single self-contained HTML
file that loads DHTMLX Gantt from a CDN and shows packages and tests on a timeline,
colored by pass, fail, or skip.

HOW TO CAPTURE INPUT

  Run tests with JSON event output and save stdout to a file, for example:

    go test -json ./... > test-out.log

  Or pipe an existing test2json stream to a file however you normally produce it.

USAGE

  %s <test2json-log-file> [options]

OPTIONS

  -o, --output <path>   Write HTML to this path (default: testlens-report.html)
  -h, --help            Show this message and exit

EXAMPLES

  %s test-out.log
  %s test-out.log -o report.html
  %s -o report.html test-out.log

`, name, name, name, name)
}

// parseArgs accepts `testlens in.log -o out.html` or `testlens -o out.html in.log`.
func parseArgs(args []string) (inPath, outPath string, err error) {
	outPath = "testlens-report.html"
	var positional []string
	sawOutputFlag := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			return "", "", errShowHelp
		case a == "-o" || a == "--output":
			sawOutputFlag = true
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value after %s", a)
			}
			outPath = args[i+1]
			i++
		case strings.HasPrefix(a, "-"):
			return "", "", fmt.Errorf("unknown flag %q (try --help)", a)
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) == 0 {
		if !sawOutputFlag {
			// No args, or only --help-style invocation: same as -h (success, no stderr).
			return "", "", errShowHelp
		}
		return "", "", fmt.Errorf("missing test2json log file (pass one path, or use --help)")
	}
	if len(positional) > 1 {
		return "", "", fmt.Errorf("expected exactly one test2json log file, got %d", len(positional))
	}
	return positional[0], outPath, nil
}
