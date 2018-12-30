// Copyright © 2018 anon hoarder
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/anonhoarder/demeter/db"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "demeter",
	Short: "DEMETER WILL EAT ALL YOUR BOOKS",
	Long: `demeter is CLI application for scraping calibre hosts and
retrieving books in epub format that are not in your local library.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
			return
		}
		dbDir := path.Join(home, ".demeter")
		err = os.MkdirAll(dbDir, 0755)
		if err != nil {
			log.Fatal(err)
			return
		}
		dbPath := path.Join(dbDir, "demeter.db")
		db.Conn, err = storm.Open(dbPath, storm.Codec(msgpack.Codec), storm.Batch())
		if err != nil {
			log.Fatal(err)
			return
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		err := db.Conn.Close()
		if err != nil {
			log.WithField("err", err).Error("Could not close database")
			return
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
