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
	JWT      string    `json:"jwt"`
	ExpireAt time.Time `json:"expireAt"`
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

func (c *Client) UserInfo(ctx context.Context) (*Profile, error) {
	if _, err := c.Login(false); err != nil {
		return nil, err
	}

	u, err := url.Parse(c.Server)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "user")
	q := u.Query()
	q.Set("tamptime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()
	log.Printf("request URL %s", u)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token.JWT))
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
		return nil, err
	}
	log.Printf("Response:\n%s", BeautifyJson(ret))

	profile := Profile{
		Email:         ret.Data.Email,
		Code:          ret.Data.Code,
		Expire:        ret.Data.Expire,
		RemainingTime: time.Duration(ret.Data.RemainTime * 1e6).String(),
	}
	log.Printf("Profile:\n%s", BeautifyJson(profile))
	return &profile, nil
}

type InvitationRecord struct {
	NickName     string `json:"nickName"`
	RewardType   int    `json:"rewardType"`
	RewardNumber int    `json:"rewardNumber"`
	RewardUnit   int    `json:"rewardUnit"`
	RewardTime   int64  `json:"rewardTime"`
}

func (c *Client) InvitationRecords(ctx context.Context, start, end int64, page, pageSize int) ([]InvitationRecord, error) {
	if _, err := c.Login(false); err != nil {
		return nil, err
	}

	url, err := pagingRequest(c.Server, "invitationRecord", start, end, page, pageSize)
	if err != nil {
		return nil, err
	}
	data, err := getRecords(ctx, url, c.token.JWT)
	if err != nil {
		return nil, err
	}

	var resp struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result  int
			Records []InvitationRecord `json:"record"`
		}
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	//log.Printf("Response:\n%s", BeautifyJson(resp))

	return resp.Data.Records, nil
}

type RechargeRecord struct {
	RechargeFor    int     `json:"rechargeFor"`
	RechargeFrom   string  `json:"rechargeFrom"`
	RechargeTo     string  `json:"rechargeTo"`
	RechargeNumber float64 `json:"rechargeNumber"`
	RechargeUnit   int     `json:"rechargeUnit"`
	RechargeTime   int64   `json:"rechargeTime"`
	Chain          string  `json:"chain"`
	Amount         float64 `json:"amount"`
	Symbol         string  `json:"symbol"`
	ArrivalTime    int64   `json:"arrivalTime"` // record created time in ms
}

func (c *Client) RechargeRecords(ctx context.Context, start, end int64, page, pageSize int) ([]RechargeRecord, error) {
	_, err := c.Login(false)
	if err != nil {
		return nil, err
	}
	url, err := pagingRequest(c.Server, "rechargeRecord", start, end, page, pageSize)
	if err != nil {
		return nil, err
	}
	data, err := getRecords(ctx, url, c.token.JWT)
	if err != nil {
		return nil, err
	}

	var resp struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result  int
			Records []RechargeRecord `json:"record"`
		}
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	//log.Printf("Response:\n%s", BeautifyJson(resp))

	return resp.Data.Records, nil
}

func (c *Client) Bind(ctx context.Context, code string) (bool, error) {
	_, err := c.Login(false)
	if err != nil {
		return false, err
	}

	u, err := url.Parse(c.Server)
	if err != nil {
		return false, err
	}
	u.Path = path.Join(u.Path, "bindInvitation")
	q := u.Query()
	q.Set("invitationCode", code)
	q.Set("tamptime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token.JWT))
	hc := &http.Client{}
	resp, err := hc.Do(request)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	log.Printf("Raw response: %s", string(body))
	var ret struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result int
		}
	}
	if err := json.Unmarshal(body, &ret); err != nil {
		return false, err
	}
	if ret.Data.Result == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func (c *Client) Address(ctx context.Context, protocol, cType int, force bool) (string, error) {
	_, err := c.Login(false)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(c.Server)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "recharge")
	q := u.Query()
	q.Set("protocol", strconv.Itoa(protocol))
	q.Set("protocol", strconv.Itoa(cType))
	q.Set("force", fmt.Sprintf("%v", force))
	q.Set("tamptime", strconv.FormatInt(time.Now().Unix()*1000, 10))
	u.RawQuery = q.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token.JWT))
	hc := &http.Client{}
	resp, err := hc.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("Raw response: %s", string(body))
	var ret struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result   int
			Protocol int
			Type     int
			Address  string `json:"addressText"`
			Remarks  string `json:"remarks"`
		}
	}
	if err := json.Unmarshal(body, &ret); err != nil {
		return "", err
	}
	if ret.State == 200 {
		return ret.Data.Address, nil
	} else {
		return "", nil
	}
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
