package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Fatal:", err.Error())
		os.Exit(1)
	}
}

func run() error {
	root := flag.String("root", "/goclaw-data", "The root folder for data to be stored at")
	flag.Parse()

	if err := UpdateConfig(*root); err != nil {
		return err
	}

	data, err := LoadData(*root)
	if err != nil {
		return err
	}

	ag, err := CreateAgent(data)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- ag.Run()
	}()

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		fmt.Println("Shutting down gracefully...")
		ag.CleanStop()
		ag.RemoveAllPlugins()
		return nil
	}
}
