package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/app"
)

const (
	PORT = 4221
)

var (
	directory = flag.String("directory", "", "--directory /tmp")
)

func main() {
	flag.Parse()

	if *directory != "" {
		_, err := os.Stat(*directory)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("directory %s does not exist\n", *directory)
			os.Exit(1)
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	config := &app.Config{
		Directory: *directory,
		Logger:    logger,
		Port:      PORT,
	}

	myApp := app.NewApp(config)

	// wait for termination signals to properly stop the server
	// go func() {
	// 	sigChan := make(chan os.Signal, 1)
	// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 	<-sigChan
	// 	logger.Info("received shutdown signal, stopping server...")
	// 	myApp.Stop()
	// }()

	err := myApp.Start()
	if err != nil {
		logger.Error("cannot start application", "error", err)
		os.Exit(1)
	}
}
