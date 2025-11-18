package worker

import (
	"log"
	"net/http"
	"sync"
	"time"

	"test/internal/models"

	"gorm.io/gorm"
)

type Worker struct {
	db      *gorm.DB
	Queue   chan uint
	stop    chan struct{}
	wg      sync.WaitGroup
	workers int
}

func NewWorker(db *gorm.DB) *Worker {
	return &Worker{
		db:      db,
		Queue:   make(chan uint, 200),
		stop:    make(chan struct{}),
		workers: 3,
	}
}

func (w *Worker) Start() {
	var pending []models.Link
	if err := w.db.Where("processed = ?", false).Find(&pending).Error; err == nil {
		for _, p := range pending {
			select {
			case w.Queue <- p.ID:
			default:

			}
		}
	}

	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.loop()
		}()
	}
}

func (w *Worker) Stop() {
	close(w.stop)
	w.wg.Wait()
}

func (w *Worker) loop() {
	for {
		select {
		case id := <-w.Queue:
			w.processLink(id)
		case <-w.stop:
			log.Println("worker received stop")
			return
		}
	}
}

func (w *Worker) processLink(id uint) {
	var l models.Link
	if err := w.db.First(&l, id).Error; err != nil {
		return
	}

	if l.Processed {
		return
	}

	client := http.Client{
		Timeout: 8 * time.Second,
	}

	status := "fail"

	resp, err := client.Get(l.URL)
	if err == nil {
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			status = "ok"
		}
	}

	l.Status = status
	l.Processed = true
	l.UpdatedAt = time.Now()
	if err := w.db.Save(&l).Error; err != nil {
		log.Println("failed to save link status:", err)
	}
}
