package lib

import (
	"net/url"
)

func (a *App) getIDSAsync(u url.URL, offset int, num int) ([]int, error) {
	ch := make(chan GetIDSResponse)
	a.Queues.IDS <- GetIDSRequest{
		Num:    num,
		Offset: offset,
		U:      u,
		Resp:   ch,
	}
	resp := <-ch
	close(ch)
	return resp.IDs, resp.Err
}

func (a *App) getBooksAsync(u url.URL, ids []int) (BooksQueryResult, error) {
	ch := make(chan GetBooksResponse)
	a.Queues.Books <- GetBooksRequest{
		IDs:  ids,
		U:    u,
		Resp: ch,
	}
	resp := <-ch
	close(ch)
	return resp.Books, resp.Err
}

func (a *App) downloadBookAsync(url string, path string) error {
	ch := make(chan DownloadBookResponse)
	a.Queues.DlBook <- DownloadBookRequest{
		URL:  url,
		Path: path,
		Resp: ch,
	}
	resp := <-ch
	close(ch)
	return resp.Err
}
