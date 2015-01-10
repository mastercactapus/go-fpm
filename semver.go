package main

import (
	"github.com/blang/semver"
	"strconv"
	"strings"
)

type SemverRequirements struct {
	requirements []requirementSet
}

type requirementSet []requirementFn
type requirementFn func(semver.Version) bool

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
	return semver.Parse(strings.Join(parts, "."))

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

// func parseHyphenRange(version string) {

// }

func NewSemverRequirements(requirements string) (*SemverRequirements, error) {
	sr := new(SemverRequirements)

	//cleanup, trim and remove duplicate whitespace
	requirements = strings.TrimSpace(requirements)
	requirements = strings.Replace(requirements, "  ", " ", -1)
	parts := strings.Split(requirements, " ")
	sr.requirements = make(requirementSet, 0, len(parts))
	var currentReq []requirementFn

	for i := range parts {
		if i == 0 || parts[i-1] == "||" {
			currentReq = make([]requirementFn, 0, len(parts)-i)
		}

		switch parts[i][0] {
		case '~':
		case '^':
		default:
			if i < len(parts)-2 && parts[i+1] == "-" { //lower part of a hyphen range
				v, err := parseDown(parts[i])
				if err != nil {
					return nil, err
				}
				currentReq = append(currentReq, v.GTE)
			} else if i > 1 && parts[i-1] == "-" { //higher part of a hyphen range
				v, round, err := parseUp(parts[i])
				if err != nil {
					return nil, err
				}
				if round {
					currentReq = append(currentReq, v.LT)
				} else {
					currentReq = append(currentReq, v.LTE)
				}
			} else { //exact or wildcard version
				vHigh, round, err := parseUp(parts[i])
				if err != nil {
					return nil, err
				}
				if round {
					currentReq = append(currentReq, vHigh.LT)
				} else { //if not rounded, then we have an exact version
					currentReq = append(currentReq, vHigh.EQ)
					continue
				}

				vLow, err := parseDown(parts[i])
				if err != nil {
					return nil, err
				}
				currentReq = append(currentReq, v.GTE)
			}
		}

		if i == len(parts)-1 || parts[i+1] == "||" {
			sr.requirements = append(sr.requirements, currentReq)
		}
	}

	requirements = strings.Replace(requirements, " || ", "||", -1)
	requirements = strings.Replace(requirements, " - ", "--", -1)

	ORs := strings.Split(requirements, "||")
	for _, orv := range ORs {
		ANDs := strings.Split(orv, " ")
		for _, andv := range ANDs {

		}
	}

	//split '||' sets (OR)
	//split ' - ' ranges
	//split by whitespace (AND)
	//parse hyphen ranges
	//parse X-Ranges
	//parse Tilde Ranges
	//parse Caret Ranges
	return nil, nil
}

func (s *SemverRequirements) Test(sv semver.Version) bool {
	for _, v := range s.requirements {
		if v.test(sv) {
			return true
		}
	}
	return false
}

func (r requirementSet) test(sv semver.Version) bool {
	for _, v := range r {
		if !v(sv) {
			return false
		}
	}
	return true
}
