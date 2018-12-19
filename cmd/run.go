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
			log.Info("no hosts were found")
			return
		}

		for _, h := range hosts {
			log.WithFields(log.Fields{
				"workers":   workers,
				"useragent": userAgent,
				"outputdir": outputDir,
				"stepsize":  stepSize,
				"host":      h.URL,
			}).Debug("Starting scrape")
			result, err := h.Scrape(workers, stepSize, userAgent, outputDir)
			if err != nil {
				log.WithFields(log.Fields{
					"host": h.URL,
					"err":  err,
				}).Error("Scraping failed")
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
			h.LastScrape = result.End

			h.ScrapeResults = append(h.ScrapeResults, *result)
			err = db.Conn.Update(&h)
			if err != nil {
				log.WithFields(log.Fields{
					"host": h.URL,
					"err":  err,
				}).Error("Could not store scrape result, exiting hard")
				return
			}

		}
	},
}

func init() {
	scrapeCmd.AddCommand(runCmd)

	runCmd.Flags().IntVarP(&stepSize, "stepsize", "n", 300, "number of books to request per query")
	runCmd.Flags().IntVarP(&workers, "workers", "w", 5, "number of workers to concurrently download books")
	runCmd.Flags().StringVarP(&userAgent, "useragent", "u", "demeter / alpha", "user agent used to identify to calibre hosts")
	runCmd.Flags().StringVarP(&outputDir, "outputdir", "d", "books", "path to downloaded books to")
}
