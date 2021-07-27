package main

import (
	"github.com/caitan-app/ciac/client"
	"github.com/urfave/cli/v2"

	"log"
	"time"
)

var (
	timestampCommand = &cli.Command{
		Action: timestamp,
		Name:   "timestamp",
		Usage:  "Get timestamp from the server",
		Flags: []cli.Flag{
		},
	}
	sendCodeCommand = &cli.Command{
		Action: sendCode,
		Name:   "code",
		Usage:  "Tell server to send code",
		Flags: []cli.Flag{
		},
	}
	registerCommand = &cli.Command{
		Action: register,
		Name:   "register",
		Usage:  "Register a new user",
		Flags: []cli.Flag{
		},
	}
	loginCommand = &cli.Command{
		Action: login,
		Name:   "login",
		Usage:  "Login to the server, cache the cookie",
		Flags: []cli.Flag{
		},
	}
	userCommand = &cli.Command{
		Action: user,
		Name:   "user",
		Usage:  "List user info",
		Flags: []cli.Flag{
		},
	}
)

func timestamp(c *cli.Context) error {
	server := c.String(ServerFlag.Name)
	t, err := client.Timestamp(server)
	if err != nil {
		return err
	}
	t2 := time.Unix(t/1000, t%1000)
	log.Printf("timestamp is %d (%s)", t, t2)
	return nil
}

func sendCode(ctx *cli.Context) error {
	return nil
}

func register(ctx *cli.Context) error {
	return nil
}

func login(ctx *cli.Context) error {
	return nil
}

func user(ctx *cli.Context) error {
	return nil
}
