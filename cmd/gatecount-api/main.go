package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/uoregon-libraries/gopkg/logger"
)

var l *logger.Logger
var trafurl = "https://portal.trafnet.com/rest"

const magicDelayStringOMFG = "--delay--"

func usage(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	fmt.Fprintln(os.Stderr)
	flag.Usage()
	os.Exit(1)
}

func main() {
	var startDaysAgo, endDaysAgo int
	var verbose bool
	flag.IntVar(&startDaysAgo, "days-ago-start", 0, "number of days ago to start TrafSys counts, e.g., 7 would mean the first day included is a week ago")
	flag.IntVar(&endDaysAgo, "days-ago-end", 1, "number of days ago to end TrafSys counts, e.g., 2 would include from start through the day before yesterday, while 0 would gather all available data from start date through today")
	flag.BoolVar(&verbose, "verbose", false, "show lots of really painfully unnecessary logging")

	flag.Parse()
	if verbose {
		l = logger.Named("gatecount-api", logger.Debug, false)
	} else {
		l = logger.Named("gatecount-api", logger.Info, false)
	}
	if startDaysAgo < 1 {
		usage("--days-ago-start must be at least 1")
	}
	if endDaysAgo < 0 {
		usage("--days-ago-end must be at least 0")
	}
	if startDaysAgo < endDaysAgo {
		usage("--days-ago-start must be greater than --days-ago-end")
	}

	var user = os.Getenv("TRAFSYS_USER")
	var pass = os.Getenv("TRAFSYS_PASS")
	if user == "" || pass == "" {
		usage("TRAFSYS_USER and TRAFSYS_PASS must both be set")
	}

	var libinsightURL = os.Getenv("LIBINSIGHT_URL")
	if libinsightURL == "" {
		usage("LIBINSIGHT_URL must be set based on the LibInsight admin API code (combining host and the POST path)")
	}

	// Pull all traffic data from TrafSys
	var token, err = getToken(trafurl, user, pass)
	if err != nil {
		l.Fatalf("Could not get bearer token from Traf-Sys: %s", err)
	}

	var counts []*trafficCount
	counts, err = getTraffic(trafurl, token, startDaysAgo, endDaysAgo)
	if err != nil {
		l.Fatalf("Could not read traffic data for from Traf-Sys: %s", err)
	}

	// Aggregate data per site, ignoring locations
	var aggregated = make(map[string]*trafficCount)
	for _, count := range counts {
		var k = count.key()
		if aggregated[k] == nil {
			var clone = &trafficCount{
				When:   count.When,
				GateID: count.GateID,
			}
			aggregated[k] = clone
		}
		aggregated[k].Ins += count.Ins
		aggregated[k].Outs += count.Outs
	}

	postCounts(libinsightURL, aggregated)
}
