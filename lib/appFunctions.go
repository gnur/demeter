package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func (a *App) getBody(u string, v interface{}) error {
	c := http.Client{
		Timeout: a.Timeout,
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", a.UserAgent)

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) getIDS(u url.URL, offset int, num int) ([]int, error) {
	v := url.Values{}
	v.Set("num", strconv.Itoa(num))
	v.Set("offset", strconv.Itoa(offset))

	u.RawQuery = v.Encode()

	r := SearchResult{}
	err := a.getBody(u.String(), &r)
	if err != nil {
		return nil, err
	}
	return r.BookIds, nil
}

func (a *App) getBooks(u url.URL, ids []int) (BooksQueryResult, error) {
	u.Path = "/ajax/books"
	v := url.Values{}
	v.Set("ids", intSliceToString(ids))
	u.RawQuery = v.Encode()

	r := BooksQueryResult{}
	err := a.getBody(u.String(), &r)
	return r, err
}

func (a *App) downloadBook(url string, path string) error {
	c := http.Client{
		Timeout: a.DownloadTimeout,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", a.UserAgent)

	response, err := c.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("Got %d statuscode", response.StatusCode)

	}
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) getAllIDS(u url.URL) ([]int, error) {

	ids := []int{}
	u.Path = "/ajax/search"

	v := url.Values{}
	v.Set("num", "0")
	u.RawQuery = v.Encode()

	r := SearchResult{}
	err := a.getBody(u.String(), &r)
	if err != nil {
		return ids, err
	}

	for i := 0; i < r.TotalNum; i += a.StepSize {
		stepIDs, err := a.getIDSAsync(u, i, a.StepSize)

		if err != nil {
			continue
		}
		ids = append(ids, stepIDs...)

	}
	return ids, nil
}
