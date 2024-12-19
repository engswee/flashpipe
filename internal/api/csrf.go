package api

import (
	"github.com/engswee/flashpipe/internal/httpclnt"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Csrf struct {
	exe         *httpclnt.HTTPExecuter
	token       string
	csrfCookies []*http.Cookie
}

// NewCsrf returns an initialised Csrf instance.
func NewCsrf(exe *httpclnt.HTTPExecuter) *Csrf {
	c := new(Csrf)
	c.exe = exe
	return c
}

func (c *Csrf) GetToken() (string, []*http.Cookie, error) {
	if c.token == "" {
		log.Debug().Msg("Get CSRF Token")
		headers := map[string]string{
			"x-csrf-token": "fetch",
		}
		resp, err := c.exe.ExecGetRequest("/api/v1/", headers)

		if err != nil {
			return "", nil, err
		}
		if resp.StatusCode == 200 {
			c.token = resp.Header.Get("x-csrf-token")
			c.csrfCookies = resp.Cookies()
			log.Debug().Msgf("Received CSRF Token - %v", c.token)
		} else {
			_, err = c.exe.LogError(resp, "Get CSRF Token")
			return "", nil, err
		}
	}
	return c.token, c.csrfCookies, nil
}

func InitHeadersAndCookies(exe *httpclnt.HTTPExecuter) (headers map[string]string, cookies []*http.Cookie, err error) {
	headers = map[string]string{}
	cookies = []*http.Cookie{}

	if exe.AuthType == "BASIC" {
		csrf := NewCsrf(exe)
		var token string
		token, cookies, err = csrf.GetToken()
		headers["x-csrf-token"] = token
	}
	return
}
