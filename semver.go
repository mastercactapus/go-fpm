package main

import (
	"github.com/blang/semver"
	"strconv"
	"strings"
)

const (
	svLT = iota
	svLTE
	svGT
	svGTE
	svEQ
)

type requirement struct {
	version semver.Version
	svType  int
}

type SemverRequirements struct {
	requirements [][]requirement
}

//strips any prefix for the version for parsing
func stripPrefix(version string) string {
	switch version[0] {
	case '>', '<':
		if version[1] == '=' {
			version = version[2:]
		} else {
			version = version[1:]
		}
	case '=', '^', '~':
		version = version[1:]
	}

	if version[0] == 'v' {
		version = version[1:]
	}

	return version
}

//parses, replacing missing/wildcards with zero
func parseDown(version string) (sv semver.Version, err error) {
	if len(version) == 0 {
		//return version 0.0.0
		return
	}
	parts := strings.SplitN(version, ".", 3)
	if len(parts) == 1 {
		if parts[0] == "x" || parts[0] == "X" || parts[0] == "*" {
			parts[0] = "0"
		}
		return semver.Parse(parts[0] + ".0.0")
	}
	if len(parts) == 2 {
		if parts[1] == "x" || parts[1] == "X" || parts[1] == "*" {
			parts[1] = "0"
		}
		return semver.Parse(parts[0] + "." + parts[1] + ".0")
	}
	if parts[2] == "x" || parts[2] == "X" || parts[2] == "*" {
		parts[2] = "0"
	}
	version = strings.Join(parts, ".")
	buildIndex := strings.IndexRune(version, '+')
	if buildIndex != -1 {
		version = version[:buildIndex]
	}
	return semver.Parse(version)
}

//parses, replacing missing/wildcards by incrementing the higher version
func parseUp(version string) (sv semver.Version, round bool, err error) {
	parts := strings.SplitN(version, ".", 3)
	var major, minor int64

	major, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}
	if len(parts) == 1 || parts[1] == "x" || parts[1] == "X" || parts[1] == "*" {
		major++
		round = true
		sv, err = semver.Parse(strconv.Itoa(int(major)) + ".0.0")
		return
	}

	minor, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}
	if len(parts) == 2 || parts[2] == "x" || parts[2] == "X" || parts[2] == "*" {
		minor++
		round = true
		sv, err = semver.Parse(strconv.Itoa(int(major)) + "." + strconv.Itoa(int(minor)) + ".0")
		return
	}

	sv, err = semver.Parse(version)
	return
}

//checks, if a prerelease, that the major,minor,patch tuple match
func validPrerelease(orig semver.Version, sv semver.Version) bool {
	if sv.Pre == nil {
		return true
	}
	if orig.Pre != nil {
		return false
	}
	return orig.Major == sv.Major && orig.Minor == sv.Minor && orig.Patch == sv.Patch
}

func checkVersion(req requirement, sv semver.Version) bool {
	switch req.svType {
	case svLT:
		return sv.LT(req.version)
	case svLTE:
		return sv.LTE(req.version)
	case svGT:
		return sv.GT(req.version)
	case svGTE:
		return sv.GTE(req.version)
	case svEQ:
		return sv.EQ(req.version)
	default:
		panic("Unknown comparator")
	}
}

func NewSemverRequirements(requirements string) (*SemverRequirements, error) {
	sr := new(SemverRequirements)

	//cleanup, trim and remove duplicate whitespace
	requirements = strings.TrimSpace(requirements)
	parts := strings.Split(requirements, " ")
	sr.requirements = make([][]requirement, 0, len(parts))
	var currentSet []requirement
	for i := range parts {
		if len(parts[i]) == 0 || parts[i] == "||" || parts[i] == "-" || parts[i] == "*" {
			continue
		}
		if i == 0 || parts[i-1] == "||" {
			currentSet = make([]requirement, 0, len(parts)-i)
		}
		clean := stripPrefix(parts[i])
		vLow, err := parseDown(clean)
		if err != nil {
			return nil, err
		}
		vHigh, round, err := parseUp(clean)
		if err != nil {
			return nil, err
		}
		var highComparator int
		if round {
			highComparator = svLT
		} else {
			highComparator = svLTE
		}

		if i < len(parts)-1 && parts[i+1] == "-" {
			currentSet = append(currentSet, requirement{vLow, svGTE})
		} else if i > 1 && parts[i-1] == "-" {
			currentSet = append(currentSet, requirement{vHigh, highComparator})
		} else {
			switch parts[i][0] {
			case '^':
			case '~':
			case '<':
				if parts[i][1] == '=' {
					currentSet = append(currentSet, requirement{vLow, svLTE})
				} else {
					currentSet = append(currentSet, requirement{vLow, svLT})
				}
			case '>':
				if parts[i][1] == '=' {
					currentSet = append(currentSet, requirement{vLow, svGTE})
				} else {
					currentSet = append(currentSet, requirement{vLow, svGT})
				}
			default:
				currentSet = append(currentSet, requirement{vLow, svGTE})
				currentSet = append(currentSet, requirement{vHigh, highComparator})
			}
		}

		if i == len(parts)-1 || parts[i+1] == "||" {
			sr.requirements = append(sr.requirements, currentSet)
		}
	}

	return sr, nil
}

func (s *SemverRequirements) SatisfiedBy(sv semver.Version) bool {
	if len(s.requirements) == 0 {
		return true
	}
	for _, reqs := range s.requirements {
		valid := true
		for _, v := range reqs {
			if !validPrerelease(v.version, sv) {
				valid = false
				break
			}
			if !checkVersion(v, sv) {
				valid = false
				break
			}
		}
		if valid {
			return true
		}
	}
	return false
}
