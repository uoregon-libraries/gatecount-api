# Internal Architecture

This won't make sense if you're not UO, so ignore it.

This project is currently running on libweb and lives under `/usr/local`.  As
it's a Go app, it's highly portable and can be moved to a better server if we
find that necessary.

The credentials are tied to our internal ADI service account, and can be seen
in the "env" file.

This is currently running in a cron job on a cron tied to the root user.  As
this tool doesn't write files, but reads a sensitive file, root makes the most
sense, though an entry in `/etc/crontab` could be used instead if we were
concerned.  Or we could simply wrap it in `su X -c "..."`

The current command is run hourly and looks like more or less like this:

`./gatecount-api --days-ago-start 7 --days-ago-end 0`

We don't need nearly this much data for an hourly run, but it does protect
against short-term server outages, and it takes under two seconds per run.
