package helper

import (
	"net/url"
	"strings"
)

func ParseHost(tmpUrl string) (host string) {
	tmpUrl2, err := url.Parse(tmpUrl)
	if err != nil {
		return
	}
	if tmpUrl2.Host == "" {
		return
	}
	return tmpUrl2.Host
}

func CheckHost(domain string) bool {
	domain = strings.TrimSpace(domain)
	if len(domain) <= 0 {
		return false
	}
	if strings.Contains(domain, "*") {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	return true
}
