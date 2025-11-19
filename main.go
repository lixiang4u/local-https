package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/local-https/helper"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if !helper.WindowsTokenElevated() {
		helper.ExitMsg("请使用管理员权限打开")
	}

	var proxyHostMap = map[string]string{
		"fuck-www.fuck-host.com": "",
		"fuck-abc.fuck-host.com": "",
		"fuck-xyz.fuck-host.com": "",
	}

	var domain = "fuck-host.org"
	cert, key, err := helper.MakeDomainCertificate(domain, helper.MapKeys(proxyHostMap))
	if err != nil {
		log.Println(fmt.Sprintf("证书生成失败：%s", err.Error()))
		return
	}
	_, err = helper.AddCertToRoot(cert)
	if err != nil {
		log.Println(fmt.Sprintf("导入证书失败：%s", err.Error()))
		return
	}
	//log.Println("[导入证书信息]", string(output))

	var p = 8060
	for tmpHost, _ := range proxyHostMap {
		tmpHost = strings.TrimSpace(tmpHost)
		if !helper.CheckHost(tmpHost) {
			continue
		}
		// 启动虚拟web服务
		p = helper.NextUsefulPort(p)
		go runLocalHttpServer(p, tmpHost)
		time.Sleep(time.Second / 3)
		_ = helper.UpdateWindowsHosts(fmt.Sprintf("127.0.0.1	%s", tmpHost))

		// 更新代理地址
		proxyHostMap[tmpHost] = fmt.Sprintf("http://127.0.0.1:%d", p)
	}

	runReverseProxyServer(proxyHostMap, cert, key)

	helper.ExitWithCtrlC()
}

func runLocalHttpServer(port int, domain string) {
	_ = helper.MkdirAll(filepath.Join(helper.AppPath(), "www", domain, "1.txt"))
	gin.SetMode(gin.ReleaseMode)
	var app = gin.Default()
	app.Use(gin.Recovery())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "HEAD"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Hit-Cache"},
		AllowCredentials: true,
	}))
	app.StaticFS("/", gin.Dir(filepath.Join(helper.AppPath(), "www", domain), true))
	app.NoRoute(func(ctx *gin.Context) {
		ctx.String(http.StatusOK, fmt.Sprintf("[http://127.0.0.1:%d] %s", port, time.Now().String()))
	})
	log.Println(fmt.Sprintf("[反向代理] %s -> 127.0.0.1:%d", domain, port))
	if err := app.Run(fmt.Sprintf(fmt.Sprintf(":%d", port))); err != nil {
		log.Println(fmt.Sprintf("[http://127.0.0.1:%d] 启动失败 %s", port, err.Error()))
	}
}

func runWebServer(domain, cert, key string) {
	var app = gin.Default()
	app.Use(gin.Recovery())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "HEAD"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Hit-Cache"},
		AllowCredentials: true,
	}))
	app.NoRoute(func(ctx *gin.Context) {
		ctx.String(http.StatusOK, fmt.Sprintf("[%s] %s", domain, time.Now().String()))
	})

	if err := app.RunTLS(fmt.Sprintf(":443"), cert, key); err != nil {
		log.Println(fmt.Sprintf("站点(%s)启动失败：%s", domain, err.Error()))
	}
}

func runReverseProxyServer(proxyHostMap map[string]string, cert, key string) {
	var proxyHandlers = helper.NewHostReverseProxyHandlerMap(proxyHostMap)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 查找对应的代理处理器
		if handler, ok := proxyHandlers[r.Host]; ok {
			handler.ServeHTTP(w, r)
			return
		}
		// 如果没有找到对应的域名映射
		http.Error(w, fmt.Sprintf("[%s] not found", r.Host), http.StatusNotFound)
	})
	go func() {
		log.Println("[启动本地http服务]")
		log.Fatal(http.ListenAndServe(":80", nil))
	}()
	go func() {
		log.Println("[启动本地https服务]")
		log.Fatal(http.ListenAndServeTLS(":443", cert, key, nil))
	}()
}
