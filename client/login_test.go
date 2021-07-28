package client

import "testing"

func TestParseToken(t *testing.T) {
	tests := []string{
		"jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjc5OTcwOTUsImlkIjoibWlnb25neCtjYWl0YW4wM0BnbWFpbC5jb20iLCJvcmlnX2lhdCI6MTYyNzM5MjI5NX0.Ah77qgmFZVCiU4p2BcgGiab4Or3QBET2fnh2gvX0RMQ; Path=/; Max-Age=604800",
	}
	for i, tt := range tests {
		token := parseToken(tt)
		t.Logf("[%d] %v", i, token)
	}
}
