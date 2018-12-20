package imageserver

import (
	"fmt"
	"net/http"
	"net/url"
)

type checkOriginParams struct {
	UserName string
	Request  *http.Request
}

func checkOrigin(params checkOriginParams) error {
	allowedDomains := getAllowedDomains(params.UserName, 0)
	fmt.Println(allowedDomains)
	if params.Request.Referer() == "" {
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
	allowedDomains := getAllowedDomains(params.UserName, 1)
	if params.UrlStr == "" {
		return fmt.Errorf("Domain not allowed.")
	}
	u, err := url.Parse(params.UrlStr)
	if err != nil {
		return fmt.Errorf("Failed to parse resource url.")
	}
	_, ok := allowedDomains[u.Hostname()]
	if !ok {
		return fmt.Errorf("Domain not allowed.")
	}
	return nil
}
