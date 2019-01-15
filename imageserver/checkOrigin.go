package imageserver

import (
	"fmt"
	"net/http"
	"net/url"
)

type CheckOriginParams struct {
	UserName string
	Request  *http.Request
}

func CheckOrigin(params CheckOriginParams) error {
	allowedDomains := getAllowedDomains(params.UserName, 1)

	if params.Request.Referer() == "" || len(allowedDomains) == 0 {
		return nil
	}
	u, err := url.Parse(params.Request.Referer())
	if err != nil {
		return fmt.Errorf("Failed to parse requeset referer.")
	}
	_, ok := allowedDomains[u.Hostname()]
	if !ok {
		return fmt.Errorf("Domain not allowed.")
	}
	return nil
}

type checkRemoteOriginParams struct {
	UserName string
	UrlStr   string
}

func checkRemoteOrigin(params checkRemoteOriginParams) error {
	allowedDomains := getAllowedDomains(params.UserName, 2)
	if params.UrlStr == "" {
		return fmt.Errorf("Domain not allowed.")
	}
	u, err := url.Parse(params.UrlStr)
	if err != nil {
		return fmt.Errorf("Failed to parse resource url.")
	}
	_, ok := allowedDomains[u.Scheme+"://"+u.Hostname()]
	if !ok {
		return fmt.Errorf("Domain not allowed.")
	}
	return nil
}
