# gatecount-api

This project has a single command which reads data counts from a Traf-Sys API
endpoint and sends translated per-site counts into a LibInsight dataset.  It's
very simple.

## Setup / Configuration

Setup is pretty simple.  TrafSys just has an API to get gate counts, and relies
on a username and password to authenticate.  LibInsight requires first
createing a Gate Count dataset, then adding libraries to it which are
configured as Unidirectional.  When using the LibInsight API, the rollover
isn't meaningful, so we set it to 1000000.

Configuration is also very simple.  You can manually set environment variables
the app requires before running it, and it will tell you if you missed any.
You can also opt to put them in a file and source them.  The included
`env-example` file includes all environment variables necessary for both
projects to function.  Simply copy `env-example` to `env` and fill in the data
as needed.

## Build

Building the app is relatively simple.  Dependencies:

- A supported 1.x [Go compiler](https://golang.org/dl/)
- Make

- Run `make`
- If you really hate Make, you can just read the `Makefile` and execute the `go
  build` instructions manually, e.g., `go build -o bin/gatecount-api
  ./cmd/gatecount-api`.

Note also that you don't need Go on the production system that runs this
application.  So long as you compile it on a system with the same architecture,
or cross-compile it targeting the production system's architecture.  e.g., if
you're on 64-bit Ubuntu Linux but want an executable that will work on a 32-bit
x86-based Windows:

    GOARCH=386 GOOS=windows go build -o bin/gatecount-api.exe ./cmd/gatecount-api

All valid `GOOS` and `GOARCH` values can be found in Go's documentation:
https://golang.org/doc/install/source#environment

## Run

Putting it all together is very simple:

```bash
cp env-example env
vim env
source ./env

make
./bin/gatecount-api --days-ago-start 7 --days-ago-end 1
```

This pulls the last 7 days of hourly "In" gate counts from trafsys, up to
and including yesterday, and puts new entries into an sqlite database.  Only
perimeter sensors are read, as internal sensors aren't useful for getting foot
traffic.

The app then aggregates counts by ignoring location (e.g., "Front Entrance"),
converts `SiteCode` to a LibInsight `gate_id`, and sends them to the LibInsight
API endpoint configured via the environment.

This command should be scheduled to overlap (e.g., run daily if pulling a week
of data) if possible, so that a missed run doesn't mean lost counts.  Posting
the same data to LibInsight multiple times, though undesireable, has no
negative effects on the dataset's information, and is far better than losing
data.

## Use outside of UO

This project could *almost* be reusable for others.  Almost.

The one real problem is the hard-coded "translation" table.  In
`cmd/gatecount-api/count.go`, we've hard-coded the mapping from TrafSys
`SiteCode` to LibInsight `gate_id`.  This could be extracted to some kind of
environment var or command-line flag, but that's a "maybe someday" kind of
task.

It would be trivial for somebody to change that code, of course, but then it's
not really reuse so much as a repurposing.

PRs are welcomed!

## Port?

If you find the amount of simplicity here disturbing, I apologize.  I'll port
this to PHP or Ruby when I have time, and add gobs of unnecessary dependencies
and configuration.  Then I'll create setup documentation that only work on a
single version of a single OS with a single version of PHP/Ruby.  And of course
I'll add out-of-date user documentation.

Actually that last piece probably *will* happen.  Dang it, I shoulda quit while
I was ahead.

(yes, that was sarcasm)
