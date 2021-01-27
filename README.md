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

Process:

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

### All together, now!

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

### Your first run

The first time you run this you probably want to import a *lot* of historic data.  We did it this way:

```bash
make
./bin/gatecount-api --verbose --days-ago-start 3660 --days-ago-end 1 2>log
```

Piping the STDERR output to a file allowed us to see exactly what was being
read from Traf-Sys and how the posts to LibInsight were goign.

### Scheduling

Your command should be scheduled to overlap its previous runs (e.g., run daily
if pulling a week of data) if possible, so that a missed run doesn't mean lost
counts.

Alternatively, you could run daily, only pulling yesterday's data, but having a
weekly "catch everything" run that pulls data for the last two weeks or
something.  That way you're certain that even if a single run fails, you'll get
the data eventually.

Note: posting the same data to LibInsight multiple times, though undesireable,
has no negative effects on the dataset's information, and is far better than
losing data.

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

## Troubleshooting

LibInsight's API is flaky.  The app had to have a lot more error handling than
I'd have preferred, and it's still not really great.  LibInsight just doesn't
tell you much when things fail, unless the failure is painfully obvious (like
forgetting a `gate_id`).  You can't even rely on a valid HTTP status code
having any extra meaning.

### All retries fail

One problem I'm noticing is that when there are batches posted to LibInsight
which aren't big, you just get an error and that's that.  Your data never goes
to LibInsight.  This is "fixed" by splitting batches in a way that ensures
they're always at least 50 records, but there's still always the chance of
getting a four-item batch due to circumstances outside our control.

One way to try and handle this is to gather *too much* data.  Instead of, for
instance, pulling just yesterday and today, you could pull everything over the
course of the last week.  Then the total number of counts is likely to be big
enough to be split nicely into large-ish batches.

## Port?

If you find the amount of simplicity here disturbing, I apologize.  I'll port
this to PHP, Ruby, and nodejs when I have time, and add gobs of unnecessary
dependencies and configuration.  Then I'll create setup documentation that only
work on a single version of a single OS with a single version of PHP/Ruby/node.
I'll also add `we haven't tested this on Windows, sorry ¯\_(ツ)_/¯` to those
instructions.  And of course I'll add tons of other miscellaneous out-of-date
user documentation.

(yes, that was sarcasm; I'm never porting this)

Actually docs getting out of date probably *will* happen.  Dang it, I shoulda
quit while I was ahead.
