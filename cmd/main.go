package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test/internal/handlers"
	"test/internal/worker"
	"time"

	"test/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := storage.NewSQLite("links.db")
	if err != nil {
		log.Fatalf("faild open", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	w := worker.NewWorker(db)
	w.Start()
	defer w.Stop()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	h := handlers.NewHandler(db, w)
	h.Register(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server stopped")
}
