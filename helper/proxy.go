package helper

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewHostReverseProxyHandlerMap(proxyHostMap map[string]string) map[string]http.Handler {
	handlers := make(map[string]http.Handler)
	for domain, target := range proxyHostMap {
		targetURL, err := url.Parse(target)
		if err != nil {
			log.Println(fmt.Sprintf("[目标URL错误] %s: %v", target, err))
			continue
		}
		domain = strings.TrimSpace(domain)
		if !CheckHost(domain) {
			continue
		}
		var proxy = httputil.NewSingleHostReverseProxy(targetURL)
		var tmpDirector = proxy.Director // 自定义Director以正确处理请求头
		proxy.Director = func(req *http.Request) {
			tmpDirector(req)
			req.Host = targetURL.Host                    // 设置正确的Host头
			req.Header.Set("X-Forwarded-Host", req.Host) // 添加X-Forwarded头信息
			req.Header.Set("X-Forwarded-Proto", "http")
		}
		handlers[domain] = proxy
	}
	return handlers
}
