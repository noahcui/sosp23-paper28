package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	logger "github.com/sirupsen/logrus"
	monitor "github.com/sosp23/replicated-store/go/Monitor"
	"github.com/sosp23/replicated-store/go/config"
	"github.com/sosp23/replicated-store/go/replicant"
)

func main() {
	id := flag.Int64("id", 0, "peer id")
	debug := flag.Bool("d", false, "enable debug logging")
	configPath := flag.String("c", "../c++/config.json", "config path")
	flag.Parse()

	if *debug {
		logger.SetLevel(logger.InfoLevel)
	} else {
		logger.SetLevel(logger.ErrorLevel)
	}

	cfg, err := config.LoadConfig(*id, *configPath)
	if err != nil {
		logger.Panic(err)
	}
	logger.Info("Config loaded")
	replicant := replicant.NewReplicant(cfg)
	logger.Info("Replicant started")
	monitor := monitor.NewMonitor(replicant, "bufferinfo", -1, 10)
	go monitor.Run()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		replicant.Stop()
	}()
	replicant.Start()
}
