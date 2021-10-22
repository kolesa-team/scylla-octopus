package main

import (
	"context"
	"github.com/kolesa-team/scylla-octopus/cmd"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer stop()

	cmd.Execute(ctx)
}
