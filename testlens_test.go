// Copyright 2026 Venn City. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"strings"
	"testing"
)

const testLogFile = "test-out.log"

func TestEndToEnd(t *testing.T) {
	outPath := t.TempDir() + "/report.html"

	events, err := ParseTest2JSON(testLogFile)
	if err != nil {
		t.Fatalf("ParseTest2JSON: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected events, got none")
	}

	gantt := EventsToGantt(events, "E2E Test Report")

	if len(gantt.Tasks) == 0 {
		t.Fatal("expected gantt tasks, got none")
	}

	if err := RenderHTML(outPath, gantt); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}

	html, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	body := string(html)

	for _, want := range []string{
		"<!DOCTYPE html>",
		"dhtmlxgantt",
		"E2E Test Report",
		"gantt.parse",
		"fitTaskViewport",
		"id=\"tl-zoom-in\"",
		"Zoom in",
		"id=\"tl-expand-all\"",
		"Collapse all",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("output HTML missing %q", want)
		}
	}
}

func TestParseTest2JSON(t *testing.T) {
	events, err := ParseTest2JSON(testLogFile)
	if err != nil {
		t.Fatalf("ParseTest2JSON: %v", err)
	}

	var starts, runs, passes int
	for _, ev := range events {
		switch ev.Action {
		case "start":
			starts++
		case "run":
			runs++
		case "pass":
			passes++
		}
	}

	if starts == 0 {
		t.Error("expected start events")
	}
	if runs == 0 {
		t.Error("expected run events")
	}
	if passes == 0 {
		t.Error("expected pass events")
	}
}

func TestEventsToGantt(t *testing.T) {
	events, err := ParseTest2JSON(testLogFile)
	if err != nil {
		t.Fatalf("ParseTest2JSON: %v", err)
	}

	gantt := EventsToGantt(events, "Test Title")

	if gantt.Title != "Test Title" {
		t.Errorf("title = %q, want %q", gantt.Title, "Test Title")
	}

	var containers, leaves int
	hasPassingLeaf := false
	for _, task := range gantt.Tasks {
		if task.Type == "project" {
			containers++
		} else {
			leaves++
		}
		if task.Status == "pass" {
			hasPassingLeaf = true
		}
		if task.StartDate == "" || task.EndDate == "" {
			t.Errorf("task %q has empty start/end date", task.Text)
		}
		if task.ID == "" {
			t.Errorf("task %q has empty ID", task.Text)
		}
	}

	if containers == 0 {
		t.Error("expected container (project) tasks for packages")
	}
	if leaves == 0 {
		t.Error("expected leaf tasks for tests")
	}
	if !hasPassingLeaf {
		t.Error("expected at least one passing task")
	}
}

func TestParseTest2JSON_FileNotFound(t *testing.T) {
	_, err := ParseTest2JSON("nonexistent.log")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
