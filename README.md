# testlens

CLI that reads a Go **test2json** log (one JSON object per line; see `go help testflag` / `cmd/test2json`) and writes a self-contained **HTML** file with an interactive **DHTMLX Gantt** timeline.

## Screenshots

Package rows on the timeline (collapsed tree, duration column, second-level time scale):

![testlens report — packages on the timeline](example_1.png)

Expanded tree with pass (green) / fail (red) tests, tooltips, and zoom controls:

![testlens report — expanded tests and status colors](example_2.png)

## Run

```bash
go run . -in test2json.log [-o output.html]
```

Default output path: `testlens-report.html`. The input file is always passed with **`-in`**, which matches idiomatic use of Go’s [`flag`](https://pkg.go.dev/flag) package (no positional arguments). Use `go run . -help` for full usage.

Input must be JSON lines compatible with `encoding/json` unmarshaling into `TestEvent` in `model.go` (`Time`, `Action`, `Package`, `Test`, `Elapsed`, `Output`).

### Install or run without cloning (module path)

From any machine with Go installed:

```bash
# Run the latest version without installing a binary (downloads module, then runs)
go run github.com/venn-city/testlens@latest -in test2json.log -o testlens-report.html

# Install a `testlens` binary into $(go env GOPATH)/bin (ensure that directory is on your PATH)
go install github.com/venn-city/testlens@latest
testlens -in test2json.log -o testlens-report.html
```

To pin a version, replace `@latest` with a tag or commit (for example `@v0.1.0`).

### From a checkout of this repository

```bash
cd /path/to/testlens
go run . -in /path/to/test2json.log -o report.html
# or build once and run the binary:
go build -o testlens .
./testlens -in /path/to/test2json.log -o report.html
```

### CI example (Jenkins): produce an HTML artifact

Typical flow: run tests with JSON event output, write one line-delimited JSON log, run testlens, then archive the HTML so it appears in the build’s **Artifacts** (or is picked up by your artifact repository).

The following **declarative** pipeline assumes the job checks out a repo whose tests produce `test2json.log` (adjust paths and the `go test` line to match your project). It uses `go run` against this module; you can swap that for `go install` and a `testlens` binary if you prefer.

```groovy
pipeline {
    agent any
    stages {
        stage('Tests (test2json log)') {
            steps {
                sh '''
                    go version
                    go test -json ./... 2>&1 | tee test2json.log || true
                '''
            }
        }
        stage('testlens HTML report') {
            steps {
                sh '''
                    go run github.com/venn-city/testlens@latest \
                      -in test2json.log \
                      -o testlens-report.html
                '''
            }
        }
    }
    post {
        always {
            archiveArtifacts artifacts: 'test2json.log,testlens-report.html', fingerprint: true, allowEmptyArchive: true
        }
    }
}
```

Notes:

- **`tee`** keeps a copy of the test2json stream on disk while tests still print to the console. Drop `|| true` if you want the stage to fail when tests fail; keep it if you always want the report generated for debugging.
- **`archiveArtifacts`** stores the files on the Jenkins controller and lists them on the build page; use your organization’s pattern (S3, Artifactory, etc.) if you upload artifacts elsewhere.
- If testlens lives **in the same repo** as the tests, replace the `go run` line with a checkout-relative path, for example: `go run ./path/to/testlens -in test2json.log -o testlens-report.html` (or `go run .` from inside the `testlens` module directory).

## Layout

| File                     | Responsibility                                                                                       |
| ------------------------ | ---------------------------------------------------------------------------------------------------- |
| `main.go`                | Argument parsing, orchestration                                                                      |
| `parser.go`              | Read log; large line buffer for huge `output` lines                                                  |
| `model.go`               | `TestEvent`, `GanttTask`, `GanttData`                                                                |
| `converter.go`           | Package → test → subtest tree; strips module prefix `github.com/venn-city/venn-platform` for display |
| `renderer.go`            | Embeds `templates/gantt.gohtml`, writes HTML                                                         |
| `templates/gantt.gohtml` | DHTMLX Gantt (CDN), readonly chart, status colors                                                    |

## Behavior notes

- Rows are driven by **`pass` / `fail` / `skip`** events with a `Test` name; **`run`** supplies start times when present.
- **Package** rows appear only for packages that have at least one finished test in the log.
- **test2json** does not include `_test.go` file paths; grouping is package + test function + subtest (after `/`).

## Changing the report

- **Data shape / hierarchy**: edit `converter.go`.
- **Styling / Gantt config**: edit `templates/gantt.gohtml` (DHTMLX `gantt.config`, CSS, templates).
- Prefer **stdlib only**; avoid new module dependencies unless there is a strong reason.
