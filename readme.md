# demeter

Demeter is a tool for sucking all the epubs you don't have yet from a calibre library. It does this by internally building a list of books it has seen based on some clever algorithms. At least, that's the idea.

It will only allow a scrape of a host every 12 hours to prevent hammering a host.

# installation and usage

download the correct demeter binary for your platform from the [releases](https://github.com/gnur/demeter/releases) page

This is a standalone binary, there is no need to install any dependencies.

move it somewhere in your \$PATH so you can call it with `demeter`

## add a host

`demeter host add http://example.com:8080`

## scrape all hosts and store results in the directory ./books

`demeter scrape run -d books`

For the rest, use the built in help.

This tool should be used for whatever you want, enjoy.

# database

Demeter builds an internal database that is stored in ~/.demeter/demeter.db

# scraping

When scraping a host, demeter does the following:

- use the API to collect all book ids
- check if there a new book ids since the previous scrape
- use the API to get the details for all the new book ids
- check the internal db if a book has already been downloaded
- download the book if it isn't and add it to the internal db
- mark the host as scraped so it won't do it again within 12 hours
- if the host failed, mark it as failed and disable it after a while

# all commands

```
$ demeter -h
demeter is CLI application for scraping calibre hosts and
retrieving books in epub format that are not in your local library.

Usage:
  demeter [command]

Available Commands:
  dl          download related commands
  help        Help about any command
  host        all host related commands
  scrape      all scrape related commands

$ demeter dl -h
download related commands

Usage:
  demeter dl [command]

Aliases:
  dl, download, downloads, dls

Available Commands:
  add          add a number of hashes to the database
  deleterecent delete all downloads from this time period
  list         list all downloads

$ demeter host -h
all host related commands

Usage:
  demeter host [command]

Available Commands:
  add         add a host to the scrape list
  disable     disable a host
  enabled     make a host active
  list        list all hosts
  rm          delete a host
  stats       Get host stats

$ demeter scrape -h
all scrape related commands

Usage:
  demeter scrape [command]

Available Commands:
  run         run all scrape jobs

$ demeter scrape run -h
demeter scrape run -h
Go over all defined hosts and if the last scrape
is old enough it will scrape that host.

Usage:
  demeter scrape run [flags]

Flags:
  -h, --help               help for run
  -d, --outputdir string   path to downloaded books to (default "books")
  -n, --stepsize int       number of books to request per query (default 50)
  -u, --useragent string   user agent used to identify to calibre hosts (default "demeter / v1")
  -w, --workers int        number of workers to concurrently download books (default 10)
```
