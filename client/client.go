package client

import (
	"log"
	"time"
)

type Config struct {
	Email     string
	Password  string
	TokenFile string `json:"tokenFile"`
}

type Client struct {
	cfg    Config
	Server string

	token *Token
}

type Token struct {
	JWT      string `json:"jwt"`
	ExpireAt time.Time
}

func New(cfg Config, server string) *Client {
	return &Client{cfg: cfg, Server: server}
}

func (c Client) Email() string {
	return c.cfg.Email
}

func (c *Client) Login(force bool) (*Token, error) {
	if force {
		return c.loginAndSave()
	}

	if token, err := c.loadToken(); err != nil {
		return nil, err
	} else if token == nil {
		// first time run, login to get a token
		return c.loginAndSave()
	} else {
		// get a old cached token, check if it is expired
		if !token.ExpireAt.After(time.Now()) { // expired
			log.Printf("cookie expired at %s, need refresh", token.ExpireAt)
			return c.loginAndSave()
		}
		c.token = token
		return token, nil
	}
}
