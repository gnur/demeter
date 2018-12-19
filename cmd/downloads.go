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

	"github.com/anonhoarder/demeter/db"
	"github.com/anonhoarder/demeter/lib"
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
		db.Conn.All(&books)

		fmt.Println("Total downloads: ", len(books))

	},
}

// dlListCmd represents the list command
var dlAddCmd = &cobra.Command{
	Use:   "add bookhash",
	Args:  cobra.ExactArgs(1),
	Short: "add a download",
	Run: func(cmd *cobra.Command, args []string) {
		h := lib.Book{
			Hash:     args[0],
			Added:    time.Now(),
			SourceID: 0,
		}

		err := db.Conn.Save(&h)
		if err != nil {
			log.WithField("err", err).Error("could not save")
			return
		}
		log.WithFields(log.Fields{
			"id":   h.ID,
			"hash": h.Hash,
		}).Info("book has been added to the database")
	},
}

func init() {
	rootCmd.AddCommand(dlCmd)
	dlCmd.AddCommand(dlListCmd)

}
