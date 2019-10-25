package lib

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

// Scrape performs the actual scrape
func (a *App) Scrape(h *Host) (*ScrapeResult, error) {
	parsed, err := url.Parse(h.URL)
	if err != nil {
		return nil, err
	}
	parsed.Path = ""

	r := ScrapeResult{
		Start:   time.Now(),
		Success: false,
	}
	defer func() {
		r.End = time.Now()
	}()

	err = os.MkdirAll(a.OutputDir, 0755)
	if err != nil {
		return &r, err
	}

	allIDs, err := a.getAllIDS(*parsed)
	if err != nil {
		return &r, err
	}
	ids := a.filterOldIDs(allIDs, h.ID)
	log.WithFields(log.Fields{
		"allIDs": len(allIDs),
		"ids":    len(ids),
	}).Info("Filtered results")

	i := 0
	toDownload := 0
	dlResultQueue := make(chan DownloadBookResponse, len(ids))
	r.Results = len(ids)
	for i < len(ids) {
		max := i + a.StepSize
		if max > len(ids) {
			max = len(ids)
		}
		bs, err := a.getBooksAsync(*parsed, ids[i:max])
		i += a.StepSize
		if err != nil {
			log.WithField("err", err).Error("Could not get books")
			continue
		}
		for _, b := range bs {
			if present, hash := bookInDatabase(&b); !present {
				if epubPath, ok := b.MainFormat["epub"]; ok {
					rawPath, err := url.QueryUnescape(epubPath)
					if err != nil {
						continue
					}
					parsed.Path = rawPath
					output := fmt.Sprintf("%s.epub", hash)
					output = path.Join(a.OutputDir, output)
					//TODO: dit in go routine
					a.Queues.DlBook <- DownloadBookRequest{
						URL:  parsed.String(),
						Path: output,
						Resp: dlResultQueue,
					}
					toDownload++
				}
			}
		}
	}

	// ######################################################
	for i := 0; i < toDownload; i++ {
		res := <-dlResultQueue
		if res.Err == nil {
			r.Downloads++
		}
	}
	close(dlResultQueue)

	return &r, nil

}
