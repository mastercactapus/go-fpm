package semver

import (
	"github.com/blang/semver"
	"testing"
)

func _semverReqCheck(t *testing.T) func(string, []string, []string) {
	return func(req string, good []string, bad []string) {
		svr, err := NewSemverRequirements(req)
		if err != nil {
			t.Fatalf("Failed to parse requirement string '%s': %s\n", req, err.Error())
		}

		t.Logf("Interpereted '%s' as '%s'\n", req, svr.String())
		for _, v := range good {
			sv, err := semver.Parse(v)
			if err != nil {
				t.Fatalf("Bad test, failed to parse semver '%s': %s\n", v, err.Error())
			}
			if !svr.SatisfiedBy(sv) {
				t.Errorf("Range '%s' rejected valid version '%s'", req, v)
			}
		}
		for _, v := range bad {
			sv, err := semver.Parse(v)
			if err != nil {
				t.Fatalf("Bad test, failed to parse semver '%s': %s\n", v, err.Error())
			}
			if svr.SatisfiedBy(sv) {
				t.Errorf("Range '%s' accepted invalid version '%s'", req, v)
			}
		}
	}
}

func TestNewSemverRequirements_direct(t *testing.T) {
	check := _semverReqCheck(t)

	// examples from the node-semver readme
	check(">=1.2.7", []string{"1.2.7", "1.2.8", "2.5.3", "1.3.9"}, []string{"1.2.6", "1.1.0"})
	check(">=1.2.7 <1.3.0", []string{"1.2.7", "1.2.8", "1.2.99"}, []string{"1.2.6", "1.3.0", "1.1.0"})
	check("1.2.7 || >=1.2.9 <2.0.0", []string{"1.2.7", "1.2.9", "1.4.6"}, []string{"1.2.8", "2.0.0"})
	check(">1.2.3-alpha.3", []string{"1.2.3-alpha.7", "3.4.5"}, []string{"3.4.5-alpha.9"})
	check(">= 0.3.0", []string{"0.3.0", "0.3.1", "1.0.0"}, []string{"0.0.0", "0.2.0", "0.2.9", "0.3.0-beta"})
}

func TestNewSemverRequirements_hyphens(t *testing.T) {
	check := _semverReqCheck(t)

	//hyphen ranges
	check("1.2.3 - 2.3.4", []string{"1.2.3", "2.0.0", "2.3.4"}, []string{"1.2.2", "2.0.0-alpha", "2.3.5", "3.0.0", "2.4.0"})
	check("1.2 - 2.3.4", []string{"1.2.0", "1.2.3", "2.0.0", "2.3.4"}, []string{"2.0.0-alpha", "2.3.5", "3.0.0"})
	check("1.2 - 2.3", []string{"1.2.3", "1.2.0", "2.0.0", "2.3.4"}, []string{"2.4.0", "1.1.99", "3.0.0"})
	check("1.2 - 2", []string{"1.2.0", "2.0.0", "2.99.99"}, []string{"3.0.0"})
}

func TestNewSemverRequirements_xranges(t *testing.T) {
	check := _semverReqCheck(t)

	//x-ranges
	check("", []string{"0.0.0", "1.0.0", "1.0.0-alpha"}, []string{})
	check("*", []string{"0.0.0", "1.0.0", "1.0.0-alpha"}, []string{})
	check("1.x", []string{"1.0.0", "1.99.99"}, []string{"0.99.99", "1.2.0-alpha", "2.0.0"})
	check("1.x.x", []string{"1.0.0", "1.99.99"}, []string{"0.99.99", "1.2.0-alpha", "2.0.0"})
	check("1.X.x", []string{"1.0.0", "1.99.99"}, []string{"0.99.99", "1.2.0-alpha", "2.0.0"})
	check("1.2.X", []string{"1.2.0", "1.2.99"}, []string{"0.99.99", "1.2.1-alpha", "2.0.0", "1.3.0"})
	check("1.2.*", []string{"1.2.0", "1.2.99"}, []string{"0.99.99", "1.2.1-alpha", "2.0.0", "1.3.0"})
	check("1", []string{"1.0.0", "1.99.99"}, []string{"0.99.99", "1.2.0-alpha", "2.0.0"})
	check("1.2", []string{"1.2.0", "1.2.99"}, []string{"0.99.99", "1.2.1-alpha", "2.0.0", "1.3.0"})
}

func TestNewSemverRequirements_caret(t *testing.T) {
	check := _semverReqCheck(t)

	//caret ranges
	check("^1.2.3", []string{"1.2.3", "1.2.4", "1.9.9"}, []string{"1.2.4-alpha", "2.0.0"})
	check("^0.2.3", []string{"0.2.3", "0.2.4", "0.2.9"}, []string{"1.2.4-alpha", "0.3.0", "2.0.0"})
	check("^0.0.3", []string{"0.0.3"}, []string{"1.2.4-alpha", "0.3.0", "0.0.3-beta", "0.0.4", "2.0.0"})
	check("^1.2.3-beta.2", []string{"1.2.3", "1.2.3-beta.3", "1.9.9", "1.2.3-beta.2"}, []string{"1.2.3-alpha", "1.2.2", "1.2.4-beta.2", "2.0.0"})
	check("^0.0.3-beta", []string{"0.0.3", "0.0.3-pr.2", "0.0.3-beta"}, []string{"0.0.4", "0.0.3-alpha", "0.0.2"})
	check("^1.2.x", []string{"1.2.0", "1.2.1", "1.99.99"}, []string{"1.1.2", "2.0.0", "1.2.0-beta"})
	check("^0.0.x", []string{"0.0.0", "0.0.1", "0.0.99"}, []string{"1.1.2", "2.0.0", "1.2.0-beta", "0.0.0-beta", "0.1.0"})
	check("^0.0", []string{"0.0.0", "0.0.1", "0.0.99"}, []string{"1.1.2", "2.0.0", "1.2.0-beta", "0.0.0-beta", "0.1.0"})
	check("^1.x", []string{"1.0.0", "1.0.1", "1.99.5"}, []string{"2.0.0", "1.1.0-beta"})
	check("^0.x", []string{"0.0.0", "0.1.0", "0.0.1", "0.1.1"}, []string{"1.0.0", "0.0.1-beta"})
}

func TestNewSemverRequirements_tilde(t *testing.T) {
	check := _semverReqCheck(t)

	//tilde ranges
	check("~1.2.3", []string{"1.2.3", "1.2.4", "1.2.99"}, []string{"1.3.0", "1.2.3-alpha"})
	check("~ 1.2.3", []string{"1.2.3", "1.2.4", "1.2.99"}, []string{"1.3.0", "1.2.3-alpha"})
	check("~1", []string{"1.0.0", "1.99.99"}, []string{"0.99.99", "1.2.0-alpha", "2.0.0"})
	check("~1.2", []string{"1.2.0", "1.2.99"}, []string{"0.99.99", "1.2.1-alpha", "2.0.0", "1.3.0"})
	check("~0.2.3", []string{"0.2.3", "0.2.4"}, []string{"0.2.2", "0.3.0", "0.2.3-alpha"})
	check("~0.2", []string{"0.2.0", "0.2.99"}, []string{"0.1.9", "0.3.0", "0.2.3-alpha"})
	check("~0", []string{"0.0.0", "0.2.4"}, []string{"1.2.2", "1.0.0", "0.2.3-alpha"})
	check("~1.2.3-beta.2", []string{"1.2.3-beta.2", "1.2.3", "1.2.4", "1.2.3-beta.3"}, []string{"1.2.2", "1.0.0", "1.2.4-beta.2", "1.3.0"})
}

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
			t.Errorf("Error got '%s' but expected '%s'\n", sv.String(), ex.String())
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
	check("0.2", "0.2.0")
	check("0", "0.0.0")
	check("1.2.3-beta.2", "1.2.3-beta.2")
	check("1.0.0-rc.1", "1.0.0-rc.1")
	check("1.0.0-rc", "1.0.0-rc")
	check("1.0.0-rc+foo", "1.0.0-rc")
	check("1.0.0+foo", "1.0.0")
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
			t.Errorf("Error, got '%s' but expected '%s'\n", sv.String(), ex.String())
		}
		if r != rounded {
			t.Errorf("Error, rounding reported '%b' but expected '%b'\n", r, rounded)
		}
	}

	check("2.3.4", "2.3.4", false)
	check("2.3", "2.4.0", true)
	check("2", "3.0.0", true)
	check("1.x", "2.0.0", true)
	check("1.2.3", "1.2.3", false)
	check("1.2", "1.3.0", true)
	check("1.x", "2.0.0", true)
	check("1.2.X", "1.3.0", true)
	check("1", "2.0.0", true)
	check("1.*", "2.0.0", true)
	check("1.0.0-rc.1", "1.0.0-rc.1", false)
	check("1.0.0-rc", "1.0.0-rc", false)
	check("1.0.0-rc+foo", "1.0.0-rc", false)
	check("1.0.0+foo", "1.0.0", false)
}
