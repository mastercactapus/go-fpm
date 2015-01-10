package main

import (
	"github.com/blang/semver"
	"testing"
)

func TestParseDown(t *testing.T) {
	check := func(input, expected string) {
		ex, err := semver.Parse(expected)
		if err != nil {
			t.Fatalf("Test error parsing expected '%s': %s\n", expected, err.Error())
		}
		sv, err := parseDown(input)
		if err != nil {
			t.Fatalf("Error parsing '%s': %s\n", input, err.Error())
		}
		if sv.NE(ex) {
			t.Fatalf("Error got '%s' but expected '%s'\n", sv.String(), ex.String())
		}
	}

	check("1.2.3", "1.2.3")
	check("1.2", "1.2.0")
	check("*", "0.0.0")
	check("1.x", "1.0.0")
	check("1.2.X", "1.2.0")
	check("", "0.0.0")
	check("1", "1.0.0")
	check("1.*", "1.0.0")
	check("1.0.0-rc.1", "1.0.0-rc.1")
	check("1.0.0-rc", "1.0.0-rc")
	check("1.0.0-rc+foo", "1.0.0-rc+foo")
	check("1.0.0+foo", "1.0.0+foo")
}
func TestParseUp(t *testing.T) {
	check := func(input, expected string, rounded bool) {
		ex, err := semver.Parse(expected)
		if err != nil {
			t.Fatalf("Test error parsing expected '%s': %s\n", expected, err.Error())
		}
		sv, r, err := parseUp(input)
		if err != nil {
			t.Fatalf("Error parsing '%s': %s\n", input, err.Error())
		}
		if sv.NE(ex) {
			t.Fatalf("Error, got '%s' but expected '%s'\n", sv.String(), ex.String())
		}
		if r != rounded {
			t.Fatalf("Error, rounding reported '%b' but expected '%b'\n", r, rounded)
		}
	}

	check("1.2.3", "1.2.3", false)
	check("1.2", "1.3.0", true)
	check("1.x", "2.0.0", true)
	check("1.2.X", "1.3.0", true)
	check("1", "2.0.0", true)
	check("1.*", "2.0.0", true)
	check("1.0.0-rc.1", "1.0.0-rc.1", false)
	check("1.0.0-rc", "1.0.0-rc", false)
	check("1.0.0-rc+foo", "1.0.0-rc+foo", false)
	check("1.0.0+foo", "1.0.0+foo", false)
}
