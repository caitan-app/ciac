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
	invitedCommand = &cli.Command{
		Action: invited,
		Name:   "invited",
		Usage:  "List invited records",
		Flags: []cli.Flag{
			StartFlag,
			EndFlag,
			PageFlag,
			PageSizeFlag,
		},
	}
	rechargedCommand = &cli.Command{
		Action: recharged,
		Name:   "recharged",
		Usage:  "List recharged records",
		Flags: []cli.Flag{
			StartFlag,
			EndFlag,
			PageFlag,
			PageSizeFlag,
		},
	}
	bindCommand = &cli.Command{
		Action: bind,
		Name:   "bind",
		Usage:  "Bind invitation code",
		Flags: []cli.Flag{
			CodeFlag,
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
	token, err := endpoint.Login(force)
	if err != nil {
		log.Printf("Login error: %s", err)
		return err
	}
	log.Printf("Login success, token is %s, expire at %s", token.JWT, token.ExpireAt)
	return nil
}

// user list user info
func user(c *cli.Context) error {
	server := c.String(ServerFlag.Name)
	log.Printf("Server is %s", server)
	cfg, err := parseConfig(c.String(ConfigFlag.Name))
	if err != nil {
		return err
	}
	endpoint := client.New(cfg, server)

	profile, err := endpoint.UserInfo(c.Context)
	if err != nil {
		log.Printf("get user info error: %s", err)
		return err
	}
	log.Printf("Profile:\n%s", client.BeautifyJson(profile))
	return nil
}

// invited list invited records
func invited(c *cli.Context) error {
	cfg, err := parseConfig(c.String(ConfigFlag.Name))
	if err != nil {
		return err
	}
	server := c.String(ServerFlag.Name)
	log.Printf("Server is %s", server)
	endpoint := client.New(cfg, server)

	start := c.Int64(StartFlag.Name)
	end := c.Int64(EndFlag.Name)
	page := c.Int(PageFlag.Name)
	pageSize := c.Int(PageSizeFlag.Name)
	records, err := endpoint.InvitationRecords(c.Context, start, end, page, pageSize)
	if err != nil {
		log.Printf("Get invitation records error: %s", err)
		return err
	}
	log.Printf("id	nickName	rewardType	rewardNumber	rewardUnit	rewardTime")
	for i, r := range records {
		rewardAt := time.Unix(r.RewardTime/1000, 0)
		log.Printf("%d	%s	%d	%d %d	%s", i+1, r.NickName, r.RewardType, r.RewardNumber, r.RewardUnit, rewardAt)
	}
	return nil
}

// recharged list recharged records
func recharged(c *cli.Context) error {
	cfg, err := parseConfig(c.String(ConfigFlag.Name))
	if err != nil {
		return err
	}
	server := c.String(ServerFlag.Name)
	log.Printf("Server is %s", server)
	endpoint := client.New(cfg, server)

	start := c.Int64(StartFlag.Name)
	end := c.Int64(EndFlag.Name)
	page := c.Int(PageFlag.Name)
	pageSize := c.Int(PageSizeFlag.Name)
	records, err := endpoint.RechargeRecords(c.Context, start, end, page, pageSize)
	if err != nil {
		log.Printf("Get recharge records error: %s", err)
		return err
	}
	log.Printf("id	nickName	rechargeFrom	rechargeTo	rechargeNumber	rechargeUnit	rechargeTime")
	for i, r := range records {
		rewardAt := time.Unix(r.RechargeTime/1000, 0)
		log.Printf("%d	%s	%s	%s	%f %d	%s", i, r.NickName, r.RechargeFrom, r.RechargeTo, r.RechargeNumber, r.RechargeUnit, rewardAt)
	}
	return nil
}

func bind(c *cli.Context) error {
	code := c.String(CodeFlag.Name)
	if code == "" {
		log.Println("you need specify invitation code")
		return nil
	}
	cfg, err := parseConfig(c.String(ConfigFlag.Name))
	if err != nil {
		return err
	}
	server := c.String(ServerFlag.Name)
	log.Printf("Server is %s", server)
	endpoint := client.New(cfg, server)
	success, err := endpoint.Bind(c.Context, code)
	if err != nil || !success {
		log.Printf("bind failed, error: %s", err)
		return err
	}
	log.Println("bind success")
	return nil
}

func parseConfig(filename string) (client.Config, error) {
	var c client.Config
	if err := hs.ParseJsonConfig(filename, &c); err != nil {
		return c, err
	}
	return c, nil
}
