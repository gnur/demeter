package lib

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// WorkerQueues holds all the queues that the worker can act upon
type WorkerQueues struct {
	IDS    chan GetIDSRequest
	Books  chan GetBooksRequest
	DlBook chan DownloadBookRequest
}

// WorkerCounter counts all the work a worker did
type WorkerCounter struct {
	GetIDS   int
	GetBooks int
	DlBook   int
	ID       int
}

// Worker is the only unit that actually makes requests
func (a *App) Worker(id int, q WorkerQueues) {
	c := WorkerCounter{
		ID: id,
	}
	ticker := time.Tick(a.WorkerInterval)
	l := log.WithField("worker", fmt.Sprintf("worker_%02d", c.ID))
	defer l.Info("Ending work routine")
	for {
		select {
		case re := <-q.IDS:
			ids, err := a.getIDS(re.U, re.Offset, re.Num)
			re.Resp <- GetIDSResponse{
				IDs: ids,
				Err: err,
			}
			c.GetIDS++
		case re := <-q.Books:
			books, err := a.getBooks(re.U, re.IDs)
			re.Resp <- GetBooksResponse{
				Books: books,
				Err:   err,
			}
			c.GetBooks++
		case re := <-q.DlBook:
			err := a.downloadBook(re.URL, re.Path)
			re.Resp <- DownloadBookResponse{
				Err: err,
			}
			c.DlBook++
		case _ = <-ticker:
			l.WithFields(log.Fields{
				"GetIDS":   c.GetIDS,
				"GetBooks": c.GetBooks,
				"DlBook":   c.DlBook,
			}).Info("Worker update")
		}
	}
}
