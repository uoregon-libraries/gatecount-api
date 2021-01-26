package main

import (
	"fmt"
	"strconv"
	"time"
	"unicode"
)

type trafficCount struct {
	SiteCode     string
	Location     string
	IsInteral    bool
	PeriodEnding string
	Ins, Outs    int

	When   time.Time `json:"-"`
	GateID int       `json:"-"`
}

// machine turns a string into a machine-friendly value that contains only
// valid letters, digits, or underscores.  Anything else is converted to an
// underscore.
func machine(s string) string {
	var out = []rune(s)
	for i, r := range out {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			out[i] = '_'
		}
	}

	return string(out)
}

// key gives us a machine-friendly way of describing an aggregated count
// instance (site and hour, with location ignored)
func (c *trafficCount) key() string {
	return fmt.Sprintf("%s/%s", machine(c.SiteCode), c.When.Format("2006-01-02T15"))
}

// String shows a human-friendly depiction of a trafficCount
func (c *trafficCount) String() string {
	var where = c.SiteCode
	var when = c.PeriodEnding
	if c.GateID > 0 {
		where = strconv.Itoa(c.GateID)
	}
	if !c.When.IsZero() {
		when = c.When.Format("2006-01-02T15")
	}
	return fmt.Sprintf("&trafficCount{Site / Gate: %s, Location: %q, When: %s, In/Out: %d/%d}", where, c.Location, when, c.Ins, c.Outs)
}

// postProcess converts TrafSys values into usable data for the LibInsight API
// by converting from the TrafSys time format to a Go time, and translating
// TrafSys SiteCode to a gate_id.
func (c *trafficCount) postProcess() error {
	var err error
	c.When, err = time.Parse("2006-01-02T15:04:05", c.PeriodEnding)
	if err != nil {
		return err
	}

	// TrafSys SiteCode values:
	// - 02 Knight
	// - 03 PSC
	// - 04 Design Library
	// - 05 Law Library
	// - 06 Math Library
	// - 07 PDX White Stag Library
	// - 08 Oregon Marine Biology

	// LibInsight gate_id values:
	// - 3 for "Knight"
	// - 4 for "PSC"
	// - 5 for "Design Library"
	// - 6 for "Law Library"
	// - 7 for "Math Library"
	// - 8 for "PDX White Stag Library"
	// - 9 for "Oregon Marine Biology"

	var siteMap = map[string]int{
		"02": 3,
		"03": 4,
		"04": 5,
		"05": 6,
		"06": 7,
		"07": 8,
		"08": 9,
	}

	c.GateID = siteMap[c.SiteCode]
	if c.GateID == 0 {
		return fmt.Errorf("no mapping for SiteCode %q", c.SiteCode)
	}

	return nil
}
