package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AsaHero/e-wallet/internal/app"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg, err := config.New()
	if err != nil {
		log.Fatalln("config init", err)
	}

	app, err := app.New(cfg)
	if err != nil {
		log.Fatalln("app init", err)
	}

	// run application
	go func() {
		if err := app.Run(); err != nil {
			if err == http.ErrServerClosed {
				return
			}

			log.Println("app run", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// app stops
	log.Println("ewallet server stopping...")
	app.Stop()
	log.Println("ewallet server stopped gracefully")
}
