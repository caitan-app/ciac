package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
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

type Profile struct {
	Email         string `json:"email"`
	Code          string `json:"invitationCode"`
	Expire        string `json:"expire"`
	RemainingTime string `json:"remainingTime"`
}

func (c *Client) UserInfo(ctx context.Context) error {
	if _, err := c.Login(false); err != nil {
		return err
	}

	u, err := url.Parse(c.Server)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "user")
	q := u.Query()
	q.Set("tamptime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()
	log.Printf("request URL %s", u)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token.JWT))
	hc := &http.Client{}
	resp, err := hc.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("Raw response: %s", string(body))

	var ret struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result     int
			NickName   string `json:"nickName"`
			Email      string
			Code       string
			Expire     string
			RemainTime int64 `json:"remainingTime"`
		}
	}
	if err = json.Unmarshal(body, &ret); err != nil {
		return err
	}
	log.Printf("Response:\n%s", beautifyJson(ret))

	profile := Profile{
		Email:         ret.Data.Email,
		Code:          ret.Data.Code,
		Expire:        ret.Data.Expire,
		RemainingTime: time.Duration(ret.Data.RemainTime * 1e6).String(),
	}
	log.Printf("Profile:\n%s", beautifyJson(profile))
	return nil
}

func (c *Client) InvitationRecords(ctx context.Context, start, end int64, page, pageSize int) error {
	if _, err := c.Login(false); err != nil {
		return err
	}

	url, err := pagingRequest(c.Server, "invitationRecord", start, end, page, pageSize)
	if err != nil {
		return err
	}
	data, err := getRecords(ctx, url, c.token.JWT)
	if err != nil {
		return err
	}

	type Record struct {
		NickName     string `json:"nickName"`
		RewardType   int    `json:"rewardType"`
		RewardNumber int    `json:"rewardNumber"`
		RewardUnit   int    `json:"rewardUnit"`
		RewardTime   int64  `json:"rewardTime"`
	}
	var resp struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result  int
			Records []Record `json:"record"`
		}
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return err
	}
	log.Printf("Response:\n%s", beautifyJson(resp))

	log.Printf("id	nickName	rewardType	rewardNumber	rewardUnit	rewardTime")
	for i, r := range resp.Data.Records {
		rewardAt := time.Unix(r.RewardTime/1000, 0)
		log.Printf("%d	%s	%d	%d %d	%s", i, r.NickName, r.RewardType, r.RewardNumber, r.RewardUnit, rewardAt)
	}
	return nil
}

func (c *Client) RechargeRecords(ctx context.Context, start, end int64, page, pageSize int) error {
	_, err := c.Login(false)
	if err != nil {
		return err
	}
	url, err := pagingRequest(c.Server, "rechargeRecord", start, end, page, pageSize)
	if err != nil {
		return err
	}
	data, err := getRecords(ctx, url, c.token.JWT)
	if err != nil {
		return err
	}

	type Record struct {
		NickName       string  `json:"rechargeFor"`
		RechargeFrom   string  `json:"rechargeFrom"`
		RechargeTo     string  `json:"rechargeTo"`
		RechargeNumber float64 `json:"rechargeNumber"`
		RechargeUnit   int     `json:"rechargeUnit"`
		RechargeTime   int64   `json:"rechargeTime"`
	}
	var resp struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result  int
			Records []Record `json:"record"`
		}
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return err
	}
	log.Printf("Response:\n%s", beautifyJson(resp))

	log.Printf("id	nickName	rechargeFrom	rechargeTo	rechargeNumber	rechargeUnit	rechargeTime")
	for i, r := range resp.Data.Records {
		rewardAt := time.Unix(r.RechargeTime/1000, 0)
		log.Printf("%d	%s	%s	%s	%f %d	%s", i, r.NickName, r.RechargeFrom, r.RechargeTo, r.RechargeNumber, r.RechargeUnit, rewardAt)
	}
	return nil
}

func pagingRequest(server, relativePath string, start, end int64, page, pageSize int) (string, error) {
	u, err := url.Parse(server)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, relativePath)
	q := u.Query()
	q.Set("tamptime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	if start > 0 {
		q.Set("start", strconv.FormatInt(start, 10))
	}
	if end > 0 {
		q.Set("end", strconv.FormatInt(end, 10))
	}
	if page > 0 {
		q.Set("pager", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("pagerNum", strconv.Itoa(pageSize))
	}
	u.RawQuery = q.Encode()
	log.Printf("request URL %s", u)
	return u.String(), nil
}

func getRecords(ctx context.Context, url, token string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	hc := &http.Client{}
	resp, err := hc.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("Raw response: %s", string(body))
	return body, nil
}
