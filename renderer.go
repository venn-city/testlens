// Copyright 2026 Venn City. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
)

//go:embed templates/gantt.gohtml
var ganttTemplate string

// RenderHTML writes the final HTML file.
func RenderHTML(outPath string, data GanttData) error {
	tmpl, err := template.New("gantt").Parse(ganttTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	payload, err := json.Marshal(struct {
		Tasks []GanttTask `json:"tasks"`
		Links []any       `json:"links"`
	}{
		Tasks: data.Tasks,
		Links: data.Links,
	})
	if err != nil {
		return fmt.Errorf("marshal gantt json: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Title     string
		GanttJSON template.JS
	}{
		Title:     data.Title,
		GanttJSON: template.JS(payload),
	})
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
