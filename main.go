// Application which greets you.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-templates/seed/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)

	go func() {
		err := server.StartServer(ctx)
		if err != nil && err == http.ErrServerClosed {
			log.Printf("Error: %v\n", err)
		}
	}()

	<-sigs
	cancel()
	log.Println("Exiting")
}

func greet() string {
	return "Hi!"
}
