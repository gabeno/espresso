/*
Main entry point to the server.

It reads configuration, sets up logging and error handling,
handles signals fron the OS and starts and stops the server.
*/
package main

import (
	"canvas/server"
	"canvas/storage"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maragudk/env"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// release is set through the linker at buildtime, generally from a git hash
// used for logging and error reporting
var release string

func main() {
	os.Exit(start())
}

func start() int {
	_ = env.Load()

	logEnv := env.GetStringOrDefault("LOG_ENV", "development")
	log, err := createLogger(logEnv)
	if err != nil {
		fmt.Println("Error setting up the logger", err)
		return 1
	}

	log = log.With(zap.String("release", release))

	defer func() {
		// if we can not sync, there is probably something wrong with
		// outputting the logs so we probably can not write using
		// fmt.Println either. So we ignore the error.

		// ensure all log messages are output before the program exits
		_ = log.Sync()
	}()

	host := env.GetStringOrDefault("HOST", "localhost")
	port := env.GetIntOrDefault("PORT", 8080)

	s := server.New(server.Options{
		Database: createDatabase(log),
		Host:     host,
		Log:      log,
		Port:     port,
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

func createDatabase(log *zap.Logger) *storage.Database {
	return storage.NewDatabase(storage.NewDatabaseOptions{
		Host:                  env.GetStringOrDefault("DB_HOST", "localhost"),
		Port:                  env.GetIntOrDefault("DB_PORT", 5432),
		User:                  env.GetStringOrDefault("DB_USER", ""),
		Password:              env.GetStringOrDefault("DB_PASSWORD", ""),
		Name:                  env.GetStringOrDefault("DB_NAME", ""),
		MaxOpenConnections:    env.GetIntOrDefault("DM_MAX_OPEN_CONNECTIONS", 10),
		MaxIdleConnections:    env.GetIntOrDefault("DB_MAX_IDLE_CONNECTIONS", 10),
		ConnectionMaxLifetime: env.GetDurationOrDefault("DB_CONECTION_MAX_LIFETIME", time.Hour),
		Log:                   log,
	})
}
