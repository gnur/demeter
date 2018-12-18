package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anonhoarder/demeter/db"
	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
)

type scrapeConfig struct {
	sync.Mutex
	u           *url.URL
	userAgent   string
	maxAge      time.Time
	timeout     time.Duration
	backoffTime time.Duration
	outputDir   string
	workers     int
	stepSize    int
	hostID      int
	logger      *log.Entry
}

func (s *scrapeConfig) getBody(u string, v interface{}) error {
	c := http.Client{
		Timeout: s.timeout,
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", s.userAgent)

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

// Scrape performs the actual scrape
func (h *Host) Scrape(workers, stepSize int, userAgent, outputDir string) (*ScrapeResult, error) {
	parsed, err := url.Parse(h.URL)
	if err != nil {
		return nil, err
	}
	parsed.Path = ""

	s := scrapeConfig{
		userAgent: userAgent,
		outputDir: outputDir,
		stepSize:  stepSize,
		workers:   workers,
		maxAge:    h.LastScrape,
		hostID:    h.ID,
		timeout:   10 * time.Second,
		logger:    log.WithField("host", parsed.Hostname()),
	}
	s.u = parsed

	r := ScrapeResult{
		Start:   time.Now(),
		Success: false,
	}
	defer func() {
		r.End = time.Now()
	}()

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return &r, err
	}

	ids, err := s.getIDS()
	if err != nil {
		return &r, err
	}
	s.logger.WithField("total", len(ids)).Info("found books")

	// ##########################################################

	dlChan := make(chan dlRequest)
	doneChan := make(chan int)
	for w := 1; w <= workers; w++ {
		go s.bookDLWorker(w, dlChan, doneChan)
	}

	i := 0
	for i < len(ids) {
		max := i + stepSize
		if max > len(ids) {
			max = len(ids)
		}
		s.logger.WithFields(log.Fields{
			"min":   i,
			"max":   max,
			"total": len(ids),
		}).Info("retrieving books")
		err = s.getBooks(parsed.String(), ids[i:max], dlChan)
		if err != nil {
			s.logger.WithField("err", err).Error("could not get books")
			s.slowDown()
		}
		i += stepSize

	}
	close(dlChan)

	downloaded := 0

	for a := 1; a <= workers; a++ {
		workerDownloads := <-doneChan
		downloaded += workerDownloads
	}
	processTime := time.Since(r.Start)
	log.WithFields(log.Fields{
		"host":       parsed.Host,
		"downloaded": downloaded,
		"duration":   processTime.String(),
		"total":      len(ids),
	}).Info("done with host")

	// ######################################################
	r.Results = len(ids)
	r.Downloads = downloaded
	r.Success = true

	return &r, nil

}

func (s *scrapeConfig) getIDS() ([]int, error) {

	ids := []int{}
	s.u.Path = "/ajax/search"

	v := url.Values{}
	v.Set("num", "0")
	s.u.RawQuery = v.Encode()

	r := SearchResult{}
	err := s.getBody(s.u.String(), &r)
	if err != nil {
		return ids, err
	}

	total := r.TotalNum
	step := strconv.Itoa(s.stepSize)
	for i := 0; i < total; i += s.stepSize {

		v := url.Values{}
		v.Set("num", step)
		v.Set("offset", strconv.Itoa(i))
		s.u.RawQuery = v.Encode()

		s.logger.WithFields(log.Fields{
			"min":   i,
			"max":   i + s.stepSize,
			"total": total,
			"url":   s.u.String(),
		}).Info("indexing ids")
		r := SearchResult{}
		err := s.getBody(s.u.String(), &r)
		if err != nil {
			s.logger.WithField("err", err).Error("could not get body")
			s.slowDown()
			continue
		}
		ids = append(ids, r.BookIds...)

	}
	return ids, nil
}

func intSliceToString(a []int) string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}

	return strings.Join(b, ",")
}

type dlRequest struct {
	url  string
	book *CalibreBook
}

func (s *scrapeConfig) bookDLWorker(id int, dlChan chan dlRequest, doneChan chan int) {
	counter := 0
	for u := range dlChan {
		counter++
		output := fmt.Sprintf("worker_%02d_download_%05d.epub", id, counter)
		output = path.Join(s.outputDir, output)
		s.logger.WithFields(log.Fields{
			"output": output,
			"url":    u,
		}).Info("Downloading file")

		response, err := http.Get(u.url)
		if err != nil {
			s.logger.WithField("err", err).Error("could not download book")
			s.slowDown()
			continue
		}
		file, err := os.Create(output)
		if err != nil {
			s.logger.WithField("err", err).Error("could not open output file")
			continue
		}
		_, err = io.Copy(file, response.Body)
		if err != nil {
			s.logger.WithField("err", err).Error("could not write file with body")
			continue
		}
		file.Close()
		response.Body.Close()
		s.markBookAsDownloaded(u.book)
	}
	doneChan <- counter
}

func (s *scrapeConfig) getBooks(rawURL string, ids []int, dlChan chan dlRequest) error {
	u, _ := url.Parse(rawURL)
	u.Path = "/ajax/books"
	v := url.Values{}
	v.Set("ids", intSliceToString(ids))
	u.RawQuery = v.Encode()
	log.WithField("url", u.String()).Debug("retrieving url")

	r := BooksQueryResult{}
	err := s.getBody(u.String(), &r)
	if err != nil {
		s.slowDown()
		return err
	}
	log.WithField("results", len(r)).Debug("found books")
	u.RawQuery = ""
	for _, b := range r {
		s.logger.WithFields(log.Fields{
			"date":   b.Timestamp,
			"title":  b.Title,
			"author": b.Authors[0],
		}).Debug("checking book")
		if s.bookTooOld(&b) {
			log.WithFields(log.Fields{
				"title":  b.Title,
				"date":   b.Timestamp,
				"author": b.Authors[0],
			}).Debug("book too old")
			continue
		}
		if !bookNotInDatabase(&b) {
			log.WithFields(log.Fields{
				"title":  b.Title,
				"date":   b.Timestamp,
				"author": b.Authors[0],
			}).Debug("book already found in database")
			continue
		}

		//TODO: replace with something like isBookInDB
		//if isBookInBooksing(&b) {
		//	k
		//}

		if epubPath, ok := b.MainFormat["epub"]; ok {
			// can get book
			u.Path = epubPath
			log.WithFields(log.Fields{
				"title":  b.Title,
				"author": b.Authors[0],
				"url":    u.String(),
			}).Info("Downloading book")
			dlChan <- dlRequest{
				url:  u.String(),
				book: &b,
			}
		}

	}

	return nil
}

func (s *scrapeConfig) bookTooOld(b *CalibreBook) bool {
	return b.Timestamp.Before(s.maxAge)
}

func (s *scrapeConfig) slowDown() {
	s.Lock()
	s.stepSize = s.stepSize / 2
	if s.stepSize < 15 {
		s.stepSize = 20
	}
	s.backoffTime = (s.backoffTime / 3) * 2
	if s.backoffTime > 30*time.Second {
		s.backoffTime = 30 * time.Second
	}
	s.logger.WithField("duration", s.backoffTime.String()).Warning("slowing down")
	time.Sleep(s.backoffTime)

	s.Unlock()
}

func bookNotInDatabase(b *CalibreBook) bool {
	hash := hashBook(b.Authors[0], b.Title)
	var book Book
	err := db.Conn.One("Hash", hash, &book)
	return err == storm.ErrNotFound
}

func (s *scrapeConfig) markBookAsDownloaded(b *CalibreBook) error {

	var book Book
	book.Added = time.Now()
	book.Author = b.Authors[0]
	book.Title = b.Title
	book.Hash = hashBook(b.Authors[0], b.Title)
	book.SourceID = s.hostID

	return db.Conn.Save(&book)
}
