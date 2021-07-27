package client

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("parse response error: %s", err)
	}
	var timestamp RespTimestamp
	if err = json.Unmarshal(body, &timestamp); err != nil {
		return 0, err
	}
	printJson(timestamp)
	return timestamp.Data.Timestamp, nil
}

func printJson(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	log.Printf("Response:\n%s", string(data))
}
