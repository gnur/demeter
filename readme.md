# demeter

demeter is a tool for downloading the .epub files you don't have from a Calibre library. It does this by building a database of books it has seen based on some clever algorithms. At least, that's the idea.

(Demeter only allows scraping a host every 12 hours to prevent overloading the server.)

# Installation and Usage

Download the appropriate demeter binary for your platform from the [releases](https://github.com/gnur/demeter/releases) page.

This is a standalone binary, there's no need to install any dependencies.

Move it somewhere in your \$PATH so you can call it with `demeter`

## Add a Host

`demeter host add http://example.com:8080`

## Scrape all hosts and store results in the directory ./books

`demeter scrape run -d books`

For the rest, use the built in help.

This tool can be used for whatever you want, enjoy.

# Database

Demeter builds an internal database that is stored in ~/.demeter/demeter.db

# Scraping

When scraping a host, demeter does the following:

- Use the API to collect all book ids
- Check if there a new book ids since the previous scrape
- Use the API to get the details for all the new book ids
- Check the internal db if a book has already been downloaded
- Download the book if it isn't and add it to the internal db
- Mark the host as scraped so it won't do it again within 12 hours
- If the host failed, mark it as failed and disable it after a while
