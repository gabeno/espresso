/*
Main entry point to the server.

It reads configuration, sets up logging and error handling,
handles signals fron the OS and starts and stops the server.
*/
package main

import (
	"canvas/server"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	os.Exit(start())
}

func start() int {
	logEnv := getStringOrDefault("LOG_ENV", "development")
	log, err := createLogger(logEnv)
	if err != nil {
		fmt.Println("Error setting up the logger", err)
		return 1
	}
	defer func() {
		// if we can not sync, there is probably something wrong with
		// outputting the logs so we probably can not write using
		// fmt.Println either. So we ignore the error.

		// ensure all log messages are output before the program exits
		_ = log.Sync()
	}()

	host := getStringOrDefault("HOST", "localhost")
	port := getIntOrDefault("PORT", 8080)

	s := server.New(server.Options{
		Host: host,
		Log:  log,
		Port: port,
	})

	var eg errgroup.Group
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	eg.Go(func() error {
		<-ctx.Done()
		if err := s.Stop(); err != nil {
			log.Info("Errot stopping server", zap.Error(err))
			return err
		}
		return nil
	})

	if err := s.Start(); err != nil {
		log.Info("Error starting server", zap.Error(err))
		return 1
	}

	if err := eg.Wait(); err != nil {
		return 1
	}

	return 0
}

func createLogger(env string) (*zap.Logger, error) {
	switch env {
	case "production":
		return zap.NewProduction()
	case "development":
		return zap.NewDevelopment()
	default:
		return zap.NewNop(), nil
	}
}

func getStringOrDefault(name, defaultV string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		return defaultV
	}
	return v
}

func getIntOrDefault(name string, defaultV int) int {
	v, ok := os.LookupEnv(name)
	if !ok {
		return defaultV
	}
	vAsInt, err := strconv.Atoi(v)
	if err != nil {
		return defaultV
	}
	return vAsInt
}
