// Copyright 2026 Venn City. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"sort"
	"strings"
	"time"
)

const modulePrefix = "github.com/venn-city/venn-platform"

// EventsToGantt builds DHTMLX task rows from test2json events.
func EventsToGantt(events []TestEvent, title string) GanttData {
	pkgStarts := make(map[string]time.Time) // key: Package (full import path)
	runStarts := make(map[string]time.Time) // key: package + "\x00" + test

	for _, ev := range events {
		switch ev.Action {
		case "start":
			if ev.Package != "" && ev.Test == "" {
				pkgStarts[ev.Package] = ev.Time
			}
		case "run":
			if ev.Package != "" && ev.Test != "" {
				k := runKey(ev.Package, ev.Test)
				runStarts[k] = ev.Time
			}
		}
	}

	var results []testResult
	for _, ev := range events {
		switch ev.Action {
		case "pass", "fail", "skip":
			if ev.Package == "" || ev.Test == "" {
				continue
			}
			k := runKey(ev.Package, ev.Test)
			start, ok := runStarts[k]
			if !ok {
				if ev.Elapsed > 0 {
					start = ev.Time.Add(-seconds(ev.Elapsed))
				} else {
					start = ev.Time
				}
			}
			end := ev.Time
			status := ev.Action
			if status == "pass" {
				status = "pass"
			}
			results = append(results, testResult{
				PackageFull: ev.Package,
				PkgShort:    stripModule(ev.Package),
				FullTest:    ev.Test,
				Base:        baseTestName(ev.Test),
				Start:       start,
				End:         end,
				Status:      status,
			})
		}
	}

	byPkg := make(map[string][]testResult)
	for _, r := range results {
		byPkg[r.PkgShort] = append(byPkg[r.PkgShort], r)
	}

	pkgOrder := make([]string, 0, len(byPkg))
	for p := range byPkg {
		pkgOrder = append(pkgOrder, p)
	}
	sort.Slice(pkgOrder, func(i, j int) bool {
		si := pkgSortStart(byPkg[pkgOrder[i]], pkgStarts)
		sj := pkgSortStart(byPkg[pkgOrder[j]], pkgStarts)
		if !si.Equal(sj) {
			return si.Before(sj)
		}
		return pkgOrder[i] < pkgOrder[j]
	})

	var tasks []GanttTask

	for _, pkgShort := range pkgOrder {
		rs := byPkg[pkgShort]
		pkgFull := rs[0].PackageFull
		pid := makeID("pkg", pkgShort)

		pkgStart, ok := pkgStarts[pkgFull]
		if !ok {
			pkgStart = minTime(rs)
		}
		pkgEnd := maxTime(rs)
		if pkgEnd.Before(pkgStart) {
			pkgEnd = pkgStart.Add(time.Millisecond)
		}

		tasks = append(tasks, GanttTask{
			ID:        pid,
			Text:      pkgShort,
			StartDate: formatDT(pkgStart),
			EndDate:   formatDT(pkgEnd),
			Open:      false,
			Progress:  1,
			Status:    "container",
			Type:      "project",
		})

		byBase := groupByBase(rs)
		bases := make([]string, 0, len(byBase))
		for b := range byBase {
			bases = append(bases, b)
		}
		sort.Slice(bases, func(i, j int) bool {
			gi, gj := byBase[bases[i]], byBase[bases[j]]
			si, sj := minTime(gi), minTime(gj)
			if !si.Equal(sj) {
				return si.Before(sj)
			}
			return bases[i] < bases[j]
		})

		for _, base := range bases {
			group := byBase[base]
			top := findExact(group, base)
			subs := findSubs(group, base)

			tid := makeID("test", pkgShort+"#"+base)

			if len(subs) == 0 && top != nil {
				tasks = append(tasks, leafTask(tid, top.FullTest, pid, *top))
				continue
			}

			var parentStart, parentEnd time.Time
			var parentStatus string
			var parentProgress float64

			if top != nil {
				parentStart = top.Start
				parentEnd = top.End
				parentStatus = top.Status
				parentProgress = progressFor(parentStatus)
			} else {
				parentStart = minTime(subs)
				parentEnd = maxTime(subs)
				parentStatus = "container"
				parentProgress = 0
			}
			if parentEnd.Before(parentStart) {
				parentEnd = parentStart.Add(time.Millisecond)
			}

			tasks = append(tasks, GanttTask{
				ID:        tid,
				Text:      base,
				StartDate: formatDT(parentStart),
				EndDate:   formatDT(parentEnd),
				Parent:    pid,
				Open:      false,
				Progress:  parentProgress,
				Status:    parentStatus,
				Type:      "project",
			})

			for _, s := range subs {
				if top != nil && s.FullTest == top.FullTest {
					continue
				}
				lid := makeID("leaf", pkgShort+"#"+s.FullTest)
				label := strings.TrimPrefix(s.FullTest, base+"/")
				tasks = append(tasks, leafTask(lid, label, tid, s))
			}
		}
	}

	if title == "" {
		title = "Go test timeline"
	}
	return GanttData{Tasks: tasks, Links: []any{}, Title: title}
}

type testResult struct {
	PackageFull string
	PkgShort    string
	FullTest    string
	Base        string
	Start       time.Time
	End         time.Time
	Status      string
}

func runKey(pkg, test string) string {
	return pkg + "\x00" + test
}

func stripModule(pkg string) string {
	p := strings.TrimPrefix(pkg, modulePrefix)
	p = strings.TrimPrefix(p, "/")
	return p
}

func baseTestName(full string) string {
	if i := strings.Index(full, "/"); i >= 0 {
		return full[:i]
	}
	return full
}

func groupByBase(rs []testResult) map[string][]testResult {
	m := make(map[string][]testResult)
	for _, r := range rs {
		m[r.Base] = append(m[r.Base], r)
	}
	return m
}

func findExact(rs []testResult, base string) *testResult {
	for i := range rs {
		if rs[i].FullTest == base {
			return &rs[i]
		}
	}
	return nil
}

func findSubs(rs []testResult, base string) []testResult {
	var out []testResult
	prefix := base + "/"
	for _, r := range rs {
		if strings.HasPrefix(r.FullTest, prefix) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Start.Equal(out[j].Start) {
			return out[i].Start.Before(out[j].Start)
		}
		return out[i].FullTest < out[j].FullTest
	})
	return out
}

// pkgSortStart is the timeline position used to order top-level packages.
func pkgSortStart(rs []testResult, pkgStarts map[string]time.Time) time.Time {
	if len(rs) == 0 {
		return time.Time{}
	}
	pkgFull := rs[0].PackageFull
	if t, ok := pkgStarts[pkgFull]; ok {
		return t
	}
	return minTime(rs)
}

func minTime(rs []testResult) time.Time {
	if len(rs) == 0 {
		return time.Time{}
	}
	t := rs[0].Start
	for _, r := range rs[1:] {
		if r.Start.Before(t) {
			t = r.Start
		}
	}
	return t
}

func maxTime(rs []testResult) time.Time {
	if len(rs) == 0 {
		return time.Time{}
	}
	t := rs[0].End
	for _, r := range rs[1:] {
		if r.End.After(t) {
			t = r.End
		}
	}
	return t
}

func leafTask(id, text, parent string, tr testResult) GanttTask {
	end := tr.End
	start := tr.Start
	if end.Before(start) {
		end = start.Add(time.Millisecond)
	}
	return GanttTask{
		ID:        id,
		Text:      text,
		StartDate: formatDT(start),
		EndDate:   formatDT(end),
		Parent:    parent,
		Open:      false,
		Progress:  progressFor(tr.Status),
		Status:    tr.Status,
	}
}

func progressFor(status string) float64 {
	switch status {
	case "pass":
		return 1
	case "skip":
		return 0.5
	default:
		return 0
	}
}

func seconds(f float64) time.Duration {
	return time.Duration(f * float64(time.Second))
}

// formatDT matches DHTMLX default xml_date style (day-month-year with time).
func formatDT(t time.Time) string {
	return t.Format("02-01-2006 15:04:05")
}

func makeID(prefix, s string) string {
	h := sha256.Sum256([]byte(s))
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(h[:12])
}
