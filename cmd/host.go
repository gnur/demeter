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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gnur/demeter/db"
	"github.com/gnur/demeter/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteID uint32

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "all host related commands",
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all hosts",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		var hosts []lib.Host
		db.Conn.All(&hosts)

		if len(hosts) == 0 {
			log.Info("no hosts were found")
		}

		for i, h := range hosts {
			if i%25 == 0 {
				fmt.Printf(`%5s|%30s|%7s|%7s|%5s|%6s|%6s|%6s`, "id", "url", "total", "dls(5)", "fails", "scrape", "dls", "active")
				fmt.Println()
			}
			h.Print(false)
			fmt.Println()
		}

	},
}

var addCmd = &cobra.Command{
	Use:   "add hosturl [hosturl] ...",
	Args:  cobra.MinimumNArgs(1),
	Short: "add one or more hosts to the scrape list",
	Run: func(cmd *cobra.Command, args []string) {
		for _, hosturl := range args {
			u, err := url.Parse(hosturl)
			if err != nil {
				log.WithField("err", err).Error("invalid url provided")
				return
			}
			u.Path = ""
			h := lib.Host{
				URL:        strings.ToLower(u.String()),
				LastScrape: time.Now().Add(-20 * 365 * 24 * time.Hour),
				Active:     true,
			}

			err = db.Conn.Save(&h)
			if err != nil {
				log.WithField("err", err).Error("could not save")
				return
			}
			log.WithFields(log.Fields{
				"id":  h.ID,
				"url": h.URL,
			}).Info("host has been added to the database")
		}
	},
}

var delCmd = &cobra.Command{
	Use:     "rm hostid",
	Aliases: []string{"del", "rm", "delete", "remove"},
	Short:   "delete a host",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var h lib.Host
		id, err := strconv.Atoi(args[0])
		if err != nil {
			log.WithField("err", err).Error("please provide a numeric ID")
			return
		}
		err = db.Conn.One("ID", id, &h)
		if err != nil {
			log.WithField("err", err).Error("No host with that ID was found")
			return
		}
		db.Conn.DeleteStruct(&h)
		log.WithField("host", h.URL).Info("host was removed")

	},
}

var detailCmd = &cobra.Command{
	Use:     "stats hostid",
	Aliases: []string{"detail", "info", "details", "stats"},
	Short:   "Get host stats",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var h lib.Host
		id, err := strconv.Atoi(args[0])
		if err != nil {
			log.WithField("err", err).Error("please provide a numeric ID")
			return
		}
		err = db.Conn.One("ID", id, &h)
		if err != nil {
			log.WithField("err", err).Error("No host with that ID was found")
			return
		}
		h.Print(true)
	},
}

var disableCmd = &cobra.Command{
	Use:     "disable hostid",
	Aliases: []string{"dis", "deactivate", "disable"},
	Short:   "disable a host",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var h lib.Host
		id, err := strconv.Atoi(args[0])
		if err != nil {
			log.WithField("err", err).Error("please provide a numeric ID")
			return
		}
		err = db.Conn.One("ID", id, &h)
		if err != nil {
			log.WithField("err", err).Error("No host with that ID was found")
			return
		}
		h.Active = false
		err = db.Conn.UpdateField(&h, "Active", false)
		if err != nil {
			log.WithFields(log.Fields{
				"host":   h.URL,
				"err":    err,
				"active": h.Active,
			}).Error("Could not store new active state")
			return
		}
		log.WithFields(log.Fields{
			"host":   h.URL,
			"id":     h.ID,
			"active": h.Active,
		}).Info("host was disabled")

	},
}

var enableCmd = &cobra.Command{
	Use:     "enable hostid",
	Aliases: []string{"en", "activate", "enable"},
	Short:   "make a host active",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var h lib.Host
		id, err := strconv.Atoi(args[0])
		if err != nil {
			log.WithField("err", err).Error("please provide a numeric ID")
			return
		}
		err = db.Conn.One("ID", id, &h)
		if err != nil {
			log.WithField("err", err).Error("No host with that ID was found")
			return
		}
		h.Active = true
		err = db.Conn.Update(&h)
		if err != nil {
			log.WithFields(log.Fields{
				"host":   h.URL,
				"err":    err,
				"active": h.Active,
			}).Error("Could not store new active state")
			return
		}
		log.WithField("host", h.URL).Info("host was activated")

	},
}

var enableAllCmd = &cobra.Command{
	Use:   "enable-all",
	Short: "make all hosts active",
	Run: func(cmd *cobra.Command, args []string) {
		var hosts []lib.Host
		db.Conn.All(&hosts)
		for _, h := range hosts {
			h.Active = true
			err := db.Conn.Update(&h)
			if err != nil {
				log.WithFields(log.Fields{
					"host":   h.URL,
					"err":    err,
					"active": h.Active,
				}).Error("Could not store new active state")
				return
			}
			log.WithField("host", h.URL).Info("host was activated")
		}

	},
}

func init() {
	rootCmd.AddCommand(hostCmd)
	hostCmd.AddCommand(addCmd)
	hostCmd.AddCommand(listCmd)
	hostCmd.AddCommand(delCmd)
	hostCmd.AddCommand(enableCmd)
	hostCmd.AddCommand(enableAllCmd)
	hostCmd.AddCommand(disableCmd)
	hostCmd.AddCommand(detailCmd)
}
