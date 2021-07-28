package main

import "github.com/urfave/cli/v2"

var (
	ConfigFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Value:   "config.json",
		Usage:   "load configuration from `file`",
	}
	ServerFlag = &cli.StringFlag{
		Name:    "server",
		Aliases: []string{"s"},
		Value:   "https://test.caitan.app",
		Usage:   "connect to `server`",
	}
	EmailFlag = &cli.StringFlag{
		Name:  "email",
		Usage: "send verification code to `email`",
	}
	ForceFlag = &cli.BoolFlag{
		Name:  "force",
		Usage: "force to login (always update token)",
	}
	VerificationCodeFlag = &cli.StringFlag{
		Name:  "vc",
		Usage: "verification `code`",
	}
	InvitationCodeFlag = &cli.StringFlag{
		Name:  "ic",
		Usage: "invitation `code`",
	}
)