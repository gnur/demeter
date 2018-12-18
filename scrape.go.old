package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {

	dlChan := make(chan string)
	doneChan := make(chan int)
	for w := 1; w <= workers; w++ {
		go bookDLWorker(w, dlChan, doneChan)
	}

	i := 0
	for i < len(ids) {
		max := i + stepSize
		if max > len(ids) {
			max = len(ids)
		}
		log.WithFields(log.Fields{
			"min":   i,
			"max":   max,
			"total": len(ids),
		}).Info("retrieving books")
		err = getBooks(parsed.String(), ids[i:max], dlChan)
		if err != nil {
			log.WithField("err", err).Error("could not get books")
			stepSize = (stepSize / 3) * 2
			if stepSize < 20 {
				stepSize = 20
			}
			time.Sleep(5 * time.Second)
		}
		i += stepSize

	}
	close(dlChan)

	downloaded := 0

	for a := 1; a <= workers; a++ {
		workerDownloads := <-doneChan
		downloaded += workerDownloads
	}
	processTime := time.Since(startTime)
	log.WithFields(log.Fields{
		"host":       parsed.Host,
		"downloaded": downloaded,
		"duration":   processTime.String(),
		"total":      len(ids),
	}).Info("done with host")
}
