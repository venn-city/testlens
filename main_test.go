package main

import (
	"errors"
	"testing"
)

func TestParseFlags_inAndOut(t *testing.T) {
	in, out, err := parseFlags([]string{"-in", "a.log", "-o", "x.html"})
	if err != nil {
		t.Fatal(err)
	}
	if in != "a.log" || out != "x.html" {
		t.Fatalf("in=%q out=%q", in, out)
	}
}

func TestParseFlags_inOnly(t *testing.T) {
	in, out, err := parseFlags([]string{"-in", "b.log"})
	if err != nil {
		t.Fatal(err)
	}
	if in != "b.log" || out != "testlens-report.html" {
		t.Fatalf("in=%q out=%q", in, out)
	}
}

func TestParseFlags_helpNoArgs(t *testing.T) {
	_, _, err := parseFlags(nil)
	if !errors.Is(err, errHelpOK) {
		t.Fatalf("err=%v", err)
	}
}

func TestParseFlags_helpH(t *testing.T) {
	_, _, err := parseFlags([]string{"-h"})
	if !errors.Is(err, errHelpOK) {
		t.Fatalf("err=%v", err)
	}
}

func TestParseFlags_oWithoutIn(t *testing.T) {
	_, _, err := parseFlags([]string{"-o", "out.html"})
	if err == nil || errors.Is(err, errHelpOK) {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestParseFlags_positionalRejected(t *testing.T) {
	_, _, err := parseFlags([]string{"-in", "a.log", "extra"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseFlags_unknownFlag(t *testing.T) {
	_, _, err := parseFlags([]string{"-in", "a.log", "-bogus"})
	if err == nil {
		t.Fatal("expected error")
	}
}
