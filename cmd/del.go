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
	"strconv"

	"github.com/anonhoarder/demeter/lib"
	log "github.com/sirupsen/logrus"

	"github.com/anonhoarder/demeter/db"

	"github.com/spf13/cobra"
)

var deleteID uint32

// listCmd represents the list command
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

func init() {
	hostCmd.AddCommand(delCmd)
}
