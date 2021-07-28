package client

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func (c *Client) loadToken() (*Token, error) {
	data, err := ioutil.ReadFile(c.cfg.TokenFile)
	if errors.Is(err, os.ErrNotExist) {
		log.Println("token file is not exist")
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var token Token
	if err = json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (c *Client) saveToken(token *Token) error {
	data, _ := json.Marshal(token)
	if err := ioutil.WriteFile(c.cfg.TokenFile, data, 0644); err != nil {
		return err
	}
	return nil
}

func (c *Client) login() (*Token, error) {
	u, err := url.Parse(c.Server)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "login")
	log.Printf("request URL %s", u)

	return login(u.String(), c.Email(), c.cfg.Password)
}

func (c *Client) loginAndSave() (*Token, error) {
	token, err := c.login()
	if err != nil {
		return nil, err
	}

	c.token = token

	if err = c.saveToken(token); err != nil {
		return nil, err
	}
	return token, nil
}

func login(url, email, password string) (*Token, error) {
	request := struct {
		Email     string `json:"mail"`
		Password  string `json:"pwd"`
		Timestamp string `json:"tamptime"`
	}{
		Email:     email,
		Password:  password,
		Timestamp: strconv.FormatInt(time.Now().Unix()*1000, 10),
	}
	resp, err := post(url, request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	log.Printf("Header: %s", beautifyJson(resp.Header))
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ret struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result int
			IV     int `json:"IV"`
		}
	}
	if err = json.Unmarshal(body, &ret); err != nil {
		return nil, err
	}
	log.Printf("Response:\n%s", beautifyJson(ret))
	if ret.Data.Result != 1 {
		return nil, errors.New(ret.Message)
	}
	cookie := resp.Header["Set-Cookie"]
	if cookie == nil || len(cookie) < 1 {
		return nil, errors.New("no cookie")
	}
	token := parseToken(cookie[0])
	log.Printf("Cookie:\n%s", beautifyJson(token))

	return &token, nil
}

func parseToken(cookie string) Token {
	kvs := strings.Split(cookie, "; ")
	var token Token
	for _, kv := range kvs {
		tuple := strings.Split(kv, "=")
		if len(tuple) != 2 {
			continue
		}

		if tuple[0] == "jwt" {
			token.JWT = tuple[1]
		} else if tuple[0] == "Max-Age" {
			age, err := strconv.Atoi(tuple[1])
			if err != nil {
				continue
			}
			token.ExpireAt = time.Now().Add(time.Second * time.Duration(age))
		}
	}
	return token
}
