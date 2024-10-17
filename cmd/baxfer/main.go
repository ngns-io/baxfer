package main

import (
	"os"

	"github.com/ngns-io/baxfer/internal/cli"
	"github.com/ngns-io/baxfer/pkg/logger"
)

func main() {
	log, err := logger.New("baxfer.log")
	if err != nil {
		panic(err)
	}
	defer log.Close()

	app := cli.NewApp(log)
	err = app.Run(os.Args)
	if err != nil {
		log.Error("Application error", "error", err)
		os.Exit(1)
	}
}