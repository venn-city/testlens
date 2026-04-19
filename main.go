package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var errHelpOK = errors.New("help")

func main() {
	inPath, outPath, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, errHelpOK) {
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

  %s -in <test2json-log-file> [-o output.html]

FLAGS

  -in path   Path to the test2json log file (required to generate a report)
  -o path    Write HTML to this path (default: testlens-report.html)
  -h -help   Show this message and exit (status 0)

  Run with no flags (or only -h/-help) to print this help. If you pass -o without
  -in, an error is reported.

EXAMPLES

  %s -in test-out.log
  %s -in test-out.log -o report.html

`, name, name, name)
}

// parseFlags implements CLI parsing with the stdlib flag package only (no positional args).
func parseFlags(args []string) (inPath, outPath string, err error) {
	fs := flag.NewFlagSet(progName(), flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var help bool
	fs.BoolVar(&help, "h", false, "show help")
	fs.BoolVar(&help, "help", false, "show help")

	outPath = "testlens-report.html"
	fs.StringVar(&outPath, "o", outPath, "write HTML to `path`")
	fs.StringVar(&inPath, "in", "", "path to test2json `log` file")

	// flag.failf calls Usage before returning parse errors; we print help once from main.
	fs.Usage = func() {}

	if err := fs.Parse(args); err != nil {
		return "", "", err
	}
	if help {
		return "", "", errHelpOK
	}
	if fs.NArg() > 0 {
		return "", "", fmt.Errorf("unexpected arguments %q — use -in for the log file", fs.Args())
	}
	if inPath == "" {
		var anyExceptHelp bool
		fs.Visit(func(f *flag.Flag) {
			if f.Name != "h" && f.Name != "help" {
				anyExceptHelp = true
			}
		})
		if anyExceptHelp {
			return "", "", fmt.Errorf("-in is required (path to the test2json log file)")
		}
		return "", "", errHelpOK
	}
	return inPath, outPath, nil
}
