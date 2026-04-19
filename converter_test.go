// Copyright 2026 Venn City. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
	"time"
)

func TestEventsToGantt_ChronologicalOrder(t *testing.T) {
	t0 := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	pkgA := modulePrefix + "/alpha"
	pkgB := modulePrefix + "/beta"

	events := []TestEvent{
		{Time: t0.Add(1 * time.Second), Action: "start", Package: pkgB},
		{Time: t0, Action: "start", Package: pkgA},
		{Time: t0.Add(2 * time.Second), Action: "run", Package: pkgA, Test: "TestZebra"},
		{Time: t0.Add(4 * time.Second), Action: "run", Package: pkgA, Test: "TestApple"},
		{Time: t0.Add(6 * time.Second), Action: "pass", Package: pkgA, Test: "TestZebra", Elapsed: 1},
		{Time: t0.Add(7 * time.Second), Action: "pass", Package: pkgA, Test: "TestApple", Elapsed: 1},
		{Time: t0.Add(3 * time.Second), Action: "run", Package: pkgB, Test: "TestB"},
		{Time: t0.Add(8 * time.Second), Action: "pass", Package: pkgB, Test: "TestB", Elapsed: 1},
		{Time: t0.Add(3 * time.Second), Action: "run", Package: pkgA, Test: "TestParent/SubLate"},
		{Time: t0.Add(2 * time.Second), Action: "run", Package: pkgA, Test: "TestParent/SubEarly"},
		{Time: t0.Add(5 * time.Second), Action: "pass", Package: pkgA, Test: "TestParent/SubEarly", Elapsed: 0.5},
		{Time: t0.Add(6 * time.Second), Action: "pass", Package: pkgA, Test: "TestParent/SubLate", Elapsed: 0.5},
		{Time: t0.Add(2 * time.Second), Action: "run", Package: pkgA, Test: "TestParent"},
		{Time: t0.Add(6 * time.Second), Action: "pass", Package: pkgA, Test: "TestParent", Elapsed: 1},
	}

	g := EventsToGantt(events, "")

	rowOf := func(match func(GanttTask) bool) int {
		for i, task := range g.Tasks {
			if match(task) {
				return i
			}
		}
		return -1
	}

	iAlpha := rowOf(func(tk GanttTask) bool {
		return tk.Type == "project" && tk.Text == "alpha" && tk.Parent == ""
	})
	iBeta := rowOf(func(tk GanttTask) bool {
		return tk.Type == "project" && tk.Text == "beta" && tk.Parent == ""
	})
	if iAlpha < 0 || iBeta < 0 {
		t.Fatalf("missing package rows: alpha=%d beta=%d", iAlpha, iBeta)
	}
	if iAlpha >= iBeta {
		t.Fatalf("packages not chronological by start: alpha@%d before beta@%d", iAlpha, iBeta)
	}

	iZebra := rowOf(func(tk GanttTask) bool { return tk.Text == "TestZebra" && tk.Type == "" })
	iApple := rowOf(func(tk GanttTask) bool { return tk.Text == "TestApple" && tk.Type == "" })
	if iZebra < 0 || iApple < 0 {
		t.Fatalf("missing leaf tests: zebra=%d apple=%d", iZebra, iApple)
	}
	if iZebra >= iApple {
		t.Fatalf("tests under package not chronological: Zebra@%d Apple@%d", iZebra, iApple)
	}

	iEarly := rowOf(func(tk GanttTask) bool { return tk.Text == "SubEarly" })
	iLate := rowOf(func(tk GanttTask) bool { return tk.Text == "SubLate" })
	if iEarly < 0 || iLate < 0 {
		t.Fatalf("missing subtests: early=%d late=%d", iEarly, iLate)
	}
	if iEarly >= iLate {
		t.Fatalf("subtests not chronological: early@%d late@%d", iEarly, iLate)
	}
}
