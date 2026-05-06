package main

import (
	"os"
	"os/signal"
	"syscall"

	"ai_interview/internal/app"
	"ai_interview/internal/config"
	"ai_interview/internal/infra/log"
)

func main() {
	if err := config.Load(); err != nil {
		log.Errorf("load config: %v", err)
		os.Exit(1)
	}

	log.Init(config.Cfg.LogLevel, config.Cfg.LogFormat, config.Cfg.Env)

	a, err := app.New(config.Cfg)
	if err != nil {
		log.Errorf("init app: %v", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		a.Shutdown()
	}()

	a.Run()
}
