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
	"fmt"
	"time"

	"github.com/gnur/demeter/db"
	"github.com/gnur/demeter/lib"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// hostCmd represents the host command
var dlCmd = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download", "downloads", "dls"},
	Short:   "download related commands",
}

// dlListCmd represents the list command
var dlListCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all downloads",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		var books []lib.Book
		err := db.Conn.All(&books)
		if err != nil {
			fmt.Printf("Error listing downloads: %e\n", err)
		}

		fmt.Println("Total downloads: ", len(books))

	},
}

// dlListCmd represents the list command
var dlAddCmd = &cobra.Command{
	Use:   "add bookhash [bookhash]..",
	Args:  cobra.MinimumNArgs(1),
	Short: "add a number of hashes to the database",
	Run: func(cmd *cobra.Command, args []string) {
		tx, err := db.Conn.Begin(true)
		if err != nil {
			log.WithField("err", err).Error("could not start transaction")
			return
		}
		for _, hash := range args {
			h := lib.Book{
				Hash:     hash,
				Added:    time.Now(),
				SourceID: 0,
			}

			err := tx.Save(&h)
			if err != nil {
				log.WithField("err", err).Error("could not save")
				continue
			}
			log.WithFields(log.Fields{
				"id":   h.ID,
				"hash": h.Hash,
			}).Info("book has been added to the database")
		}
		err = tx.Commit()
		if err != nil {
			log.WithField("err", err).Error("failed to commit")
			return
		}
	},
}

var dlDelRecentCmd = &cobra.Command{
	Use:   "deleterecent 24h",
	Args:  cobra.ExactArgs(1),
	Short: "delete all downloads from this time period",
	Run: func(cmd *cobra.Command, args []string) {
		duration, err := time.ParseDuration(args[0])
		if err != nil {
			log.Error("invalid duration provided")
			return
		}
		cutOffPoint := time.Now().Add(-duration)
		log.WithField("cutoffpoint", cutOffPoint).Info("Deleting all downloads newer then this date")
		var books []lib.Book
		db.Conn.All(&books)
		deleted := 0
		scanned := 0
		for _, b := range books {
			scanned++
			if b.Added.After(cutOffPoint) {
				//db.Conn.DeleteStruct(&b)
				deleted++
			}
		}
		log.WithFields(log.Fields{
			"deleted": deleted,
			"scanned": scanned,
		}).Info("would have deleted these downloads")
	},
}

func init() {
	rootCmd.AddCommand(dlCmd)
	dlCmd.AddCommand(dlListCmd)
	dlCmd.AddCommand(dlDelRecentCmd)
	dlCmd.AddCommand(dlAddCmd)

}
