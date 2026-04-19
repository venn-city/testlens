// Copyright 2026 Venn City. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package main

import "time"

// TestEvent matches cmd/test2json JSON lines.
// See https://pkg.go.dev/cmd/test2json
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"`
	Output  string    `json:"Output,omitempty"`
}

// GanttData is passed to the HTML template and to DHTMLX Gantt.parse().
type GanttData struct {
	Tasks []GanttTask `json:"tasks"`
	Links []any       `json:"links"`
	Title string      `json:"-"` // only for template, not in JSON blob if we use separate field
}

// GanttTask is a DHTMLX Gantt task row (JSON tags match library expectations).
type GanttTask struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Parent    string  `json:"parent,omitempty"`
	Open      bool    `json:"open"`
	Progress  float64 `json:"progress"`
	Status    string  `json:"status"` // pass, fail, skip, running, container
	Type      string  `json:"type,omitempty"`
}
