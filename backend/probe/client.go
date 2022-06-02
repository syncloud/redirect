package probe

import (
	"crypto/tls"
	"net/http"
	"time"
)

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

func NewClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 1,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
