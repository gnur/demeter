# demeter

Demeter is a tool for sucking all the epubs you don't have yet from a calibre library. It does this by internally building a list of books it has seen based on some clever algorithms. At least, that's the idea.  

It will only allow a scrape of a host every 24 hours to prevent hammering a host.

# usage

download the correct demeter binary for your platform from the releases page

move it somewhere in your $PATH so you can call it with `demeter`

## add a host

`demeter host add http://example.com:8080`

## scrape all hosts and store results in books
`demeter scrape runv2 -d books`

There is also a `demeter scrape run` command which might also work but probably runv2 is a bit better.

For the rest, use the built in help.

This tool should be used for whatever you want, enjoy.

# database
Demeter builds an internal database that is stored in ~/.demeter/demeter.db

# scraping

When scraping a host, demeter does the following:

- use the API to collect all book ids
- use the API to get the details for all the book ids
- check the internal db if a book is already downloaded
- download the book if it isn't and add it to the internal db
- mark the host as scraped so it won't do it again within 24 hours
- if the host failed, mark it as failed and disable it after a while
