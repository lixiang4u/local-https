package helper

import (
	"fmt"
	"github.com/lixiang4u/local-https/model"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewHostReverseProxyHandlerMap(proxyList []model.Proxy) map[string]http.Handler {
	handlers := make(map[string]http.Handler)
	for _, item := range proxyList {
		targetURL, err := url.Parse(item.Backend)
		if err != nil {
			log.Println(fmt.Sprintf("[目标URL错误] %s->%s: %v", item.Host, item.Backend, err))
			continue
		}
		item.Host = strings.TrimSpace(item.Host)
		if !CheckHost(item.Host) {
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
		proxy.ModifyResponse = func(resp *http.Response) error {
			if item.Cors {
				resp.Header.Set("Access-Control-Allow-Origin", "*")
			}
			resp.Header.Set("X-Client-Server", "local-https")
			return nil
		}
		handlers[item.Host] = proxy
	}
	return handlers
}
