package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dimuls/graph/postgres"
	"github.com/dimuls/graph/web"
	"github.com/sirupsen/logrus"
)

func main() {
	postgresURI := os.Getenv("POSTGRES_URI")
	bindAddr := os.Getenv("BIND_ADDR")

	storage, err := postgres.NewStorage(postgresURI)
	if err != nil {
		logrus.WithError(err).Fatal(
			"failed to create new postgres storage")
	}

	err = storage.Migrate()
	if err != nil {
		logrus.WithError(err).Fatal(
			"failed to migrate postgres storage")
	}

	webServer := web.NewServer(bindAddr, storage)

	webServer.Start()

	ss := make(chan os.Signal, 2)
	signal.Notify(ss, os.Interrupt, syscall.SIGTERM)

	s := <-ss

	logrus.Info("captured %v signal, stopping", s)

	st := time.Now()
	webServer.Stop()

	logrus.Infof("stopped in %s, exiting", time.Now().Sub(st))
}
