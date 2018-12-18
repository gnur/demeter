package lib

import (
	"fmt"
	"time"
)

// SearchResult contains the results from a calibre query
type SearchResult struct {
	TotalNum  int    `json:"total_num"`
	Offset    int    `json:"offset"`
	Sort      string `json:"sort"`
	LibraryID string `json:"library_id"`
	SortOrder string `json:"sort_order"`
	Vl        string `json:"vl"`
	BaseURL   string `json:"base_url"`
	Num       int    `json:"num"`
	BookIds   []int  `json:"book_ids"`
	Query     string `json:"query"`
}

// CalibreBook is a book from calibre
type CalibreBook struct {
	UUID          string            `json:"uuid"`
	Title         string            `json:"title"`
	ApplicationID int               `json:"application_id"`
	TitleSort     string            `json:"title_sort"`
	Cover         string            `json:"cover"`
	Pubdate       string            `json:"pubdate"`
	MainFormat    map[string]string `json:"main_format"`
	AuthorSort    string            `json:"author_sort"`
	Authors       []string          `json:"authors"`
	Timestamp     time.Time         `json:"timestamp"`
	Languages     []string          `json:"languages"`
	LastModified  time.Time         `json:"last_modified"`
	Thumbnail     string            `json:"thumbnail"`
	Formats       []string          `json:"formats"`
}

// BooksQueryResult is
type BooksQueryResult map[string]CalibreBook

// Host describes all attributes related to a host
type Host struct {
	ID            int `storm:"id,increment"`
	URL           string
	Downloads     int
	Scrapes       int
	LastScrape    time.Time
	LastDownload  time.Time
	Added         time.Time
	ScrapeResults []ScrapeResult
	Active        bool
}

// ScrapeResult is the result of a single scrape attempt
type ScrapeResult struct {
	Start      time.Time
	End        time.Time
	Success    bool
	Results    int
	Downloads  int
	Considered int
}

// Book is a oversimplified representation of a book
type Book struct {
	ID       int `storm:"id,increment"`
	Added    time.Time
	Hash     string `storm:"unique"`
	SourceID int    `storm:"index"`
	Author   string
	Title    string
}

// Print prints a host in a nicely formatted way
func (h *Host) Print() {
	fmt.Printf(`ID: %d
URL: %s
Scrapes: %d
Downloads: %d
Active: %b
Last scrape: `, h.ID, h.URL, h.Scrapes, h.Downloads, h.Active)
	if h.Scrapes == 0 {
		fmt.Println("never")
	} else {
		fmt.Println(h.LastScrape)
	}
}
