package client

import (
	"bytes"
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

type RespTimestamp struct {
	State   int
	Message string `json:"msg"`
	Data    struct {
		Result    int
		Timestamp int64
	}
}

func Timestamp(server string) (int64, error) {
	u, err := url.Parse(server)
	if err != nil {
		log.Fatalf("bad server url: %s, error: %s", server, err)
	}
	u.Path = path.Join(u.Path, "timestamp")
	log.Printf("request URL %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatalf("request server %s error: %s", server, err)
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("parse response error: %s", err)
	}
	var timestamp RespTimestamp
	if err = json.Unmarshal(body, &timestamp); err != nil {
		return 0, err
	}
	log.Printf("Response:\n%s", BeautifyJson(timestamp))
	return timestamp.Data.Timestamp, nil
}

func SendCode(server, email string) error {
	u, err := url.Parse(server)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "sendCode")
	log.Printf("request URL %s", u)
	request := struct {
		Email     string `json:"mail"`
		Timestamp string `json:"tamptime"`
	}{
		Email:     email,
		Timestamp: strconv.FormatInt(time.Now().Unix()*1000, 10),
	}
	resp, err := post(u.String(), request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result int
		}
	}
	if err = json.Unmarshal(body, &ret); err != nil {
		return err
	}
	log.Printf("Response:\n%s", BeautifyJson(ret))
	return nil
}

func Register(server, email, password, verify, invite string) error {
	u, err := url.Parse(server)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "register")
	log.Printf("request URL %s", u)
	request := struct {
		Email      string `json:"mail"`
		Password   string `json:"pwd"`
		VerifyCode string `json:"code,omitempty"`
		InviteCode string `json:"invitationCode,omitempty"`
		Timestamp  string `json:"tamptime"`
	}{
		Email:      email,
		Password:   password,
		VerifyCode: verify,
		InviteCode: invite,
		Timestamp:  strconv.FormatInt(time.Now().Unix()*1000, 10),
	}
	resp, err := post(u.String(), request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Printf("Status: %s", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret struct {
		State   int
		Message string `json:"msg"`
		Data    struct {
			Result int
		}
	}
	if err = json.Unmarshal(body, &ret); err != nil {
		return err
	}
	log.Printf("Response:\n%s", BeautifyJson(ret))
	return nil
}

func BeautifyJson(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return fmt.Sprintf("%s", string(data))
}

func post(url string, request interface{}) (*http.Response, error) {
	log.Printf("Request:\n%s", BeautifyJson(request))
	b, _ := json.Marshal(request)
	r := bytes.NewBuffer(b)
	return http.Post(url, "application/json", r)
}
