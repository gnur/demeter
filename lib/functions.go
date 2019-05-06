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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anonhoarder/demeter/db"
	log "github.com/sirupsen/logrus"
)

var yearRemove = regexp.MustCompile(`\((1|2)[0-9]{3}\)`)
var drukRemove = regexp.MustCompile(`(?i)/ druk [0-9]+`)

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
	failures    int
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
		userAgent:   userAgent,
		outputDir:   outputDir,
		stepSize:    stepSize,
		workers:     workers,
		maxAge:      h.LastScrape,
		hostID:      h.ID,
		timeout:     2 * time.Second,
		backoffTime: 1 * time.Second,
		logger:      log.WithField("host", parsed.Hostname()),
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
	earlyExit := false
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
			exit := s.slowDown()
			if exit {
				earlyExit = true
				s.logger.WithField("failures", s.failures).Error("shutting down early because of failures")
				break
			}
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
	if !earlyExit {
		r.Success = true
	}

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
	hash string
}

func (s *scrapeConfig) bookDLWorker(id int, dlChan chan dlRequest, doneChan chan int) {
	counter := 0
	earlyExit := false
	l := s.logger.WithField("worker", fmt.Sprintf("worker_%02d", id))
	for u := range dlChan {
		counter++
		if earlyExit {
			l.WithFields(log.Fields{
				"counter": counter,
				"url":     u.url,
				"hash":    u.hash,
			}).Warning("not downloading because of an early exit")
			continue
		}
		output := fmt.Sprintf("worker_%02d_download_%05d_%s.epub", id, counter, u.hash)
		output = path.Join(s.outputDir, output)
		l.WithFields(log.Fields{
			"output":  output,
			"counter": counter,
			"url":     u.url,
			"hash":    u.hash,
		}).Info("Downloading file")

		timeout := time.Duration(3 * time.Minute)
		client := http.Client{
			Timeout: timeout,
		}
		response, err := client.Get(u.url)
		if err != nil {
			l.WithField("err", err).Error("could not download book")
			earlyExit = s.slowDown()
			if earlyExit {
				l.Warning("early exit")
			}
			continue
		}
		defer response.Body.Close()
		file, err := os.Create(output)
		defer file.Close()
		if err != nil {
			l.WithField("err", err).Error("could not open output file")
			continue
		}
		_, err = io.Copy(file, response.Body)
		if err != nil {
			l.WithField("err", err).Error("could not write file with body")
			earlyExit = s.slowDown()
			if earlyExit {
				l.Warning("early exit")
			}
			continue
		}
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
		if len(b.Authors) == 0 {
			log.WithFields(log.Fields{
				"title": b.Title,
				"date":  b.Timestamp,
			}).Debug("book has no authors")
			continue
		}
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
		inDB, hash := bookInDatabase(&b)
		if inDB {
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
				hash: hash,
			}
		}

	}

	return nil
}

func (s *scrapeConfig) bookTooOld(b *CalibreBook) bool {
	return b.Timestamp.Before(s.maxAge)
}

func (s *scrapeConfig) slowDown() bool {
	s.Lock()
	s.failures++
	s.stepSize = s.stepSize / 2
	if s.stepSize < 15 {
		s.stepSize = 20
	}
	s.backoffTime = (s.backoffTime / 2) * 3
	if s.backoffTime > 45*time.Second {
		s.backoffTime = 45 * time.Second
	}
	s.timeout = (s.timeout / 2) * 3
	if s.timeout > 30*time.Second {
		s.backoffTime = 30 * time.Second
	}
	s.logger.WithFields(log.Fields{
		"duration": s.backoffTime.String(),
		"failures": s.failures,
		"stepsize": s.stepSize,
	}).Warning("slowing down")
	time.Sleep(s.backoffTime)

	s.Unlock()
	return s.failures > 10
}

func bookInDatabase(b *CalibreBook) (bool, string) {
	title := fix(b.Title, true, false)
	author := fix(b.Authors[0], true, true)
	hash := hashBook(author, title)
	var book Book
	err := db.Conn.One("Hash", hash, &book)
	return err == nil, hash
}

func (s *scrapeConfig) markBookAsDownloaded(b *CalibreBook) error {

	var book Book
	book.Added = time.Now()
	book.Author = fix(b.Authors[0], true, true)
	book.Title = fix(b.Title, true, false)
	book.Hash = hashBook(b.Authors[0], b.Title)
	book.SourceID = s.hostID

	return db.Conn.Save(&book)
}

func fix(s string, capitalize, correctOrder bool) string {
	if s == "" {
		return "Unknown"
	}
	if capitalize {
		s = strings.Title(strings.ToLower(s))
		s = strings.Replace(s, "'S", "'s", -1)
	}
	if correctOrder && strings.Contains(s, ",") {
		sParts := strings.Split(s, ",")
		if len(sParts) == 2 {
			s = strings.TrimSpace(sParts[1]) + " " + strings.TrimSpace(sParts[0])
		}
	}

	s = yearRemove.ReplaceAllString(s, "")
	s = drukRemove.ReplaceAllString(s, "")
	s = strings.Replace(s, ".", " ", -1)
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.TrimSpace(s)

	return strings.Map(func(in rune) rune {
		switch in {
		case '“', '‹', '”', '›':
			return '"'
		case '_':
			return ' '
		case '‘', '’':
			return '\''
		}
		return in
	}, s)
}
