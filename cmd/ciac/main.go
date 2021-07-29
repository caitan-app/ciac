package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var app *cli.App

func init() {
	app = &cli.App{
		Name:    filepath.Base(os.Args[0]),
		Usage:   "Crypto Investment Advisor Client",
		Version: "0.1.2",
	}

	app.Commands = []*cli.Command{
		timestampCommand,
		sendCodeCommand,
		registerCommand,
		loginCommand,
		userCommand,
		invitedCommand,
		rechargedCommand,
		bindCommand,
	}
	app.Flags = []cli.Flag{
		ConfigFlag,
		ServerFlag,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	if err := app.RunContext(ctx, os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
