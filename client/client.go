package client

type Config struct {
	Email     string
	Password  string
	TokenFile string `json:"tokenFile"`
}
