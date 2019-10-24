package lib

import (
	"net/url"
	"time"
)

// App holds all the config for the V2 demeter type
type App struct {
	UserAgent       string
	Timeout         time.Duration
	DownloadTimeout time.Duration
	WorkerInterval  time.Duration
	StepSize        int
	OutputDir       string
	Queues          WorkerQueues
}

// GetIDSRequest holds the information to retrieve the book ids from a calibre host
type GetIDSRequest struct {
	Num    int
	Offset int
	U      url.URL
	Resp   chan GetIDSResponse
}

// GetIDSResponse is the response for IDS request
type GetIDSResponse struct {
	IDs []int
	Err error
}

// GetBooksRequest converts IDs into books
type GetBooksRequest struct {
	IDs  []int
	U    url.URL
	Resp chan GetBooksResponse
}

// GetBooksResponse is the response
type GetBooksResponse struct {
	Books BooksQueryResult
	Err   error
}

// DownloadBookRequest holds the request to DL a book
type DownloadBookRequest struct {
	URL  string
	Path string
	Resp chan DownloadBookResponse
}

// DownloadBookResponse holds the result of a book dl
type DownloadBookResponse struct {
	Err error
}
