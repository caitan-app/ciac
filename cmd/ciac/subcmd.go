package main

import (
	"github.com/caitan-app/ciac/client"
	"github.com/urfave/cli/v2"
	"github.com/xyths/hs"

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
			EmailFlag,
		},
	}
	registerCommand = &cli.Command{
		Action: register,
		Name:   "register",
		Usage:  "Register a new user",
		Flags: []cli.Flag{
			VerificationCodeFlag,
			InvitationCodeFlag,
		},
	}
	loginCommand = &cli.Command{
		Action: login,
		Name:   "login",
		Usage:  "Login to the server, cache the cookie",
		Flags: []cli.Flag{
			ForceFlag,
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

func sendCode(c *cli.Context) error {
	server := c.String(ServerFlag.Name)
	log.Printf("Server is %s", server)
	email := c.String(EmailFlag.Name)
	if email == "" {
		conf := c.String(ConfigFlag.Name)
		cfg, err := parseConfig(conf)
		if err != nil {
			log.Printf("no email specified, and has no valid config(config file is %s, got error: %s)", conf, err)
			return err
		}
		email = cfg.Email
	}
	log.Printf("Email is %s", email)

	return client.SendCode(server, email)
}

func register(c *cli.Context) error {
	server := c.String(ServerFlag.Name)
	vc := c.String(VerificationCodeFlag.Name)
	ic := c.String(InvitationCodeFlag.Name)
	log.Printf("Server is %s", server)
	conf := c.String(ConfigFlag.Name)
	cfg, err := parseConfig(conf)
	if err != nil {
		log.Printf("no email specified, and has no valid config(config file is %s, got error: %s)", conf, err)
		return err
	}
	email := cfg.Email
	log.Printf("Email is %s", email)

	return client.Register(server, email, cfg.Password, vc, ic)
}

func login(c *cli.Context) error {
	server := c.String(ServerFlag.Name)
	log.Printf("Server is %s", server)
	cfg, err := parseConfig(c.String(ConfigFlag.Name))
	if err != nil {
		return err
	}
	force := c.Bool(ForceFlag.Name)
	endpoint := client.New(cfg, server)
	var token *client.Token
	if token, err = endpoint.Login(force); err != nil {
		log.Printf("Login error: %s", err)
		return err
	}
	log.Printf("Login success, token is %s, expire at %s", token.JWT, token.ExpireAt)
	return nil
}

func user(ctx *cli.Context) error {
	return nil
}

func parseConfig(filename string) (client.Config, error) {
	var c client.Config
	if err := hs.ParseJsonConfig(filename, &c); err != nil {
		return c, err
	}
	return c, nil
}
