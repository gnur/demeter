// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"math/rand"
	"sync"
	"time"

	"github.com/anonhoarder/demeter/db"
	"github.com/anonhoarder/demeter/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stepSize int
var workers int
var userAgent string
var outputDir string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run all scrape jobs",
	Long: `Go over all defined hosts and if the last scrape
is old enough it will scrape that host.`,
	Run: func(cmd *cobra.Command, args []string) {
		var hosts []lib.Host
		db.Conn.Find("Active", true, &hosts)

		if len(hosts) == 0 {
			log.Info("no active hosts were found")
			return
		}

		semaphore := make(chan bool, 30)
		var wg sync.WaitGroup

		for _, h := range hosts {
			wg.Add(1)
			go func(h lib.Host) {
				log.WithField("host", h.URL).Info("Starting work")
				defer func() {
					log.WithField("host", h.URL).Info("Finished work")
					wg.Done()
				}()
				jitter := time.Duration(rand.Intn(3600)) * time.Second
				cutOffPoint := time.Now().Add(-jitter).Add(-24 * time.Hour)
				if !h.LastScrape.Before(cutOffPoint) {
					log.WithFields(log.Fields{
						"host":        h.URL,
						"last_scrape": h.LastScrape,
					}).Info("not scraping because it is too recent")
					return

				}
				log.WithFields(log.Fields{
					"workers":   workers,
					"useragent": userAgent,
					"outputdir": outputDir,
					"stepsize":  stepSize,
					"host":      h.URL,
				}).Debug("Starting scrape")
				result, err := h.Scrape(workers, stepSize, userAgent, outputDir, semaphore)
				if err != nil {
					log.WithFields(log.Fields{
						"host": h.URL,
						"err":  err,
					}).Error("Scraping failed")
					failedScrapes := 0
					for _, s := range h.ScrapeResults {
						if !s.Success {
							failedScrapes++
						}
					}

				} else {
					log.WithFields(log.Fields{
						"host":      h.URL,
						"downloads": result.Downloads,
						"duration":  time.Since(result.Start).String(),
						"err":       err,
					}).Info("Scraping done")
				}
				h.Downloads += result.Downloads
				h.Scrapes++
				if result.Downloads > 0 {
					h.LastDownload = result.End
				}
				dls, fails := h.Stats(10)
				if dls == 0 && fails >= 5 {
					h.Active = false
					err = db.Conn.UpdateField(&h, "Active", false)
					log.WithFields(log.Fields{
						"host":    h.URL,
						"scrapes": h.Scrapes,
					}).Warning("Disabling host because there were no new downloads recently")
				}
				h.LastScrape = result.End

				h.ScrapeResults = append(h.ScrapeResults, *result)
				err = db.Conn.Update(&h)
				if err != nil {
					log.WithFields(log.Fields{
						"host": h.URL,
						"err":  err,
					}).Error("Could not store scrape result, exiting hard")
				}
			}(h)
		}
		log.Info("waiting for all members in wg to finish")
		wg.Wait()
	},
}

var runCmdV2 = &cobra.Command{
	Use:   "runv2",
	Short: "run all scrape jobs with new method",
	Long: `Go over all defined hosts and if the last scrape
is old enough it will scrape that host.`,
	Run: func(cmd *cobra.Command, args []string) {
		var hosts []lib.Host
		db.Conn.Find("Active", true, &hosts)

		if len(hosts) == 0 {
			log.Info("no active hosts were found")
			return
		}

		qs := lib.WorkerQueues{
			IDS:    make(chan lib.GetIDSRequest),
			Books:  make(chan lib.GetBooksRequest),
			DlBook: make(chan lib.DownloadBookRequest),
		}

		a := lib.App{
			UserAgent:       userAgent,
			Timeout:         3 * time.Minute,
			DownloadTimeout: 5 * time.Minute,
			WorkerInterval:  5 * time.Minute,
			StepSize:        stepSize,
			OutputDir:       outputDir,
			Queues:          qs,
		}

		for i := 0; i < workers; i++ {
			go a.Worker(i, qs)
		}
		var wg sync.WaitGroup

		for _, h := range hosts {
			go func(h lib.Host) {
				log.WithField("host", h.URL).Info("Starting work")
				jitter := time.Duration(rand.Intn(3600)) * time.Second
				cutOffPoint := time.Now().Add(-jitter).Add(-24 * time.Hour)
				if !h.LastScrape.Before(cutOffPoint) {
					log.WithFields(log.Fields{
						"host":        h.URL,
						"last_scrape": h.LastScrape,
					}).Info("not scraping because it is too recent")
					return
				}
				wg.Add(1)
				result, err := a.Scrape(&h)
				if err != nil {
					log.WithFields(log.Fields{
						"host": h.URL,
						"err":  err,
					}).Error("Scraping failed")
					failedScrapes := 0
					for _, s := range h.ScrapeResults {
						if !s.Success {
							failedScrapes++
						}
					}

				} else {
					log.WithFields(log.Fields{
						"host":      h.URL,
						"downloads": result.Downloads,
						"duration":  time.Since(result.Start).String(),
						"err":       err,
					}).Info("Scraping done")
				}
				h.Downloads += result.Downloads
				h.Scrapes++
				if result.Downloads > 0 {
					h.LastDownload = result.End
				}
				dls, fails := h.Stats(10)
				if dls == 0 && fails >= 5 {
					h.Active = false
					err = db.Conn.UpdateField(&h, "Active", false)
					log.WithFields(log.Fields{
						"host":    h.URL,
						"scrapes": h.Scrapes,
					}).Warning("Disabling host because there were no new downloads recently")
				}
				h.LastScrape = result.End

				h.ScrapeResults = append(h.ScrapeResults, *result)
				err = db.Conn.Update(&h)
				if err != nil {
					log.WithFields(log.Fields{
						"host": h.URL,
						"err":  err,
					}).Error("Could not store scrape result, exiting hard")
				}
				wg.Done()
			}(h)
		}
		time.Sleep(5 * time.Second)
		wg.Wait()
	},
}

func init() {
	scrapeCmd.AddCommand(runCmd)
	scrapeCmd.AddCommand(runCmdV2)

	runCmd.Flags().IntVarP(&stepSize, "stepsize", "n", 300, "number of books to request per query")
	runCmd.Flags().IntVarP(&workers, "workers", "w", 5, "number of workers to concurrently download books")
	runCmd.Flags().StringVarP(&userAgent, "useragent", "u", "demeter / alpha", "user agent used to identify to calibre hosts")
	runCmd.Flags().StringVarP(&outputDir, "outputdir", "d", "books", "path to downloaded books to")

	runCmdV2.Flags().IntVarP(&workers, "workers", "w", 10, "number of workers to concurrently download books")
	runCmdV2.Flags().IntVarP(&stepSize, "stepsize", "n", 50, "number of books to request per query")
	runCmdV2.Flags().StringVarP(&userAgent, "useragent", "u", "demeter / v2", "user agent used to identify to calibre hosts")
	runCmdV2.Flags().StringVarP(&outputDir, "outputdir", "d", "books", "path to downloaded books to")
}
