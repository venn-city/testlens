package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ParseTest2JSON reads a test2json log file (one JSON object per line).
func ParseTest2JSON(path string) ([]TestEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	var events []TestEvent
	sc := bufio.NewScanner(f)
	// Large test output lines can exceed default 64KB buffer.
	const maxScan = 64 * 1024 * 1024
	buf := make([]byte, maxScan)
	sc.Buffer(buf, maxScan)

	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var ev TestEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		events = append(events, ev)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	return events, nil
}
